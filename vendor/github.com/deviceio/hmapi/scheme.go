package hmapi

var (
	HTTP  = &scheme{"http"}
	HTTPS = &scheme{"https"}
)

type scheme struct {
	string
}

func (t scheme) String() string {
	return t.string
}
