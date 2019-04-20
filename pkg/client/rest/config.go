package rest

import (
	"net/http"
	"time"
)

var defaultTimeout = time.Second * 30

type Config struct {
	// Host must be a host string, a host:port pair, or a URL to the base of the apiserver.
	Host string

	// The maximum length of time to wait before giving up on a server request. A value of zero means no timeout.
	Timeout time.Duration
}

func NewConfig(host string) *Config {
	return &Config{
		Host:    host,
		Timeout: defaultTimeout,
	}
}

func RESTClientFor(config *Config) (*RESTClient, error) {
	httpClient := http.DefaultClient
	if config.Timeout > 0 {
		httpClient.Timeout = config.Timeout
	}
	return NewRESTClient(config, httpClient)
}
