package rait

import (
	"net"

	"github.com/Catofes/RAIT/pkg/misc"
	"github.com/hashicorp/hcl/v2"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Peers struct {
	Peers []Peer `hcl:"peers,block"`
}

type Peer struct {
	PublicKey string   `hcl:"public_key,attr"` // mandatory, wireguard public key, base64 encoded
	Name      string   `hcl:"name,optional"`   // optional, peer human readable name
	Remarks   hcl.Body `hcl:"remarks,remain"`  // optional, additional information
	Endpoint  Endpoint `hcl:"endpoint,block"`  // mandatory, node endpoints
}

func (s *Peer) GenerateMac() net.HardwareAddr {
	if s.Endpoint.Mac != "" {
		if mac, err := net.ParseMAC(s.Endpoint.Mac); err == nil {
			return mac
		}
	}
	mac := misc.NewMacFromKey(s.PublicKey + s.Endpoint.AddressFamily)
	s.Endpoint.Mac = mac.String()
	zap.S().Debugf("peer mac: %s from %s", mac, s.PublicKey+s.Endpoint.AddressFamily)
	return mac
}

func (s *Peer) GenerateInnerAddress() net.IP {
	if s.Endpoint.InnerAddress == "" {
		s.Endpoint.InnerAddress = misc.NewLLAddrFromKey(s.PublicKey + s.Endpoint.AddressFamily + "wireguard").String()
	}
	return net.ParseIP(s.Endpoint.InnerAddress)
}

type Endpoint struct {
	AddressFamily string `hcl:"address_family,attr"`    // mandatory, socket address family, ip4 or ip6
	Mac           string `hcl:"mac,optional"`           // optional, mac address
	Port          int    `hcl:"port,attr"`              // mandatory, socket listen port
	InnerAddress  string `hcl:"inner_address,optional"` // optional, remote inner address
	Address       string `hcl:"address,optional"`       // optional, ip address or resolvable domain name
}

func NewPeers(path string, privateKeys []wgtypes.Key) ([]Peer, error) {
	var peersTmp = &Peers{}
	if err := misc.UnmarshalHCL(path, peersTmp); err != nil {
		return nil, err
	}
	peers := peersTmp.Peers

	// in place filter to remove self from peers
	n := 0
	for _, peer := range peers {
		for _, privateKey := range privateKeys {
			if peer.PublicKey != privateKey.PublicKey().String() {
				peer.GenerateMac()
				peers[n] = peer
				n++
				break
			}
		}
	}
	return peers[:n], nil
}
