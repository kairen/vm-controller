# Sample Controller
A controller to operate VM resource in the Private Cloud(VM API) through Kubernetes. This controller is extened from [sample-controller](https://github.com/kubernetes/sample-controller).

> This repo contains an API server provides REST operations for managing VM. It implements the precondition for testing the sample-controller.

## Building from Source
Move repo into your go path under `$GOPATH/src`:
```sh
$ mkdir -p $GOPATH/src/k8s.io
$ mv sample-controller $GOPATH/src/github.com/kairen/vm-controller
$ cd $GOPATH/src/github.com/kairen/vm-controller
$ make dep
$ make
```

## Running

**Prerequisites**:
* Kubernetes cluster. Can start a cluster by Minikube: `minikube start --kubernetes-version=v1.13.5`.
* Make sure your have all the build dependencies installed for libvirt. To run on Ubuntu 16.04:

```sh
$ sudo apt-get install -y qemu-kvm libvirt-bin virtinst bridge-utils cpu-checker
$ sudo mkdir -p /var/lib/libvirt/iso
$ sudo wget http://ftp.ubuntu-tw.org/mirror/ubuntu-releases/16.04.6/ubuntu-16.04.6-server-amd64.iso -O /var/lib/libvirt/iso/ubuntu.iso
```

### API Server

#### Debug 
Run the following command as the debug mode:
```sh
$ go run cmd/apiserver/main.go --iso-path=/var/lib/libvirt/iso/ubuntu.iso
```

#### Run on Docker
To run the API server on Docker:
```sh
$ docker run -d -p 8080:8080 --restart=on-failure \
    -v /var/lib/libvirt:/var/lib/libvirt \
    -v /var/run/libvirt/libvirt-sock:/var/run/libvirt/libvirt-sock \
    --name vm-apiserver \
    kairen/apiserver:v0.1.0 --iso-path=/var/lib/libvirt/iso/ubuntu.iso
```
> You can view the API definition from [http://host:port/swagger/index.html](http://localhost:port/swagger/index.html) when the server started.

### Controller

#### Debug out of the cluster
Run the following command to debug:
```sh
$ export POD_NAME=test-1
$ export POD_NAMESPACE=kube-system
$ go run cmd/controller/main.go --kubeconfig $HOME/.kube/config --logtostderr --v=2 --api-url=http://172.22.2.68:8080
```

#### Deploy in the cluster
Run the following command to deploy the controller:
```sh
$ kubectl apply -f artifacts/deploy
$ kubectl -n kube-system get po -l app=vm-controller
```