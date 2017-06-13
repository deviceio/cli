package sdk

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/deviceio/hmapi"
)

type DeviceProcess interface {
	Create(ctx context.Context, cmd string, args []string) (DeviceProcessInstance, error)
}

type DeviceProcessInstance interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Delete(ctx context.Context) error
	Stdin(ctx context.Context) io.WriteCloser
	Stdout(ctx context.Context) io.Reader
	Stderr(ctx context.Context) io.Reader
}

type deviceProcess struct {
	device       *device
	resourcePath string
}

func (t *deviceProcess) Create(ctx context.Context, cmd string, args []string) (DeviceProcessInstance, error) {
	form := t.device.client.hmclient.
		Resource(t.resourcePath).
		Form("create").
		AddFieldAsString("cmd", cmd)

	for _, arg := range args {
		form.AddFieldAsString("arg", arg)
	}

	resp, err := form.Submit(ctx)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New("failed process creation: " + string(body))
	}

	return &deviceProcessInstance{
		resourcePath: resp.Header.Get("Location"),
		device:       t.device,
	}, nil
}

type deviceProcessInstance struct {
	device       *device
	resourcePath string
}

func (t *deviceProcessInstance) Start(ctx context.Context) error {
	resp, err := t.device.client.hmclient.
		Resource(t.resourcePath).
		Form("start").
		Submit(ctx)

	if err != nil {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return &ErrInvalidAPIResponse{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	return nil
}

func (t *deviceProcessInstance) Stop(ctx context.Context) error {
	resp, err := t.device.client.hmclient.
		Resource(t.resourcePath).
		Form("stop").
		Submit(ctx)

	if err != nil {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return &ErrInvalidAPIResponse{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	return nil
}

func (t *deviceProcessInstance) Delete(ctx context.Context) error {
	resp, err := t.device.client.hmclient.
		Resource(t.resourcePath).
		Form("delete").
		Submit(ctx)

	if err != nil {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return &ErrInvalidAPIResponse{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	return nil
}

func (t *deviceProcessInstance) Stdin(ctx context.Context) io.WriteCloser {
	datar, dataw := io.Pipe()

	writer := &deviceProcessStdinWriter{
		ctx:   ctx,
		datar: datar,
		dataw: dataw,
	}

	go func() {
		resp, err := t.device.client.hmclient.
			Resource(t.resourcePath).
			Form("stdin").
			AddFieldAsOctetStream("data", datar).
			Submit(ctx)

		if err != nil {
			writer.reqerr <- err
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)
			writer.reqerr <- errors.New(string(body))
		}
	}()

	return writer
}

func (t *deviceProcessInstance) Stdout(ctx context.Context) io.Reader {
	resp, err := t.device.client.hmclient.
		Resource(t.resourcePath).
		Link("stdout").
		Get(ctx)

	reader := &deviceProcessOutputReader{
		ctx:     ctx,
		resp:    resp,
		resperr: err,
	}

	if resp != nil && resp.StatusCode >= 300 {
		data, _ := ioutil.ReadAll(resp.Body)
		reader.resperrbody = string(data)
	}

	return reader
}

func (t *deviceProcessInstance) Stderr(ctx context.Context) io.Reader {
	resp, err := t.device.client.hmclient.
		Resource(t.resourcePath).
		Link("stderr").
		Get(ctx)

	reader := &deviceProcessOutputReader{
		ctx:     ctx,
		resp:    resp,
		resperr: err,
	}

	if resp != nil && resp.StatusCode >= 300 {
		data, _ := ioutil.ReadAll(resp.Body)
		reader.resperrbody = string(data)
	}

	return reader
}

type deviceProcessStdinWriter struct {
	reqerr chan error
	ctx    context.Context
	datar  *io.PipeReader
	dataw  *io.PipeWriter
}

func (t *deviceProcessStdinWriter) Write(p []byte) (n int, err error) {
	select {
	case <-t.ctx.Done():
		return 0, t.ctx.Err()
	case err := <-t.reqerr:
		return 0, err
	default:
		return t.dataw.Write(p)
	}
}

func (t *deviceProcessStdinWriter) Close() error {
	return t.datar.Close()
}

type deviceProcessOutputReader struct {
	ctx         context.Context
	resp        *hmapi.LinkResponse
	resperr     error
	resperrbody string
}

func (t *deviceProcessOutputReader) Read(p []byte) (n int, err error) {
	select {
	case <-t.ctx.Done():
		return 0, t.ctx.Err()
	default:
	}

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
