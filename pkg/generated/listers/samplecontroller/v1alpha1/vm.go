/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	v1alpha1 "github.com/kairen/vm-controller/pkg/apis/samplecontroller/v1alpha1"
)

// VMLister helps list VMs.
type VMLister interface {
	// List lists all VMs in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.VM, err error)
	// VMs returns an object that can list and get VMs.
	VMs(namespace string) VMNamespaceLister
	VMListerExpansion
}

// vMLister implements the VMLister interface.
type vMLister struct {
	indexer cache.Indexer
}

// NewVMLister returns a new VMLister.
func NewVMLister(indexer cache.Indexer) VMLister {
	return &vMLister{indexer: indexer}
}

// List lists all VMs in the indexer.
func (s *vMLister) List(selector labels.Selector) (ret []*v1alpha1.VM, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.VM))
	})
	return ret, err
}

// VMs returns an object that can list and get VMs.
func (s *vMLister) VMs(namespace string) VMNamespaceLister {
	return vMNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// VMNamespaceLister helps list and get VMs.
type VMNamespaceLister interface {
	// List lists all VMs in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.VM, err error)
	// Get retrieves the VM from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.VM, error)
	VMNamespaceListerExpansion
}

// vMNamespaceLister implements the VMNamespaceLister
// interface.
type vMNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all VMs in the indexer for a given namespace.
func (s vMNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.VM, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.VM))
	})
	return ret, err
}

// Get retrieves the VM from the indexer for a given namespace and name.
func (s vMNamespaceLister) Get(name string) (*v1alpha1.VM, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("vm"), name)
	}
	return obj.(*v1alpha1.VM), nil
}