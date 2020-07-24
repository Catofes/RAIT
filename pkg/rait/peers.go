package rait

import "gitlab.com/NickCao/RAIT/v3/pkg/misc"

type Peers struct {
	Peers []Peer `hcl:"peers,block"`
}

type Peer struct {
	PublicKey     string `hcl:"public_key,attr"`     // mandatory, wireguard public key, base64 encoded
	AddressFamily string `hcl:"address_family,attr"` // mandatory, socket address family, ip4 or ip6
	SendPort      int    `hcl:"send_port,attr"`      // mandatory, socket send port
	Endpoint      string `hcl:"endpoint,optional"`   // optional, ip address or resolvable domain name
	Name          string `hcl:"name,optional"`       // optional, peer human readable name
}

func NewPeers(path string, pubkey string) ([]Peer, error) {
	var peersTmp = &Peers{}
	if err := misc.UnmarshalHCL(path, peersTmp); err != nil {
		return nil, err
	}
	peers := peersTmp.Peers

	// in place filter to remove self from peers
	n := 0
	for _, peer := range peers {
		if peer.PublicKey != pubkey {
			peers[n] = peer
			n++
		}
	}
	return peers[:n], nil
}