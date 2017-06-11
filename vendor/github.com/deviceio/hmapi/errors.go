package hmapi

import "fmt"

type LinkNotFound struct {
	Resource string
	LinkName string
}

func (t *LinkNotFound) Error() string {
	return fmt.Sprintf("no such link with name '%v' defined on resource '%v'", t.LinkName, t.Resource)
}

type FormNotFound struct {
	Resource string
	FormName string
}

func (t *FormNotFound) Error() string {
	return fmt.Sprintf("no such form with name '%v' defined on resource '%v'", t.FormName, t.Resource)
}

type UnsupportedMediaType struct {
	MediaType MediaType
}

func (t *UnsupportedMediaType) Error() string {
	return fmt.Sprintf("media type '%v' is not supported", t.MediaType.String())
}
