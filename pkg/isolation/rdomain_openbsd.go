package isolation

import "golang.zx2c4.com/wireguard/wgctrl/wgtypes"

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

func (r RdomainIsolation) LinkEnsure(ifname string, config wgtypes.Config, mtu, ifgroup int) error {
	panic("to be implemented")
}

func (r RdomainIsolation) LinkAbsent(ifname string) error {
	panic("to be implemented")
}

func (r RdomainIsolation) LinkFilter(prefix string, ifgroup int) ([]string, error) {
	panic("to be implemented")
}
