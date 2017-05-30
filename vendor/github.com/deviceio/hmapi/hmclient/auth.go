package hmclient

import "net/http"

type Auth interface {
	Sign(*http.Request)
}

type AuthNone struct{}

func (t *AuthNone) Sign(r *http.Request) {
}
