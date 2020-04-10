package types

import (
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
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
	*netlink.Addr
}

func (a *Addr) UnmarshalText(text []byte) error {
	var err error
	a.Addr, err = netlink.ParseAddr(string(text))
	return err
}
