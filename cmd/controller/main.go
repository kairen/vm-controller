/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"context"
	goflag "flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	flag "github.com/spf13/pflag"

	_ "k8s.io/kubernetes/pkg/client/metrics/prometheus" // for client metric registration
	_ "k8s.io/kubernetes/pkg/util/reflector/prometheus" // for reflector metric registration
	_ "k8s.io/kubernetes/pkg/util/workqueue/prometheus" // for workqueue metric registration
	_ "k8s.io/kubernetes/pkg/version/prometheus"        // for version metric registration

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"

	vmrest "github.com/kairen/vm-controller/pkg/client/rest"
	vmclientset "github.com/kairen/vm-controller/pkg/client/vm"
	"github.com/kairen/vm-controller/pkg/constants"
	clientset "github.com/kairen/vm-controller/pkg/generated/clientset/versioned"
	"github.com/kairen/vm-controller/pkg/k8sutil"
	"github.com/kairen/vm-controller/pkg/operator"
	"github.com/kairen/vm-controller/pkg/util/probe"
	"github.com/kairen/vm-controller/pkg/version"
)

var (
	kubeconfig string
	listenAddr string
	apiURL     string
	namespace  string
	name       string
	ver        bool
)

func parserFlags() {
	flag.StringVarP(&kubeconfig, "kubeconfig", "", "", "Absolute path to the kubeconfig file.")
	flag.StringVarP(&apiURL, "api-url", "", "http://127.0.0.1:8080", "The VM apiserver URL(http://host:port).")
	flag.StringVarP(&listenAddr, "listen-addr", "", "0.0.0.0:8081", "The address on which the HTTP server will listen to.")
	flag.BoolVarP(&ver, "version", "", false, "Display the version.")
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()
}

func main() {
	klog.InitFlags(nil)
	parserFlags()

	if ver {
		fmt.Fprintf(os.Stdout, "%s\n", version.GetVersion())
		os.Exit(0)
	}

	name := os.Getenv(constants.EnvPodName)
	if len(name) == 0 {
		klog.Fatalf("must set env (%s)", constants.EnvPodNamespace)
	}

	namespace = os.Getenv(constants.EnvPodNamespace)
	if len(namespace) == 0 {
		klog.Fatalf("must set env (%s)", constants.EnvPodName)
	}

	id, err := os.Hostname()
	if err != nil {
		klog.Fatalf("failed to get hostname: %v", err)
	}

	http.HandleFunc(probe.HTTPHealthzEndpoint, probe.HealthzHandler)
	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(listenAddr, nil)

	vmcfg := vmrest.NewConfig(apiURL)
	vmclient, err := vmclientset.NewForConfig(vmcfg)
	if err != nil {
		klog.Fatalf("Failed to build vm clientset: %s", err.Error())
	}

	cfg, err := k8sutil.GetRestConfig(kubeconfig)
	if err != nil {
		klog.Fatalf("Failed to build kubeconfig: %s", err.Error())
	}

	kubeclient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Failed to build Kubernetes client: %s", err.Error())
	}

	sampleclient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Failed to build sample client: %s", err.Error())
	}

	lock, err := resourcelock.New(resourcelock.EndpointsResourceLock,
		namespace,
		"vm-controller",
		kubeclient.CoreV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: createRecorder(kubeclient, name, namespace),
		})
	if err != nil {
		klog.Fatalf("Failed to create lock: %v", err)
	}

	stopCh := make(chan struct{}, 1)
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:          lock,
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				klog.V(3).Infoln("Started leading.")
				ctx, cancel = context.WithCancel(ctx)
				defer cancel()

				op := operator.NewMainOperator(kubeclient, sampleclient, vmclient)
				if err := op.Initialize(); err != nil {
					klog.Fatalf("Error initing operator: %s", err.Error())
				}

				if err := op.Run(stopCh); err != nil {
					klog.Fatalf("Error starting operator: %s", err.Error())
				}
				klog.Infoln("close.")
			},
			OnStoppedLeading: func() {
				klog.Fatalf("The leader election lost.")
			},
		},
	})

	for {
		select {
		case <-signalCh:
			klog.Infof("Shutdown signal received, exiting...")
			close(stopCh)
			panic("Unreachable...")
		}
	}
}

func createRecorder(kubeclient kubernetes.Interface, name, namespace string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: v1core.New(kubeclient.Core().RESTClient()).Events(namespace)})
	return eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: name})
}
