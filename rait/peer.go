package rait

import (
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
)

type Peer struct {
	PublicKey  wgtypes.Key
	EndpointIP net.IP
	SendPort   int
}

func NewPeer(config *PeerConfig) (*Peer, error) {
	var p Peer
	var err error
	p.PublicKey, err = wgtypes.ParseKey(config.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse peer publickey: %w", err)
	}
	p.EndpointIP = net.ParseIP(config.Endpoint)
	if p.EndpointIP == nil{
		p.EndpointIP = net.ParseIP("127.0.0.1")
	}
	p.SendPort = config.SendPort
	return &p, nil
}
