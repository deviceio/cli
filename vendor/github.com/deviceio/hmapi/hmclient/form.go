package hmclient

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/deviceio/hmapi"
)

type Form interface {
	SetField(name string, media hmapi.MediaType, value interface{}) Form
	SetFieldAsString(name string, value string) Form
	SetFieldAsBool(name string, value bool) Form
	SetFieldAsOctetStream(name string, value io.Reader) Form
	Submit() (FormResult, error)
}

type FormResult interface {
	RawResponse() *http.Response
}

type form struct {
	name     string
	fields   []*formField
	resource *resource
}

func (t *form) SetField(name string, media hmapi.MediaType, value interface{}) Form {
	t.fields = append(t.fields, &formField{
		name:      name,
		mediaType: media,
		value:     value,
	})

	return t
}

func (t *form) SetFieldAsString(name string, value string) Form {
	t.SetField(name, hmapi.MediaTypeHMAPIString, value)
	return t
}

func (t *form) SetFieldAsBool(name string, value bool) Form {
	t.SetField(name, hmapi.MediaTypeHMAPIBoolean, value)
	return t
}

func (t *form) SetFieldAsOctetStream(name string, value io.Reader) Form {
	t.SetField(name, hmapi.MediaTypeOctetStream, value)
	return t
}

func (t *form) Submit() (FormResult, error) {
	hmres, err := t.resource.Get()

	if err != nil {
		return nil, err
	}

	hmform, ok := hmres.Forms[t.name]

	if !ok {
		return nil, fmt.Errorf("no such form with name '%v' found on resource", t.name)
	}

	bodyr, bodyw := io.Pipe()

	request, err := http.NewRequest(
		hmform.Method,
		t.resource.client.baseuri+hmform.Action,
		bodyr,
	)

	if err != nil {
		return nil, err
	}

	switch hmform.Enctype {
	case hmapi.MediaTypeMultipartFormData:
		request.Header.Set("Content-Type", hmapi.MediaTypeMultipartFormData.String())
	default:
		return nil, &UnsupportedMediaType{
			MediaType: hmform.Enctype,
		}
	}

	// NOTE: this preflights the http request while we are still writing to the
	// request body.
	chresp := make(chan *http.Response)
	cherr := make(chan error)
	go func(chresp chan *http.Response, cherr chan error) {
		resp, err := t.resource.client.do(request)
		chresp <- resp
		cherr <- err
	}(chresp, cherr)

	switch hmform.Enctype {
	case hmapi.MediaTypeMultipartFormData:
		mpwriter := multipart.NewWriter(bodyw)
		mpwriter.SetBoundary(hmapi.MultipartFormDataBoundry)

		for _, field := range t.fields {
			switch field.mediaType {
			case hmapi.MediaTypeOctetStream:
				fieldwriter, _ := mpwriter.CreateFormField(field.name)
				io.Copy(fieldwriter, field.value.(io.Reader))

			case hmapi.MediaTypeHMAPIString:
				mpwriter.WriteField(field.name, field.value.(string))

			case hmapi.MediaTypeHMAPIBoolean:
				mpwriter.WriteField(field.name, strconv.FormatBool(field.value.(bool)))

			default:
				return nil, &UnsupportedMediaType{
					MediaType: hmform.Enctype,
				}
			}
		}

		mpwriter.Close()
		bodyw.Close()
	}

	resp, err := <-chresp, <-cherr

	if err != nil {
		return nil, err
	}

	result := &formResult{
		httpResponse: resp,
	}

	return result, nil
}

type formField struct {
	name      string
	mediaType hmapi.MediaType
	value     interface{}
}

type formResult struct {
	httpResponse *http.Response
}

func (t *formResult) RawResponse() *http.Response {
	return t.httpResponse
}
