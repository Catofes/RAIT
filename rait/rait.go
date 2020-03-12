package rait

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type RAIT struct {
	PrivateKey wgtypes.Key
	PublicKey  wgtypes.Key
	SendPort   int
	IFPrefix   string
	DummyName  string
	DummyIP    []*netlink.Addr
	NetNS      string
	// Fields bellow are initialized at runtime
	OriginalNSHandle       netns.NsHandle
	SpecifiedNSHandle      netns.NsHandle
	OriginalNetlinkHandle  *netlink.Handle
	SpecifiedNetlinkHandle *netlink.Handle
}

func NewRAIT(config *RAITConfig) (*RAIT, error) {
	var r RAIT
	var err error
	r.PrivateKey, err = wgtypes.ParseKey(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private privatekey: %w", err)
	}
	r.PublicKey = r.PrivateKey.PublicKey()

	for _, RawDummyIP := range config.DummyIP {
		var ip *netlink.Addr
		var err error
		ip, err = netlink.ParseAddr(RawDummyIP)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ip: %w", err)
		}
		r.DummyIP = append(r.DummyIP, ip)
	}

	r.SendPort = config.SendPort
	r.IFPrefix = config.IFPrefix
	r.DummyName = config.DummyName
	r.NetNS = config.NetNS
	return &r, nil
}
