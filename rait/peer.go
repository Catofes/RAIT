package rait

import (
	"fmt"
	"gitlab.com/NickCao/RAIT/rait/internal/types"
	"gitlab.com/NickCao/RAIT/rait/internal/utils"
)

// Peer represents a peer
type Peer struct {
	PublicKey     types.Key // mandatory, default nil, the public key of the peer
	AddressFamily types.AF  // optional, default ip4, the address family of the specified endpoint, ip4 or ip6
	Endpoint      string    // optional, default localhost, the endpoint ip address or resolvable domain name
	SendPort      int       // mandatory, default 0, the sending port of the peer
}

// PeerList represents a list of peer, to work around the lack of top level array in toml
type PeerList struct {
	Peers []*Peer
}

func LoadPeersFromPath(path string) ([]*Peer, error) {
	var peers PeerList
	var err error
	err = utils.DecodeTOMLFromPath(path, &peers)
	if err != nil {
		return nil, fmt.Errorf("failed to load peers from path: %w", err)
	}
	for _, peer := range peers.Peers {
		if peer.AddressFamily == types.AF_UNSPEC {
			peer.AddressFamily = types.AF_INET
		}
		if peer.Endpoint == "" {
			switch peer.AddressFamily {
			case types.AF_INET:
				peer.Endpoint = "127.0.0.1"
			case types.AF_INET6:
				peer.Endpoint = "::1"
			}
		}
	}
	return peers.Peers, err
}
