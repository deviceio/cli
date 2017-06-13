package hmapi

import (
	"context"
	"encoding/json"
	"net/http"
)

type ResourceRequest interface {
	Get(context.Context) (*Resource, error)
	Form(name string) FormRequest
	Link(name string) LinkRequest
	Content(name string) ContentRequest
}

type Resource struct {
	Links   map[string]*Link    `json:"links,omitempty"`
	Forms   map[string]*Form    `json:"forms,omitempty"`
	Content map[string]*Content `json:"content,omitempty"`
}

type resourceRequest struct {
	path   string
	client *client
}

func (t *resourceRequest) Form(name string) FormRequest {
	return &formRequest{
		fields:   []*formField{},
		name:     name,
		resource: t,
	}
}

func (t *resourceRequest) Link(name string) LinkRequest {
	return &linkRequest{
		name:     name,
		resource: t,
	}
}

func (t *resourceRequest) Content(name string) ContentRequest {
	return &contentRequest{}
}

func (t *resourceRequest) Get(ctx context.Context) (*Resource, error) {
	request, err := http.NewRequest(GET.String(), t.client.baseuri+t.path, nil)

	if err != nil {
		return nil, err
	}

	request = request.WithContext(ctx)

	resp, err := t.client.do(request)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &ErrUnexpectedHTTPResponseStatus{
			ExpectedStatus: http.StatusOK,
			ActualStatus:   resp.StatusCode,
			ClientRequest:  request,
			ClientResponse: resp,
		}
	}

	var resource *Resource

	if err = json.NewDecoder(resp.Body).Decode(&resource); err != nil {
		return nil, &ErrResourceUnmarshalFailure{
			UnmarshalError: err,
			ClientRequest:  request,
			ClientResponse: resp,
		}
	}

	if resource.Content == nil {
		resource.Content = map[string]*Content{}
	}

	if resource.Forms == nil {
		resource.Forms = map[string]*Form{}
	}

	if resource.Links == nil {
		resource.Links = map[string]*Link{}
	}

	return resource, nil
}
