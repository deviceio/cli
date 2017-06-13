package hmapi

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
)

type FormRequest interface {
	AddField(name string, media MediaType, value interface{}) FormRequest
	AddFieldAsString(name string, value string) FormRequest
	AddFieldAsBool(name string, value bool) FormRequest
	AddFieldAsOctetStream(name string, value io.Reader) FormRequest
	AddFieldAsInt(name string, value int) FormRequest
	Submit(ctx context.Context) (*FormResponse, error)
}

type FormResponse struct {
	*http.Response
}

type FormSubmission interface {
	Response() *http.Response
	Err() error
	Done() <-chan struct{}
	Cancel()
}

type Form struct {
	Action  string       `json:"action,omitempty"`
	Method  method       `json:"method"`
	Type    MediaType    `json:"type,omitempty"`
	Enctype MediaType    `json:"enctype,omitempty"`
	Fields  []*FormField `json:"fields,omitempty"`
}

type FormField struct {
	Name     string      `json:"name"`
	Type     MediaType   `json:"type,omitempty"`
	Encoding MediaType   `json:"encoding,omitempty"`
	Required bool        `json:"required"`
	Multiple bool        `json:"multiple"`
	Value    interface{} `json:"value,omitempty"`
}

type formRequest struct {
	name     string
	fields   []*formField
	resource *resourceRequest
}

type formField struct {
	name      string
	mediaType MediaType
	value     interface{}
}

func (t *formRequest) AddField(name string, media MediaType, value interface{}) FormRequest {
	t.fields = append(t.fields, &formField{
		name:      name,
		mediaType: media,
		value:     value,
	})

	return t
}

func (t *formRequest) AddFieldAsString(name string, value string) FormRequest {
	t.AddField(name, MediaTypeHMAPIString, value)
	return t
}

func (t *formRequest) AddFieldAsBool(name string, value bool) FormRequest {
	t.AddField(name, MediaTypeHMAPIBoolean, value)
	return t
}

func (t *formRequest) AddFieldAsOctetStream(name string, value io.Reader) FormRequest {
	t.AddField(name, MediaTypeOctetStream, value)
	return t
}

func (t *formRequest) AddFieldAsInt(name string, value int) FormRequest {
	t.AddField(name, MediaTypeHMAPIInt, value)
	return t
}

func (t *formRequest) Submit(ctx context.Context) (retresp *FormResponse, reterr error) {
	hmres, err := t.resource.Get(ctx)

	if err != nil {
		return nil, err
	}

	hmform, ok := hmres.Forms[t.name]

	if !ok {
		return nil, &ErrResourceNoSuchForm{
			FormName: t.name,
			Resource: t.resource.path,
		}
	}

	bodyr, bodyw := io.Pipe()

	request, err := http.NewRequest(
		hmform.Method.String(),
		t.resource.client.baseuri+hmform.Action,
		bodyr,
	)

	if err != nil {
		return nil, err
	}

	request = request.WithContext(ctx)

	switch hmform.Enctype {
	case MediaTypeMultipartFormData:
		request.Header.Set("Content-Type", MediaTypeMultipartFormData.String())
	default:
		return nil, &ErrUnsupportedMediaType{
			MediaType: hmform.Enctype,
		}
	}

	chresp := make(chan *http.Response)
	chresperr := make(chan error)
	chformerr := make(chan error)

	go func() {
		resp, err := t.resource.client.do(request)
		chresperr <- err
		chresp <- resp
	}()

	go func() {
		switch hmform.Enctype {
		case MediaTypeMultipartFormData:
			t.writeMultipartForm(bodyw, hmform, chformerr)
			bodyw.Close()
		}
	}()

waitforcomplete:
	for {
		select {
		case formerr := <-chformerr:
			reterr = formerr

		case resperr := <-chresperr:
			reterr = resperr

		case resp := <-chresp:
			retresp = &FormResponse{resp}
			break waitforcomplete

		case <-ctx.Done():
			reterr = ctx.Err()
			break waitforcomplete
		}
	}

	return retresp, reterr
}

func (t *formRequest) writeMultipartForm(writer io.Writer, form *Form, cherr chan error) {
	mpwriter := multipart.NewWriter(writer)
	mpwriter.SetBoundary(MultipartFormDataBoundry)
	defer mpwriter.Close()

	for _, field := range t.fields {
		switch field.mediaType {
		case MediaTypeOctetStream:
			fieldreader, ok := field.value.(io.Reader)

			if !ok {
				cherr <- errors.New("octetstream field is not a io.Reader")
				return
			}

			fieldwriter, err := mpwriter.CreateFormField(field.name)

			if err != nil {
				cherr <- err
				return
			}

			if _, err = io.Copy(fieldwriter, fieldreader); err != nil {
				cherr <- err
				return
			}

		case MediaTypeHMAPIInt:
			if err := mpwriter.WriteField(field.name, strconv.FormatInt(int64(field.value.(int)), 10)); err != nil {
				cherr <- err
				return
			}

		case MediaTypeHMAPIString:
			if err := mpwriter.WriteField(field.name, field.value.(string)); err != nil {
				cherr <- err
				return
			}

		case MediaTypeHMAPIBoolean:
			if err := mpwriter.WriteField(field.name, strconv.FormatBool(field.value.(bool))); err != nil {
				cherr <- err
				return
			}

		default:
			cherr <- &ErrUnsupportedMediaType{
				MediaType: form.Enctype,
			}
			return
		}
	}

	cherr <- nil
}

type formResponse struct {
	httpResponse *http.Response
}

func (t *formResponse) HttpResponse() *http.Response {
	return t.httpResponse
}
