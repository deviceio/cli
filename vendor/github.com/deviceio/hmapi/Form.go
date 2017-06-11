package hmapi

import (
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
	Submit() (FormResponse, error)
}

type FormResponse interface {
	HttpResponse() *http.Response
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

func (t *formRequest) Submit() (FormResponse, error) {
	hmres, err := t.resource.Get()

	if err != nil {
		return nil, err
	}

	hmform, ok := hmres.Forms[t.name]

	if !ok {
		return nil, &FormNotFound{
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

	switch hmform.Enctype {
	case MediaTypeMultipartFormData:
		request.Header.Set("Content-Type", MediaTypeMultipartFormData.String())
	default:
		return nil, &UnsupportedMediaType{
			MediaType: hmform.Enctype,
		}
	}

	chresp, cherr := t.preflightHTTPRequest(request)

	switch hmform.Enctype {
	case MediaTypeMultipartFormData:
		mpwriter := multipart.NewWriter(bodyw)
		mpwriter.SetBoundary(MultipartFormDataBoundry)

		for _, field := range t.fields {
			switch field.mediaType {
			case MediaTypeOctetStream:
				fieldwriter, _ := mpwriter.CreateFormField(field.name)
				io.Copy(fieldwriter, field.value.(io.Reader))

			case MediaTypeHMAPIString:
				mpwriter.WriteField(field.name, field.value.(string))

			case MediaTypeHMAPIBoolean:
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

	result := &formResponse{
		httpResponse: resp,
	}

	return result, nil
}

func (t *formRequest) preflightHTTPRequest(r *http.Request) (chan *http.Response, chan error) {
	chresp := make(chan *http.Response)
	cherr := make(chan error)

	go func(chresp chan *http.Response, cherr chan error, req *http.Request) {
		resp, err := t.resource.client.do(r)
		chresp <- resp
		cherr <- err
	}(chresp, cherr, r)

	return chresp, cherr
}

type formResponse struct {
	httpResponse *http.Response
}

func (t *formResponse) HttpResponse() *http.Response {
	return t.httpResponse
}
