package operator

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"

	vmclientset "github.com/kairen/vm-controller/pkg/client/vm"
	clientset "github.com/kairen/vm-controller/pkg/generated/clientset/versioned"
	samplescheme "github.com/kairen/vm-controller/pkg/generated/clientset/versioned/scheme"
	informers "github.com/kairen/vm-controller/pkg/generated/informers/externalversions"
	"github.com/kairen/vm-controller/pkg/operator/foo"
	"github.com/kairen/vm-controller/pkg/operator/vm"
)

const agentName = "sample-controller"

type Operator struct {
	vmclientset     vmclientset.Interface
	kubeclientset   kubernetes.Interface
	sampleclientset clientset.Interface

	kubeInformer kubeinformers.SharedInformerFactory
	informer     informers.SharedInformerFactory

	fooController *foo.Controller
	vmController  *vm.Controller
}

func NewMainOperator(kubeclientset kubernetes.Interface, sampleclientset clientset.Interface, vmclientset vmclientset.Interface) *Operator {
	return &Operator{
		kubeclientset:   kubeclientset,
		sampleclientset: sampleclientset,
		vmclientset:     vmclientset,
	}
}

func (op *Operator) Initialize() error {
	op.kubeInformer = kubeinformers.NewSharedInformerFactory(op.kubeclientset, time.Second*30)
	op.informer = informers.NewSharedInformerFactory(op.sampleclientset, time.Second*30)

	utilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: op.kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: agentName})

	op.fooController = foo.NewController(op.kubeclientset, op.sampleclientset,
		op.kubeInformer.Apps().V1().Deployments(),
		op.informer.Samplecontroller().V1alpha1().Foos(), recorder)
	op.vmController = vm.NewController(op.vmclientset, op.sampleclientset,
		op.informer.Samplecontroller().V1alpha1().VMs(), recorder)
	return nil
}

func (op *Operator) Run(stopCh <-chan struct{}) error {
	op.kubeInformer.Start(stopCh)
	op.informer.Start(stopCh)

	var err error
	go func() { err = op.fooController.Run(2, stopCh) }()
	if err != nil {
		return fmt.Errorf("failed to run Foo controller: %s", err.Error())
	}

	go func() { err = op.vmController.Run(2, stopCh) }()
	if err != nil {
		return fmt.Errorf("failed to run VM controller: %s", err.Error())
	}

	<-stopCh
	klog.Info("Shutting down controllers")
	return nil
}
