package sdk

import (
	"context"
	"errors"
	"io"
	"io/ioutil"

	"github.com/deviceio/hmapi"
)

type DeviceFilesystem interface {
	Reader(ctx context.Context, path string, offset, count int) io.Reader
	Writer(ctx context.Context, path string, append bool) io.WriteCloser
}

type deviceFilesystemReader struct {
	resp        *hmapi.FormResponse
	resperr     error
	resperrbody string
}

func (t *deviceFilesystemReader) Read(p []byte) (n int, err error) {
	if t.resperr != nil {
		return 0, err
	}

	if t.resp != nil && t.resp.StatusCode >= 300 {
		return 0, &ErrInvalidAPIResponse{
			StatusCode: t.resp.StatusCode,
			Message:    t.resperrbody,
		}
	}

	trailerError := t.resp.Trailer.Get("Error")

	if trailerError != "" {
		return 0, errors.New(trailerError)
	}

	n, err = t.resp.Body.Read(p)

	return n, err
}

type deviceFilesystemWriter struct {
	reqerr chan error
	dataw  *io.PipeWriter
	datar  *io.PipeReader
	ctx    context.Context
}

func (t *deviceFilesystemWriter) Write(p []byte) (n int, err error) {
	select {
	case err := <-t.reqerr:
		return 0, err
	case <-t.ctx.Done():
		return 0, t.ctx.Err()
	default:
		return t.dataw.Write(p)
	}
}

func (t *deviceFilesystemWriter) Close() error {
	return t.datar.Close()
}

type deviceFilesystem struct {
	device       *device
	resourcePath string
}

func (t *deviceFilesystem) Reader(ctx context.Context, path string, offset, count int) io.Reader {
	resp, err := t.device.client.hmclient.
		Resource(t.resourcePath).
		Form("read").
		AddFieldAsString("path", path).
		AddFieldAsInt("offset", offset).
		AddFieldAsInt("count", count).
		Submit(ctx)

	fsReader := &deviceFilesystemReader{
		resp:    resp,
		resperr: err,
	}

	if resp != nil && resp.StatusCode >= 300 {
		data, _ := ioutil.ReadAll(resp.Body)
		fsReader.resperrbody = string(data)
	}

	return fsReader
}

func (t *deviceFilesystem) Writer(ctx context.Context, path string, append bool) io.WriteCloser {
	datar, dataw := io.Pipe()

	reqerr := make(chan error)

	go func() {
		resp, err := t.device.client.hmclient.
			Resource(t.resourcePath).
			Form("write").
			AddFieldAsString("path", path).
			AddFieldAsBool("append", append).
			AddFieldAsOctetStream("data", datar).
			Submit(ctx)

		if err != nil {
			reqerr <- err
		}

		if resp.StatusCode >= 300 {
			body, _ := ioutil.ReadAll(resp.Body)
			reqerr <- &ErrInvalidAPIResponse{
				StatusCode: resp.StatusCode,
				Message:    string(body),
			}
		}
	}()

	writer := &deviceFilesystemWriter{
		reqerr: make(chan error),
		dataw:  dataw,
		datar:  datar,
		ctx:    ctx,
	}

	return writer
}
