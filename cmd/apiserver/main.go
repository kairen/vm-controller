package main

import (
	"context"

	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/kairen/vm-controller/pkg/apiserver/driver"
	"github.com/kairen/vm-controller/pkg/apiserver/driver/kvm"
	"github.com/kairen/vm-controller/pkg/apiserver/handlers/v1alpha1"
	"github.com/kairen/vm-controller/pkg/apiserver/router"

	flag "github.com/spf13/pflag"
)

const (
	readTimeout    = 10 * time.Second
	writeTimeout   = 10 * time.Second
	initRetryDelay = 5 * time.Second
	maxHeaderBytes = 1 << 20
)

var (
	addr        string
	swagger     bool
	vmdriver    string
	protocol    string
	address     string
	isoPath     string
	diskDirPath string
)

func parseFlags() {
	flag.StringVarP(&addr, "listen", "", ":8080", "API server listen address.")
	flag.StringVarP(&vmdriver, "vm-driver", "", "kvm", "VM driver is one of: [kvm qemu].")
	flag.StringVarP(&protocol, "remote-protocol", "", "unix", "KVM remote protocol is one of: [unix tcp].")
	flag.StringVarP(&address, "remote-address", "", "/var/run/libvirt/libvirt-sock", "KVM remote address(path or host address).")
	flag.StringVarP(&isoPath, "iso-path", "", "/var/lib/libvirt/iso/ubuntu.iso", "Boot ISO file path.")
	flag.StringVarP(&diskDirPath, "disk-dir-path", "", "/var/lib/libvirt/images", "VM volume dir path.")
	flag.BoolVarP(&swagger, "swagger", "", true, "Enable swagger API page.")
	flag.Parse()
}

func main() {
	parseFlags()

	log.SetPrefix("[VM REST] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Ltime)

	var driver driver.Interface
	switch vmdriver {
	case "kvm":
		driver = kvm.NewKVM(protocol, address)
		driver.SetDiskDir(diskDirPath)
		driver.SetISO(isoPath)
	}

	r := router.New()
	r.LinkSwaggerAPI(swagger)
	r.LinkHandler(v1alpha1.New(driver))

	server := &http.Server{
		Addr:           addr,
		Handler:        r.GetEngine(),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}

	go func() {
		log.Println("Server starting...")
		if err := server.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Failed to shutdown server:", err)
	}
	log.Println("Server exiting...")
}
