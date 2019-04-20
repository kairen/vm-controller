package v1alpha1

import (
	"github.com/kairen/vm-controller/pkg/apiserver/types"
	"github.com/kairen/vm-controller/pkg/client/rest"
)

type ServerGetter interface {
	Server() ServerInterface
}

type ServerInterface interface {
	List() ([]types.Server, error)
	Create(obj *types.Server) (*types.Server, error)
	GetByUUID(uuid string) (*types.Server, error)
	GetStatusByUUID(uuid string) (*types.ServerStatus, error)
	CheckName(name string) error
	Delete(uuid string) error
}

// server implements ServerInterface
type server struct {
	client rest.Interface
}

// newServer returns a server
func newServer(c *V1Alpha1Client) *server {
	return &server{
		client: c.RESTClient(),
	}
}

func (c *server) Create(obj *types.Server) (*types.Server, error) {
	result := &types.Server{}
	err := c.client.Post(obj).
		Suffix(V1Alpha1, "servers").
		Do().Into(result)
	return result, err
}

func (c *server) List() ([]types.Server, error) {
	result := []types.Server{}
	err := c.client.Get().
		Suffix(V1Alpha1, "servers").
		Do().Into(&result)
	return result, err
}

func (c *server) GetByUUID(uuid string) (*types.Server, error) {
	result := &types.Server{}
	err := c.client.Get().
		Suffix(V1Alpha1, "servers", uuid).
		Do().Into(result)
	return result, err
}

func (c *server) GetStatusByUUID(uuid string) (*types.ServerStatus, error) {
	result := &types.ServerStatus{}
	err := c.client.Get().
		Suffix(V1Alpha1, "servers", uuid, "status").
		Do().Into(result)
	return result, err
}

func (c *server) CheckName(name string) error {
	err := c.client.Get().
		Suffix(V1Alpha1, "check", name).
		Do().Into(nil)
	return err
}

func (c *server) Delete(uuid string) error {
	err := c.client.Delete(nil).
		Suffix(V1Alpha1, "servers", uuid).
		Do().Into(nil)
	return err
}
