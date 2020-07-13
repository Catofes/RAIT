package isolation

import (
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func init() {
	Register("rdomain", NewRdomainIsolation)
}

type RdomainIsolation struct {
	TransitDomain   string
	InterfaceDomain string
}

func NewRdomainIsolation(transitDomain, interfaceDomain string) (Isolation, error) {
	return &RdomainIsolation{
		TransitDomain:   transitDomain,
		InterfaceDomain: interfaceDomain,
	}, nil
}

func (r RdomainIsolation) LinkEnsure(attrs *LinkAttrs, config wgtypes.Config) error {
	panic("to be implemented")
}

func (r RdomainIsolation) LinkAbsent(link *LinkAttrs) error {
	panic("to be implemented")
}

func (r RdomainIsolation) LinkList() ([]*LinkAttrs, error) {
	panic("to be implemented")
}
