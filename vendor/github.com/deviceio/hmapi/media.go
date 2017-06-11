package hmapi

const (
	MediaTypeHMAPIResource     = MediaType("application/vnd.hmapi.Resource+json")
	MediaTypeHMAPIBoolean      = MediaType("application/vnd.hmapi.Bool")
	MediaTypeHMAPIFloat32      = MediaType("application/vnd.hmapi.Float32")
	MediaTypeHMAPIFloat64      = MediaType("application/vnd.hmapi.Float64")
	MediaTypeHMAPIInt          = MediaType("application/vnd.hmapi.Int")
	MediaTypeHMAPIInt32        = MediaType("application/vnd.hmapi.Int32")
	MediaTypeHMAPIInt64        = MediaType("application/vnd.hmapi.Int64")
	MediaTypeHMAPIString       = MediaType("application/vnd.hmapi.String")
	MediaTypeHMAPIUInt         = MediaType("application/vnd.hmapi.UInt")
	MediaTypeHMAPIUInt32       = MediaType("application/vnd.hmapi.UInt32")
	MediaTypeHMAPIUInt64       = MediaType("application/vnd.hmapi.UInt64")
	MediaTypeOctetStream       = MediaType("application/octet-stream")
	MediaTypeJSON              = MediaType("application/json")
	MediaTypeTextPlain         = MediaType("text/plain")
	MediaTypeMultipartFormData = MediaType(`multipart/form-data;boundary="hmapi_boundry_E58FCE5B6201466A8A9A6ECCDFBD31D3"`)
)

type MediaType string

func (t MediaType) String() string {
	return string(t)
}
