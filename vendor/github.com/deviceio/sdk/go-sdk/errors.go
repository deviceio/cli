package sdk

import "fmt"

type ErrInvalidAPIResponse struct {
	StatusCode int
	Message    string
}

func (t *ErrInvalidAPIResponse) Error() string {
	return fmt.Sprintf("StatusCode: %v Message: %v", t.StatusCode, t.Message)
}
