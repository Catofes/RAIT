package rait

import (
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
)

// RAIT is the model corresponding to rait.conf, for default value of fields, see NewRAIT
type RAIT struct {
	PrivateKey string      `hcl:"private_key,attr"` // mandatory, wireguard private key, base64 encoded
	Peers      string      `hcl:"peers,attr"`       // mandatory, list of peers, in hcl format
	Transport  []Transport `hcl:"transport,block"`  // mandatory, underlying transport for wireguard sockets
	Isolation  *Isolation  `hcl:"isolation,block"`  // optional, params for the separation of underlay and overlay
	Babeld     *Babeld     `hcl:"babeld,block"`     // optional, integration with babeld
}

type Transport struct {
	AddressFamily string `hcl:"address_family,attr"`  // mandatory, socket address family, ip4 or ip6
	SendPort      int    `hcl:"send_port,attr"`       // mandatory, socket send port
	MTU           int    `hcl:"mtu,attr"`             // mandatory, interface mtu
	IFPrefix      string `hcl:"ifprefix,attr"`        // mandatory, interface naming prefix, should not collide between transports
	BindAddress   string `hcl:"bind_addr,optional"`   // optional, socket bind address, only has effect when -b is set
	FwMark        int    `hcl:"fwmark,optional"`      // optional, fwmark set on out going packets
	RandomPort    bool   `hcl:"random_port,optional"` // optional, whether to randomize listen port
}

type Isolation struct {
	IFGroup int    `hcl:"ifgroup,attr"`     // mandatory, interface group, for recognizing managed interfaces
	Transit string `hcl:"transit,optional"` // optional, the namespace to create sockets in
	Target  string `hcl:"target,optional"`  // optional, the namespace to move interfaces into
}

type Babeld struct {
	SocketType string `hcl:"socket_type,optional"` // optional, control socket type, tcp or unix
	SocketAddr string `hcl:"socket_addr,optional"` // optional, control socket address
	Param      string `hcl:"param,optional"`       // optional, interfaces params
	ExtraCmd   string `hcl:"extra_cmd,optional"`   // optional, additional command passed to socket at the end of sync
}

func NewRAIT(path string) (*RAIT, error) {
	var r = &RAIT{
		Peers: "/etc/rait/peers.conf",
		Isolation: &Isolation{
			IFGroup: 54,
		},
		Babeld: &Babeld{
			SocketType: "unix",
			SocketAddr: "/run/babeld.ctl",
			Param:      "type tunnel link-quality true split-horizon false rxcost 32 hello-interval 20 max-rtt-penalty 1024 rtt-max 1024",
		},
	}
	if err := misc.UnmarshalHCL(path, r); err != nil {
		return nil, err
	}
	return r, nil
}
