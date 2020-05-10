package types

import (
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
)

// Key is a wrapper around wgtypes.Key, and implements encoding.TextUnmarshaler
type Key wgtypes.Key

func (k *Key) UnmarshalText(text []byte) (err error) {
	var _k wgtypes.Key
	_k, err = wgtypes.ParseKey(string(text))
	*k = Key(_k)
	return
}

type AF string

func (a AF) Equal(b AF) bool {
	return a == b
}

func (a AF) ResolveIP(name string) (addr *net.IPAddr, err error) {
	switch a {
	case "ip4":
		if name == "" {
			name = "127.0.0.1"
		}
		addr, err = net.ResolveIPAddr("ip4", name)
	case "ip6":
		if name == "" {
			name = "::1"
		}
		addr, err = net.ResolveIPAddr("ip6", name)
	default:
		err = fmt.Errorf("ResolveIP: unsupported address family: %s", a)
	}
	return
}
