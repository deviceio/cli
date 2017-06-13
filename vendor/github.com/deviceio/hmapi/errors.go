package hmapi

import (
	"fmt"
	"net/http"
)

type ErrResourceNoSuchLink struct {
	Resource string
	LinkName string
}

func (t *ErrResourceNoSuchLink) Error() string {
	return fmt.Sprintf("no such link with name '%v' defined on resource '%v'", t.LinkName, t.Resource)
}

type ErrResourceNoSuchForm struct {
	Resource string
	FormName string
}

func (t *ErrResourceNoSuchForm) Error() string {
	return fmt.Sprintf("no such form with name '%v' defined on resource '%v'", t.FormName, t.Resource)
}

type ErrUnsupportedMediaType struct {
	MediaType MediaType
}

func (t *ErrUnsupportedMediaType) Error() string {
	return fmt.Sprintf("media type '%v' is not supported", t.MediaType.String())
}

type ErrUnexpectedHTTPResponseStatus struct {
	ExpectedStatus int
	ActualStatus   int
	ClientRequest  *http.Request
	ClientResponse *http.Response
}

func (t *ErrUnexpectedHTTPResponseStatus) Error() string {
	return fmt.Sprintf("expected status %v received %v", t.ExpectedStatus, t.ActualStatus)
}

type ErrResourceUnmarshalFailure struct {
	UnmarshalError error
	ClientRequest  *http.Request
	ClientResponse *http.Response
}

func (t *ErrResourceUnmarshalFailure) Error() string {
	return t.UnmarshalError.Error()
}
