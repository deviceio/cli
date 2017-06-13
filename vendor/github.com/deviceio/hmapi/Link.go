package hmapi

import (
	"context"
	"net/http"
)

type Link struct {
	Href     string    `json:"href,omitempty"`
	Type     MediaType `json:"type,omitempty"`
	Encoding MediaType `json:"encoding,omitempty"`
}

type LinkRequest interface {
	Get(context.Context) (*LinkResponse, error)
}

type LinkResponse struct {
	*http.Response
}

type linkRequest struct {
	name     string
	resource *resourceRequest
}

func (t *linkRequest) Get(ctx context.Context) (*LinkResponse, error) {
	res, err := t.resource.Get(ctx)

	if err != nil {
		return nil, err
	}

	hmlink, ok := res.Links[t.name]

	if !ok {
		return nil, &ErrResourceNoSuchLink{
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

	return &LinkResponse{
		resp,
	}, nil
}
