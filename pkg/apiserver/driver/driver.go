package driver

import (
	"github.com/kairen/vm-controller/pkg/apiserver/types"
)

type Interface interface {
	List() ([]types.Server, error)
	Get(uuid string) (*types.Server, error)
	GetStatus(uuid string) (*types.ServerStatus, error)
	Create(server *types.Server) (*types.Server, error)
	Delete(uuid string) error
	CheckName(name string) int
	SetISO(path string)
	SetDiskDir(path string)
}
