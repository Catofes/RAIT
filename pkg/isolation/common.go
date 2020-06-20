package isolation

import (
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Isolation interface {
	LinkEnsure(name string, mtu, group int, config wgtypes.Config) error
	LinkAbsent(name string) error
	LinkList(prefix string, group int) ([]string, error)
}

type GenericIsolation struct {
	Isolation
}

func NewGenericIsolation(kind, transitNamespace, interfaceNamespace string) (*GenericIsolation, error) {
	switch kind {
	case "netns":
		return &GenericIsolation{Isolation: NewNetnsIsolation(transitNamespace, interfaceNamespace)}, nil
	case "vrf":
		return &GenericIsolation{Isolation: NewVrfIsolation(transitNamespace, interfaceNamespace)}, nil
	default:
		return nil, fmt.Errorf("unsupported isolation type %s", kind)
	}
}

func (i *GenericIsolation) LinkConstrain(names []string, prefix string, group int) error {
	currentLinks, err := i.LinkList(prefix, group)
	if err != nil {
		return err
	}
	for _, link := range currentLinks {
		var unneeded = true
		for _, otherLink := range names {
			if link == otherLink {
				unneeded = false
			}
		}
		if unneeded {
			err = i.LinkAbsent(link)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
