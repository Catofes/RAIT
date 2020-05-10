package rait

import (
	"fmt"
	"gitlab.com/NickCao/RAIT/rait/internal/types"
	"gitlab.com/NickCao/RAIT/rait/internal/utils"
)

// Client represents a client
type Client struct {
	PrivateKey         types.Key // mandatory, the private key of the client
	AddressFamily      types.AF  // optional, the address family of the client, ip4 or ip6
	SendPort           int       // mandatory, the sending port of the client
	InterfacePrefix    string    // optional, the common prefix to name the wireguard interfaces
	InterfaceGroup     int       // optional, the ifgroup for the wireguard interfaces
	InterfaceNamespace string    // optional, the netns to move wireguard interface into
	TransitNamespace   string    // optional, the netns to create wireguard sockets in
	MTU                int       // optional, the MTU of the wireguard interfaces
	FwMark             int       // optional, the fwmark on packets sent by wireguard
	Peers              string    // optional, the url of the peer list
}

func LoadClientFromPath(path string) (client *Client, err error) {
	client = &Client{
		AddressFamily:   "ip4",
		InterfacePrefix: "rait",
		InterfaceGroup:  0x36,
		MTU:             1400,
		FwMark:          0x36,
		Peers:           "/etc/rait/peers.conf",
	}
	err = utils.DecodeTOMLFromPath(path, client)
	if err != nil {
		err = fmt.Errorf("LoadClientFromPath: failed to load client from path: %w", err)
	}
	return
}
