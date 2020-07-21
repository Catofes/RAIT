package rait

import (
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
)

type RAIT struct {
	PrivateKey string      `hcl:"private_key,attr"` // wireguard private key
	Transport  []Transport `hcl:"transport,block"`  // underlying transport for wireguard socket
	Isolation  *Isolation  `hcl:"isolation,block"`  // separation of underlay and overlay
	Babeld     *Babeld     `hcl:"babeld,block"`     // integration with babeld
	Peers      string      `hcl:"peers,attr"`       // list of peers
}

type Transport struct {
	AddressFamily     string `hcl:"address_family,attr"`
	SendPort          int    `hcl:"send_port,attr"`
	MTU               int    `hcl:"mtu,attr"`
	IFPrefix          string `hcl:"ifprefix,attr"`
	BindAddress       string `hcl:"bind_addr,optional"`
	FwMark            int    `hcl:"fwmark,optional"`
	DynamicListenPort bool   `hcl:"dynamic_port,optional"`
}

type Isolation struct {
	Type    string `hcl:"type,attr"`
	IFGroup int    `hcl:"ifgroup,attr"`
	Transit string `hcl:"transit,optional"`
	Target  string `hcl:"target,optional"`
}

type Babeld struct {
	SocketType string `hcl:"socket_type,optional"`
	SocketAddr string `hcl:"socket_addr,optional"`
	Param      string `hcl:"param,optional"`
	ExtraCmd   string `hcl:"extra_cmd,optional"`
}

func NewRAIT(path string) (*RAIT, error) {
	var r = RAIT{
		Isolation: &Isolation{
			Type:    "netns",
			IFGroup: 54,
		},
		Babeld: &Babeld{
			SocketType: "unix",
			SocketAddr: "/run/babeld.ctl",
			Param:      "type tunnel link-quality true split-horizon false rxcost 32 hello-interval 20 max-rtt-penalty 1024 rtt-max 1024",
		},
		Peers: "/etc/rait/peers.conf",
	}
	if err := misc.UnmarshalHCL(path, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

type Peer struct {
	PublicKey     string `hcl:"public_key,attr"`
	AddressFamily string `hcl:"address_family,attr"`
	SendPort      int    `hcl:"send_port,attr"`
	Endpoint      string `hcl:"endpoint,optional"`
}

func NewPeers(path string, pubkey string) ([]Peer, error) {
	var peers struct {
		Peers []Peer `hcl:"peers,block"`
	}
	if err := misc.UnmarshalHCL(path, &peers); err != nil {
		return nil, err
	}

	n := 0
	for _, x := range peers.Peers {
		if x.PublicKey != pubkey {
			peers.Peers[n] = x
			n++
		}
	}
	return peers.Peers[:n], nil
}
