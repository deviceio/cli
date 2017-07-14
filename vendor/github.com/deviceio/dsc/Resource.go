package dsc

// Resource represents the minimal interface dsc needs to apply content enforcement.
// Any type implimenting this interface can be used within a dsc module.
type Resource interface {
	// Requires returns a slice of named resources the Module will apply before
	// this resource is applied.
	Requires() (names []string)

	// Notifies returns a slice of named resources that the Module will apply after
	// this resource has been applied.
	Notifies() (names []string)

	// Refreshes returns a slice of named resources that the Module will refresh
	// after this resource has been applied OR after this resource has been refreshed.
	Refreshes() (names []string)

	// Apply is used to enforce the resource by the Module
	Apply() (changed bool, err error)

	// Refresh is used to refresh the resource by the Module. Refresh is resource
	// dependant functionality and many resources may choose to do nothing. Refresh
	// is useful apply a minor state change to a resource where it is applicable
	// ex: refreshing a Service resource will cause the target Service to restart
	Refresh() (err error)
}
