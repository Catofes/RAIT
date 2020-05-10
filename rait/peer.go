package rait

import (
	"fmt"
	"gitlab.com/NickCao/RAIT/rait/internal/types"
	"gitlab.com/NickCao/RAIT/rait/internal/utils"
)

// Peer represents a peer
type Peer struct {
	PublicKey     types.Key // mandatory, the public key of the peer
	AddressFamily types.AF  // optional, the address family of the specified endpoint, ip4 or ip6
	Endpoint      string    // optional, the endpoint ip address or resolvable domain name
	SendPort      int       // mandatory, the sending port of the peer
}

func LoadPeersFromPath(path string) (peers []*Peer, err error) {
	var plist = struct {
		Peers []*Peer
	}{}
	err = utils.DecodeTOMLFromPath(path, &plist)
	peers = plist.Peers
	if err != nil {
		err = fmt.Errorf("LoadPeersFromPath: failed to load peers from path: %w", err)
	}
	return
}
