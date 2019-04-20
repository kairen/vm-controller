package kvm

import (
	"bytes"
	"crypto/rand"
	"html/template"
	"net"

	"github.com/digitalocean/go-libvirt"
	"github.com/pkg/errors"
)

const domainTemplate = `
<domain type='kvm'>
  <name>{{.Name}}</name> 
  <memory unit='MB'>{{.Memory}}</memory>
  <vcpu>{{.CPU}}</vcpu>
  <features>
    <acpi/>
    <apic/>
    <pae/>
  </features>
  <cpu mode='host-passthrough'/>
  <os>
    <type>hvm</type>
    <boot dev='cdrom'/>
    <boot dev='hd'/>
    <bootmenu enable='no'/>
  </os>
  <devices>
    <disk type='file' device='cdrom'>
      <source file='{{.ISO}}'/>
      <target dev='hdc' bus='scsi'/>
      <readonly/>
    </disk>
    <disk type='file' device='disk'>
      <driver name='qemu' type='raw' cache='default' io='threads' />
      <source file='{{.DiskPath}}'/>
      <target dev='hda' bus='virtio'/>
    </disk>
    <interface type='network'>
      <source network='default'/>
      <mac address='{{.MAC}}'/>
      <model type='virtio'/>
    </interface>
    <serial type='pty'>
      <target port='0'/>
    </serial>
    <console type='pty'>
      <target type='serial' port='0'/>
    </console>
  </devices>
</domain>
`

type Spec struct {
	Name     string
	Memory   int
	CPU      int
	ISO      string
	DiskPath string
	MAC      string
}

func randomMAC() (net.HardwareAddr, error) {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}
	buf[0] = buf[0] & 0xfc
	return buf, nil
}

func (kvm *KVM) createDomain(spec *Spec) (*libvirt.Domain, error) {
	if spec.MAC == "" {
		mac, err := randomMAC()
		if err != nil {
			return nil, errors.Wrap(err, "generating mac address")
		}
		spec.MAC = mac.String()
	}

	tmpl := template.Must(template.New("domain").Parse(domainTemplate))
	var xml bytes.Buffer
	if err := tmpl.Execute(&xml, spec); err != nil {
		return nil, errors.Wrap(err, "executing domain xml")
	}

	var flags libvirt.DomainDefineFlags
	domain, err := kvm.virt.DomainDefineXMLFlags(xml.String(), flags)
	if err != nil {
		return nil, errors.Wrapf(err, "Error defining domain xml: %s", xml.String())
	}
	return &domain, nil
}
