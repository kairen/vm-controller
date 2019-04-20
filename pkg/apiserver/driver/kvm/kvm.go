package kvm

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	libvirt "github.com/digitalocean/go-libvirt"
	qemu "github.com/quadrifoglio/go-qemu"
	"github.com/kairen/vm-controller/pkg/apiserver/types"
)

var stateMsg = map[libvirt.DomainState]types.ServerState{
	libvirt.DomainNostate:     types.ServerNoState,
	libvirt.DomainRunning:     types.ServerRunningState,
	libvirt.DomainBlocked:     types.ServerBlockedState,
	libvirt.DomainPaused:      types.ServerPauseState,
	libvirt.DomainShutdown:    types.ServerShutdownState,
	libvirt.DomainShutoff:     types.ServerShutoffState,
	libvirt.DomainCrashed:     types.ServerCrashedState,
	libvirt.DomainPmsuspended: types.ServerPmsuspendedState,
}

type KVM struct {
	virt *libvirt.Libvirt

	protocol string
	address  string

	isoFile     string
	diskDirPath string
}

func NewKVM(protocol, address string) *KVM {
	return &KVM{protocol: protocol, address: address}
}

func (kvm *KVM) SetISO(filePath string) {
	kvm.isoFile = filePath
}

func (kvm *KVM) SetDiskDir(path string) {
	kvm.diskDirPath = path
}

func (kvm *KVM) connect() error {
	c, err := net.DialTimeout(kvm.protocol, kvm.address, 5*time.Second)
	if err != nil {
		return err
	}

	kvm.virt = libvirt.New(c)
	if err := kvm.virt.Connect(); err != nil {
		return err
	}
	return nil
}

func (kvm *KVM) disconnect() error {
	if err := kvm.virt.Disconnect(); err != nil {
		return err
	}
	return nil
}

func (kvm *KVM) Create(s *types.Server) (*types.Server, error) {
	if err := kvm.connect(); err != nil {
		return nil, err
	}
	defer kvm.disconnect()

	spec := &Spec{
		Name:     s.Name,
		Memory:   s.Memory,
		CPU:      s.CPU,
		DiskPath: fmt.Sprintf("%s/%s.qcow2", kvm.diskDirPath, s.Name),
		ISO:      kvm.isoFile,
	}

	img := qemu.NewImage(spec.DiskPath, qemu.ImageFormatQCOW2, uint64(s.DiskSize*1073741824))
	err := img.Create()
	if err != nil {
		log.Fatal(err)
	}

	d, err := kvm.createDomain(spec)
	if err != nil {
		return nil, err
	}

	var flags uint32
	if _, err := kvm.virt.DomainCreateWithFlags(*d, flags); err != nil {
		return nil, err
	}

	server := &types.Server{
		ID:   d.ID,
		UUID: uuidToString(d.UUID),
		Name: d.Name,
	}
	return server, nil
}

func (kvm *KVM) List() ([]types.Server, error) {
	if err := kvm.connect(); err != nil {
		return nil, err
	}
	defer kvm.disconnect()

	servers := []types.Server{}
	domains, err := kvm.virt.Domains()
	if err != nil {
		return nil, err
	}

	for _, d := range domains {
		uuid := uuidToString(d.UUID)
		server := types.Server{ID: d.ID, Name: d.Name, UUID: uuid}
		servers = append(servers, server)
	}
	return servers, nil
}

func (kvm *KVM) Get(uuid string) (*types.Server, error) {
	if err := kvm.connect(); err != nil {
		return nil, err
	}
	defer kvm.disconnect()

	domains, err := kvm.virt.Domains()
	if err != nil {
		return nil, err
	}

	for _, d := range domains {
		vmUUID := uuidToString(d.UUID)
		if uuid == vmUUID {
			return &types.Server{ID: d.ID, Name: d.Name, UUID: vmUUID}, nil
		}
	}
	return nil, nil
}

func (kvm *KVM) GetStatus(uuid string) (*types.ServerStatus, error) {
	if err := kvm.connect(); err != nil {
		return nil, err
	}
	defer kvm.disconnect()

	domains, err := kvm.virt.Domains()
	if err != nil {
		return nil, err
	}

	var name string
	for _, d := range domains {
		vmUUID := uuidToString(d.UUID)
		if uuid == vmUUID {
			name = d.Name
		}
	}

	if len(name) == 0 {
		return nil, nil
	}

	state, err := kvm.virt.DomainState(name)
	if err != nil {
		return nil, err
	}

	// TODO(k2r2): Get stats to calculate result.
	// result = 100 * (cpu_time 2 - cpu_time 1) / n sec
	status := &types.ServerStatus{
		CPUUtilization: rand.Intn(20) + 5,
		State:          stateMsg[state],
	}
	return status, nil
}

func (kvm *KVM) Delete(uuid string) error {
	if err := kvm.connect(); err != nil {
		return err
	}
	defer kvm.disconnect()

	domains, err := kvm.virt.Domains()
	if err != nil {
		return err
	}

	var name string
	for _, d := range domains {
		vmUUID := uuidToString(d.UUID)
		if uuid == vmUUID {
			name = d.Name
		}
	}

	if err := kvm.virt.Destroy(name, libvirt.DomainDestroyGraceful); err != nil {
		return err
	}

	if err := kvm.virt.Undefine(name, libvirt.DomainUndefineManagedSave); err != nil {
		return err
	}
	return nil
}

func (kvm *KVM) CheckName(name string) int {
	if err := kvm.connect(); err != nil {
		return -1
	}
	defer kvm.disconnect()

	domains, err := kvm.virt.Domains()
	if err != nil {
		return -1
	}

	for _, d := range domains {
		if name == d.Name {
			return 0
		}
	}
	return 1
}

func uuidToString(uuid libvirt.UUID) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
