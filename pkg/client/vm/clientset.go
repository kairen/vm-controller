package vm

import (
	"github.com/kairen/vm-controller/pkg/client/rest"
	"github.com/kairen/vm-controller/pkg/client/vm/typed/v1alpha1"
)

type Interface interface {
	V1Alpha1() v1alpha1.Interface
}

type Clientset struct {
	v1alpha1 *v1alpha1.V1Alpha1Client
}

func (c *Clientset) V1Alpha1() v1alpha1.Interface {
	return c.v1alpha1
}

func NewForConfig(c *rest.Config) (*Clientset, error) {
	config := *c

	var cs Clientset
	var err error

	cs.v1alpha1, err = v1alpha1.NewForConfig(&config)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}
