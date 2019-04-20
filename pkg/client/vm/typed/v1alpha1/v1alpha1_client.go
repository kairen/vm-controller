package v1alpha1

import (
	"github.com/kairen/vm-controller/pkg/client/rest"
)

const V1Alpha1 = "api/v1alpha1"

type Interface interface {
	RESTClient() rest.Interface
	ServerGetter
}

type V1Alpha1Client struct {
	restClient rest.Interface
	config     *rest.Config
}

func (c *V1Alpha1Client) Server() ServerInterface {
	return newServer(c)
}

func NewForConfig(c *rest.Config) (*V1Alpha1Client, error) {
	config := *c
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &V1Alpha1Client{client, &config}, nil
}

func (c *V1Alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
