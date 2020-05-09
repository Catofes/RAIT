package types

import (
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

// Endpoint is a wrapper around string, and implements Resolve
type Endpoint string

func (e *Endpoint) Resolve(af AF) (*net.IPAddr, error) {
	if af == AF_UNSPEC {
		af = AF_INET
	}
	if *e == "" {
		switch af {
		case AF_INET:
			*e = "127.0.0.1"
		case AF_INET6:
			*e = "::1"
		}
	}
	return net.ResolveIPAddr(string(af), string(*e))
}
