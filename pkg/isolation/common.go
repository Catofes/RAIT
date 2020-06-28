package isolation

import (
	"fmt"
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var isolationRegistry = make(map[string]func(string, string) (Isolation, error))

// Register registers a isolation type into a internal registry to be used by NewGenericIsolation
func Register(name string, fn func(string, string) (Isolation, error)) {
	isolationRegistry[name] = fn
}

// Isolation represents a management interface for wireguard links
// together with the isolation technique employed to isolate overlay from underlay
type Isolation interface {
	// LinkEnsure ensures the existence and state of the given link is as expected
	// this method should be idempotent as it's also used to sync the state of links
	LinkEnsure(ifname string, config wgtypes.Config, mtu, ifgroup int) error
	// LinkAbsent ensures the absence of the given link
	LinkAbsent(ifname string) error
	// LinkFilter returns the wireguard links within the constrains as seen by the isolation
	LinkFilter(prefix string, ifgroup int) ([]string, error)
}

// GenericIsolation is a wrapper around concrete implementations and provides a higher level api over them
type GenericIsolation struct {
	Isolation
}

// NewGenericIsolation provides a unified constructor for concrete implementations
// current supported isolation types are netns and vrf
func NewGenericIsolation(kind, transitScope, interfaceScope string) (*GenericIsolation, error) {
	if isoFn, ok := isolationRegistry[kind]; ok {
		iso, err := isoFn(transitScope, interfaceScope)
		if err != nil {
			return nil, err
		}
		return &GenericIsolation{iso}, nil
	}
	return nil, fmt.Errorf("unsupported isolation type %s", kind)
}

// LinkSync accepts a list of link as the desired state, and removes the extraneous links
func (i *GenericIsolation) LinkConstrain(names []string, prefix string, ifgroup int) error {
	linkList, err := i.LinkFilter(prefix, ifgroup)
	if err != nil {
		return err
	}

	for _, link := range linkList {
		if !misc.In(names, link) {
			err = i.LinkAbsent(link)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
