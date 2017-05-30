package hmclient

type scheme string

const (
	SchemeHTTP  scheme = scheme("http")
	SchemeHTTPS scheme = scheme("https")
)
