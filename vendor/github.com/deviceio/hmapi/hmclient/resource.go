package hmclient

import (
	"encoding/json"

	"net/http"

	"github.com/deviceio/hmapi"
)

type Resource interface {
	Get() (*hmapi.Resource, error)
	Form(name string) Form
	Link(name string) Link
	Content(name string) Content
}

type resource struct {
	path   string
	client *client
}

func (t *resource) Form(name string) Form {
	return &form{
		fields:   []*formField{},
		name:     name,
		resource: t,
	}
}

func (t *resource) Link(name string) Link {
	return nil
}

func (t *resource) Content(name string) Content {
	return nil
}

func (t *resource) Get() (*hmapi.Resource, error) {
	request, err := http.NewRequest("GET", t.client.baseuri+t.path, nil)

	if err != nil {
		return nil, err
	}

	resp, err := t.client.do(request)

	if err != nil {
		return nil, err
	}

	var jsonResource *hmapi.Resource

	if err = json.NewDecoder(resp.Body).Decode(&jsonResource); err != nil {
		return nil, err
	}

	return jsonResource, nil
}
