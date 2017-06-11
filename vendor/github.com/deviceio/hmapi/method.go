package hmapi

type method string

func (t method) String() string {
	return string(t)
}

const (
	CONNECT = method("CONNECT")
	DELETE  = method("DELETE")
	GET     = method("GET")
	HEAD    = method("HEAD")
	OPTIONS = method("OPTIONS")
	PATCH   = method("PATCH")
	POST    = method("POST")
	PUT     = method("PUT")
	TRACE   = method("TRACE")
)
