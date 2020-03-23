package rait

import (
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// TODO: wait for them to satisfy the TextUnmarshaler interface
type Key struct {
	wgtypes.Key
}

func (k *Key) UnmarshalText(text []byte) error {
	var err error
	k.Key, err = wgtypes.ParseKey(string(text))
	return err
}

type Addr struct {
	*netlink.Addr
}

func (a *Addr) UnmarshalText(text []byte) error {
	var err error
	a.Addr, err = netlink.ParseAddr(string(text))
	return err
}
