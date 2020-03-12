package rait

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type RAIT struct {
	PrivateKey wgtypes.Key
	PublicKey  wgtypes.Key
	SendPort   int
	Interface  string
	Addresses  []*netlink.Addr
	Namespace  string
}

func NewRAIT(config *RAITConfig) (*RAIT, error) {
	var r RAIT
	var err error
	r.PrivateKey, err = wgtypes.ParseKey(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private privatekey: %w", err)
	}
	r.PublicKey = r.PrivateKey.PublicKey()
	for _, RawAddress := range config.Addresses {
		var ip *netlink.Addr
		var err error
		ip, err = netlink.ParseAddr(RawAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to parse address: %w", err)
		}
		r.Addresses = append(r.Addresses, ip)
	}
	r.SendPort = config.SendPort
	r.Interface = config.Interface
	r.Namespace = config.Namespace
	return &r, nil
}
