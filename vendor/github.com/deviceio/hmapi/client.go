package hmapi

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

type Client interface {
	Resource(path string) ResourceRequest
}

type ClientConfig struct {
	Auth       Auth
	Host       string
	HTTPClient *http.Client
	Port       int
	Scheme     *scheme
}

type client struct {
	baseuri string
	config  *ClientConfig
}

func NewClient(config *ClientConfig) Client {
	if config.Scheme == nil {
		config.Scheme = HTTP
	}

	if config.Auth == nil {
		config.Auth = &AuthNone{}
	}

	if config.Host == "" {
		config.Host = "localhost"
	}

	if config.Port == 0 {
		config.Port = 80
	}

	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	}

	return &client{
		baseuri: fmt.Sprintf(
			"%v://%v:%v",
			config.Scheme.String(),
			config.Host,
			config.Port,
		),
		config: config,
	}
}

func (t *client) Resource(path string) ResourceRequest {
	return &resourceRequest{
		client: t,
		path:   path,
	}
}

func (t *client) do(r *http.Request) (*http.Response, error) {
	t.config.Auth.Sign(r)
	return t.config.HTTPClient.Do(r)
}
