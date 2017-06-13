package sdk

import (
	"crypto/tls"
	"net/http"

	"github.com/deviceio/hmapi"
)

type Client interface {
	Device(deviceid string) Device
}

type ClientConfig struct {
	HubHost    string
	HubPort    int
	UserID     string
	TOTPSecret string
	PrivateKey string
	HMClient   hmapi.Client
}

type client struct {
	hmclient hmapi.Client
}

func NewClient(config ClientConfig) Client {
	if config.HMClient == nil {
		config.HMClient = hmapi.NewClient(&hmapi.ClientConfig{
			Host:   config.HubHost,
			Port:   config.HubPort,
			Scheme: hmapi.HTTPS,
			HTTPClient: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			},
			Auth: &ClientAuth{
				UserID:         config.UserID,
				UserTOTPSecret: config.TOTPSecret,
				UserPrivateKey: config.PrivateKey,
			},
		})
	}

	return &client{
		hmclient: config.HMClient,
	}
}

func (t *client) Device(deviceid string) Device {
	return &device{
		id:     deviceid,
		client: t,
	}
}
