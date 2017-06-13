package sdk

import "fmt"

type Device interface {
	Filesystem() DeviceFilesystem
	System() DeviceSystem
	Network() DeviceNetwork
	Process() DeviceProcess
}

type device struct {
	id     string
	client *client
}

func (t *device) Filesystem() DeviceFilesystem {
	return &deviceFilesystem{
		device:       t,
		resourcePath: fmt.Sprintf("/device/%v/filesystem", t.id),
	}
}

func (t *device) System() DeviceSystem {
	return nil
}

func (t *device) Network() DeviceNetwork {
	return nil
}

func (t *device) Process() DeviceProcess {
	return &deviceProcess{
		device:       t,
		resourcePath: fmt.Sprintf("/device/%v/process", t.id),
	}
}
