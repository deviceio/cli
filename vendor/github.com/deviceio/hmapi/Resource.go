package hmapi

type Resource struct {
	Links   map[string]*Link    `json:"links,omitempty"`
	Forms   map[string]*Form    `json:"forms,omitempty"`
	Content map[string]*Content `json:"content,omitempty"`
}
