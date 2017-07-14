package dsc

// Relation can be used as an embeded struct in a resource implementation to satisfy
// the Requires, Refreshes, Notifies methods of a dsc.Resource
type Relation struct {
	Require []string
	Refresh []string
	Notify  []string
}
