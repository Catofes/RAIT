package rait

import (
	"encoding/json"
	"fmt"
	"github.com/vishvananda/netlink"
	"math/rand"
	"net"
	"time"
)

var _, IP4NetAll, _ = net.ParseCIDR("0.0.0.0/0")
var _, IP6NetAll, _ = net.ParseCIDR("::/0")

func RandomLinklocal() *netlink.Addr {
	rand.Seed(time.Now().UnixNano())
	digits := []int{0x00, 0x16, 0x3e, rand.Intn(0x7f + 1), rand.Intn(0xff + 1), rand.Intn(0xff + 1)}
	digits = append(digits, 0, 0)
	copy(digits[5:], digits[3:])
	digits[3] = 0xff
	digits[4] = 0xfe
	digits[0] = digits[0] ^ 2
	var parts string
	for i := 0; i < len(digits); i += 2 {
		parts += ":"
		parts += fmt.Sprintf("%x", digits[i])
		parts += fmt.Sprintf("%x", digits[i+1])
	}
	a, _ := netlink.ParseAddr("fe80:" + parts + "/64")
	return a
}

type JSONConfig struct {
	RAIT  RAITConfig
	Peers []PeerConfig
}

func LoadFromJSON(data []byte) (*RAIT, []*Peer, error) {
	var config JSONConfig
	var err error
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode json: %w", err)
	}
	var peers []*Peer
	for _, peerconfig := range config.Peers {
		peer, err := NewPeer(&peerconfig)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load peer: %w", err)
		}
		peers = append(peers, peer)
	}
	var rait *RAIT
	rait, err = NewRAIT(&config.RAIT)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load rait: %w", err)
	}
	return rait, peers, nil
}
