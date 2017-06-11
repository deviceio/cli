package hmapi

import (
	"errors"
	"io"
	"net/http"
	"strings"
)

type Link struct {
	Href     string    `json:"href,omitempty"`
	Type     MediaType `json:"type,omitempty"`
	Encoding MediaType `json:"encoding,omitempty"`
}

type LinkRequest interface {
	Get() (LinkResponse, error)
}

type LinkResponse interface {
	HttpResponse() *http.Response
	AsOctetStream() (io.Reader, error)
}

type linkRequest struct {
	name     string
	resource *resourceRequest
}

func (t *linkRequest) Get() (LinkResponse, error) {
	res, err := t.resource.Get()

	if err != nil {
		return nil, err
	}

	hmlink, ok := res.Links[t.name]

	if !ok {
		return nil, &LinkNotFound{
			LinkName: t.name,
			Resource: t.resource.path,
		}
	}

	request, err := http.NewRequest(
		string(GET),
		t.resource.client.baseuri+hmlink.Href,
		nil,
	)

	if err != nil {
		return nil, err
	}

	resp, err := t.resource.client.do(request)

	if err != nil {
		return nil, err
	}

	return &linkResponse{
		httpResponse: resp,
	}, nil
}

type linkResponse struct {
	httpResponse *http.Response
}

func (t *linkResponse) HttpResponse() *http.Response {
	return t.httpResponse
}

func (t *linkResponse) AsOctetStream() (io.Reader, error) {
	ct := t.httpResponse.Header.Get("Content-Type")

	if !strings.Contains(strings.ToLower(ct), strings.ToLower(MediaTypeOctetStream.String())) {
		return nil, errors.New("response has invalid content type to become octet stream")
	}

	return t.httpResponse.Body, nil
}
