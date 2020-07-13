package isolation

import (
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var isolationRegistry = make(map[string]func(string, string) (Isolation, error))

// Register registers a isolation type into a internal registry to be used by NewGenericIsolation
func Register(name string, fn func(string, string) (Isolation, error)) {
	isolationRegistry[name] = fn
}

// LinkAttrs represents a single link managed by isolation
type LinkAttrs struct {
	MTU   int
	Name  string
	Group int
}

// Isolation represents a management interface for wireguard links
// together with the isolation technique employed to isolate overlay from underlay
type Isolation interface {
	// LinkEnsure ensures the existence and state of the given link is as expected
	// this method should be idempotent as it's also used to sync the state of links
	LinkEnsure(attrs *LinkAttrs, config wgtypes.Config) error
	// LinkAbsent ensures the absence of the given link
	LinkAbsent(attrs *LinkAttrs) error
	// LinkList returns the wireguard links as seen by the isolation
	LinkList() ([]*LinkAttrs, error)
}

// NewIsolation provides a unified constructor for concrete implementations
// current supported isolation types are netns and vrf
func NewIsolation(kind, transitScope, interfaceScope string) (Isolation, error) {
	if isoFn, ok := isolationRegistry[kind]; ok {
		iso, err := isoFn(transitScope, interfaceScope)
		if err != nil {
			return nil, err
		}
		return iso, nil
	}
	return nil, fmt.Errorf("unsupported isolation type %s", kind)
}

func LinkFilter(links []*LinkAttrs, filterFunc func(*LinkAttrs) bool) (filtered []*LinkAttrs) {
	for _, link := range links {
		if filterFunc(link) {
			filtered = append(filtered, link)
		}
	}
	return
}

func LinkString(links []*LinkAttrs) (stringed []string) {
	for _, link := range links {
		stringed = append(stringed, link.Name)
	}
	return
}

func LinkIn(list []*LinkAttrs, item *LinkAttrs) bool {
	for _, v := range list {
		if v.Name == item.Name {
			return true
		}
	}
	return false
}
