package hmclient

import (
	"fmt"

	"github.com/deviceio/hmapi"
)

type UnsupportedMediaType struct {
	MediaType hmapi.MediaType
}

func (t *UnsupportedMediaType) Error() string {
	return fmt.Sprintf("%v is not supported", t.MediaType.String())
}
