package rait

import (
	"fmt"
	"gitlab.com/NickCao/RAIT/rait/utils"
)

// Peer represents a wireguard peer
type Peer struct {
	PublicKey     string // mandatory, the public key of the peer
	AddressFamily string // the address family of the specified endpoint, ip4 or ip6
	Endpoint      string // the endpoint ip address or resolvable domain name
	SendPort      int    // mandatory, the sending port of the peer
}

// PeerList contains a list of peers (to workaround the lack of top level array in toml)
type PeerList struct {
	Peers []*Peer
}

// SetDefaults sets sane defaults for fields that are not present when decoding toml
func (pl *PeerList) SetDefaults() {
	for _, p := range pl.Peers {
		if p.AddressFamily == "" {
			p.AddressFamily = "ip4"
		}
		if p.Endpoint == "" {
			switch p.AddressFamily {
			case "ip4":
				p.Endpoint = "127.0.0.1"
			case "ip6":
				p.Endpoint = "::1"
			}
		}
	}
}

// LoadPeersFromPath returns PeerList loaded from given path
func LoadPeersFromPath(path string) (PeerList, error) {
	var pl PeerList
	err := utils.DecodeTOMLFromPath(path, &pl)
	if err != nil {
		return PeerList{}, fmt.Errorf("LoadPeersFromPath: failed to load peers: %w", err)
	}
	pl.SetDefaults()
	return pl, nil
}
