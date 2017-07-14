package dsc

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/sys/windows/svc"
	svcmgr "golang.org/x/sys/windows/svc/mgr"
)

// Apply enforces this Service on a windows system
func (t *Service) Apply() (bool, error) {
	var absent bool
	var args []string
	var changed = false
	var config svcmgr.Config
	var description string
	var err error
	var mgr *svcmgr.Mgr
	var name string
	var path string
	var service *svcmgr.Service
	var started bool

	if absent, err = t.absent(); err != nil {
		return false, err
	}

	if name, err = t.name(); err != nil {
		return false, err
	}

	if description, err = t.description(); err != nil {
		return false, err
	}

	if path, err = t.path(); err != nil {
		return false, err
	}

	if args, err = t.args(); err != nil {
		return false, err
	}

	if started, err = t.started(); err != nil {
		return false, err
	}

	pathAndArgs := strings.Join(
		append([]string{path}, args...),
		" ",
	)

	if mgr, err = svcmgr.Connect(); err != nil {
		return false, err
	}
	defer mgr.Disconnect()

	if service, err = mgr.OpenService(name); err != nil && !absent {
		if service, err = mgr.CreateService(name, path, svcmgr.Config{
			StartType:   svcmgr.StartAutomatic,
			DisplayName: name,
			Description: name,
		}, args...); err != nil {
			return false, nil
		}

		changed = true

		return changed, nil
	}
	defer service.Close()

	if absent {
		if err = service.Delete(); err != nil {
			return false, err
		}

		changed = true

		return changed, nil
	}

	if config, err = service.Config(); err != nil {
		return false, err
	}

	if config.BinaryPathName != pathAndArgs {
		changed = true

		config.BinaryPathName = pathAndArgs

		if err = service.UpdateConfig(config); err != nil {
			return false, err
		}
	}

	if config.DisplayName != name {
		changed = true

		config.DisplayName = name

		if err = service.UpdateConfig(config); err != nil {
			return false, err
		}
	}

	if config.Description != description {
		changed = true

		config.Description = description

		if err = service.UpdateConfig(config); err != nil {
			return false, err
		}
	}

	if started {
		var status svc.Status

		if status, err = service.Query(); err != nil {
			return false, err
		}

		if status.State != svc.Running {
			if err = service.Start([]string{}...); err != nil {
				return false, err
			}
		}
	} else {
		var status svc.Status

		if status, err = service.Query(); err != nil {
			return false, err
		}

		timeout := time.Now().Add(10 * time.Second)
		for status.State != svc.Stopped {
			if timeout.Before(time.Now()) {
				return false, fmt.Errorf("timeout waiting for service to go to state=%d", svc.Stopped)
			}
			time.Sleep(300 * time.Millisecond)

			if status, err = service.Query(); err != nil {
				return false, fmt.Errorf("could not retrieve service status: %v", err)
			}
		}
	}

	return changed, nil
}

// Refresh TODO: impliment service restart
func (t *Service) Refresh() error {
	return nil
}

// Requires implementation for dsc.Resource
func (t *Service) Requires() []string {
	if t.Relation == nil || t.Relation.Require == nil {
		return []string{}
	}

	return t.Relation.Require
}

// Notifies implementation for dsc.Resource
func (t *Service) Notifies() []string {
	if t.Relation == nil || t.Relation.Notify == nil {
		return []string{}
	}

	return t.Relation.Notify
}

// Refreshes implementation for dsc.Resource
func (t *Service) Refreshes() []string {
	if t.Relation == nil || t.Relation.Refresh == nil {
		return []string{}
	}

	return t.Relation.Refresh
}
