package dsc

import (
	"errors"
	"fmt"
	"log"
)

// Module groups multiple Resources together to be enforced. dsc can then apply
// the module as a whole.
type Module struct {
	// resources contains the original resource map as supplied by the user
	resources map[string]Resource

	// done contains resources that have completed enforcement
	done map[string]Resource
}

// NewModule creates a new module from a map of resources.
func NewModule(resources map[string]Resource) *Module {
	return &Module{
		resources: resources,
		done:      map[string]Resource{},
	}
}

// Run analyzes the resource graph and starts the enforcement process
func (t *Module) Run() error {
	for k1, v1 := range t.resources {
		if _, err := t.apply(k1, v1); err != nil {
			return err
		}
	}

	return nil
}

func (t *Module) apply(name string, r Resource) (bool, error) {
	require := r.Requires()

	for _, required := range require {
		if _, ok := t.resources[required]; !ok {
			return false, errors.New(fmt.Sprintf(
				"Required resource '%v' not found in module",
				required,
			))
		}
	}

	notify := r.Notifies()

	for _, notified := range notify {
		if _, ok := t.resources[notified]; !ok {
			return false, errors.New(fmt.Sprintf(
				"Notified resource '%v' not found in module",
				notified,
			))
		}
	}

	refresh := r.Refreshes()

	for _, refreshed := range refresh {
		if _, ok := t.resources[refreshed]; !ok {
			return false, errors.New(fmt.Sprintf(
				"Refreshed resource '%v' not found in module",
				refreshed,
			))
		}
	}

	for _, required := range require {
		res, _ := t.resources[required]

		if _, err := t.apply(required, res); err != nil {
			return false, err
		}
	}

	hasChanged := false

	if _, ok := t.done[name]; !ok {
		log.Println(fmt.Sprintf("Applying Resource '%v'", name))

		changed, err := r.Apply()
		if err != nil {
			return false, err
		}

		t.done[name] = r
		hasChanged = changed
	}

	if hasChanged {
		for _, refreshed := range refresh {
			res, _ := t.resources[refreshed]

			if _, err := t.apply(refreshed, res); err != nil {
				return false, err
			}
		}
	}

	return hasChanged, nil
}
