package vm

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	samplev1alpha1 "github.com/kairen/vm-controller/pkg/apis/samplecontroller/v1alpha1"
	"github.com/kairen/vm-controller/pkg/apiserver/types"
	vmrest "github.com/kairen/vm-controller/pkg/client/rest"
	vmclientset "github.com/kairen/vm-controller/pkg/client/vm"
	"github.com/kairen/vm-controller/pkg/constants"
	clientset "github.com/kairen/vm-controller/pkg/generated/clientset/versioned"
	informers "github.com/kairen/vm-controller/pkg/generated/informers/externalversions/samplecontroller/v1alpha1"
	listers "github.com/kairen/vm-controller/pkg/generated/listers/samplecontroller/v1alpha1"
	"github.com/kairen/vm-controller/pkg/util"
)

const (
	SuccessSynced         = "Synced"
	ErrResourceDeleting   = "ErrResourceDeleting"
	MessageResourceSynced = "VM synced successfully"
)

type Controller struct {
	vmclientset     vmclientset.Interface
	sampleclientset clientset.Interface

	vmsLister listers.VMLister
	vmsSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
	recorder  record.EventRecorder
}

func NewController(
	vmclientset vmclientset.Interface,
	sampleclientset clientset.Interface,
	vmInformer informers.VMInformer,
	recorder record.EventRecorder) *Controller {
	// Create an instance for the Foo controller
	controller := &Controller{
		vmclientset:     vmclientset,
		sampleclientset: sampleclientset,
		vmsLister:       vmInformer.Lister(),
		vmsSynced:       vmInformer.Informer().HasSynced,
		workqueue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "VMs"),
		recorder:        recorder,
	}

	klog.Info("Setting up the VM event handlers")
	vmInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueVM,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueVM(new)
		},
	})
	return controller
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("Starting VM controller")
	klog.Info("Waiting for VM informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.vmsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting VM workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started VM workers")
	<-stopCh
	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("VM controller expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.syncHandler(key); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("VM controller error syncing '%s': %s, requeuing", key, err.Error())
		}

		c.workqueue.Forget(obj)
		klog.Infof("VM controller successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	vm, err := c.vmsLister.VMs(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("vm '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	switch vm.Status.Phase {
	case samplev1alpha1.VMNone:
		if err := c.makePendingPhase(vm); err != nil {
			return err
		}
	case samplev1alpha1.VMPending:
		if err := c.createServer(vm); err != nil {
			return err
		}
	case samplev1alpha1.VMActive:
		if !vm.ObjectMeta.DeletionTimestamp.IsZero() {
			if err := c.deleteServer(vm); err != nil {
				return err
			}
		} else {
			if err := c.updateUtilization(vm); err != nil {
				return err
			}
		}
	case samplev1alpha1.VMTerminating:
		err := c.sampleclientset.SamplecontrollerV1alpha1().VMs(vm.Namespace).Delete(vm.Name, nil)
		if err != nil {
			return err
		}
	}
	c.recorder.Event(vm, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) makePendingPhase(vm *samplev1alpha1.VM) error {
	vmCopy := vm.DeepCopy()
	return c.updateVMStatus(vmCopy, samplev1alpha1.VMPending, nil)
}

func (c *Controller) createServer(vm *samplev1alpha1.VM) error {
	vmCopy := vm.DeepCopy()
	_, ok := c.vmclientset.V1Alpha1().Server().CheckName(vm.Spec.VMName).(*vmrest.ResponeError)
	if ok {
		reason := fmt.Errorf("VM name is prohibit to use for some reason")
		if err := c.updateVMStatus(vmCopy, samplev1alpha1.VMFailed, reason); err != nil {
			return err
		}
	}

	server := &types.Server{
		Name:     vmCopy.Spec.VMName,
		CPU:      int(vmCopy.Spec.CPU),
		Memory:   int(vmCopy.Spec.Memory),
		DiskSize: int(vmCopy.Spec.DiskSize),
	}
	create, reason := c.vmclientset.V1Alpha1().Server().Create(server)
	if reason != nil {
		if err := c.updateVMStatus(vmCopy, samplev1alpha1.VMFailed, reason); err != nil {
			return err
		}
		return reason
	}

	vmCopy.Status.ID = create.UUID
	vmCopy.SetFinalizers([]string{constants.DefaultFinalizer})
	if err := c.updateVMStatus(vmCopy, samplev1alpha1.VMActive, nil); err != nil {
		return err
	}
	return nil
}

func (c *Controller) deleteServer(vm *samplev1alpha1.VM) error {
	vmCopy := vm.DeepCopy()
	err := c.vmclientset.V1Alpha1().Server().Delete(vmCopy.Status.ID)
	if err != nil {
		msg := fmt.Sprintf("Failed to delete server by VM API.")
		c.recorder.Event(vm, corev1.EventTypeWarning, ErrResourceDeleting, msg)
		return err
	}

	vmCopy.SetFinalizers([]string{})
	if err := c.updateVMStatus(vmCopy, samplev1alpha1.VMTerminating, nil); err != nil {
		return err
	}
	return nil
}

func (c *Controller) updateUtilization(vm *samplev1alpha1.VM) error {
	vmCopy := vm.DeepCopy()
	t := util.SubtractTime(vmCopy.Status.LastUpdateTime.Time)
	if t.Seconds() > constants.PeriodSec {
		stats, err := c.vmclientset.V1Alpha1().Server().GetStatusByUUID(vmCopy.Status.ID)
		if err != nil {
			return err
		}

		vmCopy.Status.CPUUtilization = int32(stats.CPUUtilization)
		if err := c.updateVMStatus(vmCopy, samplev1alpha1.VMActive, nil); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) updateVMStatus(vm *samplev1alpha1.VM, phase samplev1alpha1.VMPhase, reason error) error {
	vm.Status.Reason = ""
	if reason != nil {
		vm.Status.Reason = reason.Error()
	}

	vm.Status.Phase = phase
	vm.Status.LastUpdateTime = metav1.NewTime(time.Now())
	_, err := c.sampleclientset.SamplecontrollerV1alpha1().VMs(vm.Namespace).Update(vm)
	return err
}

func (c *Controller) enqueueVM(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}
