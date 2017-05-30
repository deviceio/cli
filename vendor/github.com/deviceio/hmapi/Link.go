package hmapi

// Link represents a link to a resource that can be navigated via a standard HTTP GET
// request.
type Link struct {
	// Href provides the resource URI that servies the resources
	Href string `json:"href,omitempty"`

	// Type indicates the MediaType of the resource
	Type MediaType `json:"type,omitempty"`

	// Encoding indicates the encoding of the Type provided as a media type
	Encoding MediaType `json:"encoding,omitempty"`
}
