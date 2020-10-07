package rait

import (
	"github.com/Catofes/RAIT/pkg/misc"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// RAIT is the model corresponding to rait.conf, for default value of fields, see NewRAIT
type RAIT struct {
	Name      string      `hcl:"name,optional"`   // optional, human readable node name
	Peers     string      `hcl:"peers,attr"`      // mandatory, list of peers, in hcl format
	Transport []Transport `hcl:"transport,block"` // mandatory, underlying transport for wireguard sockets
	Isolation *Isolation  `hcl:"isolation,block"` // optional, params for the separation of underlay and overlay
	Babeld    *Babeld     `hcl:"babeld,block"`    // optional, integration with babeld
	Remarks   hcl.Body    `hcl:"remarks,remain"`  // optional, additional information
}

type Transport struct {
	PrivateKey    string `hcl:"private_key,attr"`       // mandatory, wireguard private key, base64 encoded
	AddressFamily string `hcl:"address_family,attr"`    // mandatory, socket address family, ip4 or ip6
	Port          int    `hcl:"port,attr"`              // mandatory, socket listen port
	MTU           int    `hcl:"mtu,attr"`               // mandatory, interface mtu
	IFPrefix      string `hcl:"ifprefix,attr"`          // mandatory, interface naming prefix, should not collide between transports
	InnerAddress  string `hcl:"inner_address,optional"` //optional, interface inner ip, should not collide in a network
	Mac           string `hcl:"mac,optional"`
	VNI           int    `hcl:"vni"`
	Address       string `hcl:"address,optional"`     // optional, public ip address or resolvable domain name
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
	SocketType     string   `hcl:"socket_type,optional"`     // optional, control socket type, tcp or unix
	SocketAddr     string   `hcl:"socket_addr,optional"`     // optional, control socket address
	AddonInterface []string `hcl:"addon_interface,optional"` // optional, control socket address
	Param          string   `hcl:"param,optional"`           // optional, interfaces params
	ExtraCmd       string   `hcl:"extra_cmd,optional"`       // optional, additional command passed to socket at the end of sync
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

func (r *RAIT) PublicConf(dest string) error {
	f := hclwrite.NewEmptyFile()
	pubs := Peers{}
	pubs.Peers = make([]Peer, 0)
	for _, t := range r.Transport {
		privKey, err := wgtypes.ParseKey(t.PrivateKey)
		if err != nil {
			return err
		}
		pub := Peer{
			PublicKey: privKey.PublicKey().String(),
			Name:      r.Name,
		}
		pub.Endpoint = Endpoint{
			AddressFamily: t.AddressFamily,
			Address:       t.Address,
			Mac:           t.Mac,
			InnerAddress:  t.InnerAddress,
			Port:          t.Port,
		}
		//pub.GenerateMac()
		pubs.Peers = append(pubs.Peers, pub)
	}
	gohcl.EncodeIntoBody(&pubs, f.Body())
	w, err := misc.NewWriteCloser(dest)
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = w.Write(f.Bytes())
	return err
}
