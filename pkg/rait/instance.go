package rait

import (
	"github.com/go-playground/validator/v10"
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
)

// Instance is at the heart of rait
// it serves as the single source of truth for subsequent configuration of wireguard tunnels
type Instance struct {
	PrivateKey    string `validate:"required,base64"`          // required, the private key of current node
	AddressFamily string `validate:"required,oneof=ip4 ip6"`   // required, [ip4]/ip6, the address family of current node
	SendPort      int    `validate:"required,min=1,max=65535"` // required, the sending (destination) port of wireguard sockets
	BindAddress   string `validate:"omitempty,ip"`             // the local address for wireguard sockets to bind to

	InterfacePrefix string `validate:"required"`                    // [rait], the common prefix to name the wireguard interfaces
	InterfaceGroup  int    `validate:"min=0,max=2147483647"`        // [54], the ifgroup for the wireguard interfaces
	MTU             int    `validate:"required,min=1280,max=65535"` // [1400], the MTU of the wireguard interfaces
	FwMark          int    `validate:"min=0,max=4294967295"`        // [0x36], the fwmark on packets sent by wireguard sockets

	Isolation          string `validate:"required,oneof=netns vrf"` // [netns]/vrf, the isolation method to separate overlay from underlay
	InterfaceNamespace string // the netns or vrf to move wireguard interface into
	TransitNamespace   string // the netns or vrf to create wireguard sockets in
	// The creation of netns is handled automatically, while the creation of vrf must be done manually

	Peers string // [/etc/rait/peers.conf], the url of the peer list
}

func InstanceFromPath(path string) (*Instance, error) {
	var instance = Instance{
		AddressFamily:   "ip4",
		InterfacePrefix: "rait",
		InterfaceGroup:  54,
		MTU:             1400,
		FwMark:          0x36,
		Isolation:       "netns",
		Peers:           "/etc/rait/peers.conf",
	}
	if err := misc.DecodeTOMLFromPath(path, &instance); err != nil {
		return nil, err
	}
	if err := validator.New().Struct(&instance); err != nil {
		return nil, err
	}
	return &instance, nil
}
