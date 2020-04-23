package types

import (
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"strings"
)

// Key is a wrapper around wgtypes.Key, and implements encoding.TextUnmarshaler
type Key struct {
	wgtypes.Key
}

func (k *Key) UnmarshalText(text []byte) error {
	var err error
	k.Key, err = wgtypes.ParseKey(string(text))
	return err
}

// Addr is a wrapper around *netlink.Addr, and implements encoding.TextUnmarshaler
type Addr struct {
	net.IP
}

func (a *Addr) UnmarshalText(text []byte) error {
	a.IP = net.ParseIP(string(text))
	if a.IP == nil {
		a.IP = ResolveAddr(string(text))
	}
	return nil
}

func ResolveAddr(addr string) net.IP {
	var chunks []string
	chunks = strings.Split(addr, ":")
	if len(chunks) != 2 {
		return nil
	}
	var ips []net.IP
	var err error
	ips, err = net.LookupIP(chunks[1])
	if err != nil {
		return nil
	}
	switch chunks[0] {
	case "ip4":
		for _, ip := range ips {
			if ip.To4() != nil {
				return ip
			}
		}
	case "ip6":
		for _, ip := range ips {
			if ip.To4() == nil {
				return ip
			}
		}
	default:
		return nil
	}
	return nil
}
