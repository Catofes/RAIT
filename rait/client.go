package rait

import (
	"fmt"
	"gitlab.com/NickCao/RAIT/rait/internal/types"
	"gitlab.com/NickCao/RAIT/rait/internal/utils"
)

// Client represents a client
type Client struct {
	PrivateKey         types.Key // mandatory, default nil, the private key of the client
	AddressFamily      types.AF  // optional, default ip4, the address family of the client, ip4 or ip6
	SendPort           int       // mandatory, default 0, the sending port of the client
	InterfacePrefix    string    // optional, default rait, the common prefix to name the wireguard interfaces
	InterfaceNamespace string    // optional, default empty, the netns to move wireguard interface into
	TransitNamespace   string    // optional, default empty, the netns to create wireguard sockets in
	MTU                int       // optional, default 1400, the MTU of the wireguard interfaces
	FwMark             *int      // optional, default nil, the fwmark on packets sent by wireguard
	Peers              string    // optional, default /etc/rait/peers.conf, the url of the peer list
}

func LoadClientFromPath(path string) (*Client, error) {
	var client = &Client{
		AddressFamily:   types.AF_INET,
		InterfacePrefix: "rait",
		MTU:             1400,
		Peers:           "/etc/rait/peers.conf",
	}
	var err error
	err = utils.DecodeTOMLFromPath(path, client)
	if err != nil {
		return nil, fmt.Errorf("failed to load client from path: %w", err)
	}
	return client, err
}
