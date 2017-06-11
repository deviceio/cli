package hmapi

import (
	"encoding/json"
	"net/http"
)

type ResourceRequest interface {
	Get() (*Resource, error)
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
	return nil
}

func (t *resourceRequest) Get() (*Resource, error) {
	request, err := http.NewRequest(GET.String(), t.client.baseuri+t.path, nil)

	if err != nil {
		return nil, err
	}

	resp, err := t.client.do(request)

	if err != nil {
		return nil, err
	}

	var jsonResource *Resource

	if err = json.NewDecoder(resp.Body).Decode(&jsonResource); err != nil {
		return nil, err
	}

	return jsonResource, nil
}
