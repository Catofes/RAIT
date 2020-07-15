package rait

import (
	"github.com/go-playground/validator/v10"
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
	"go.uber.org/zap"
)

// Peer represents a single rait node
// which corresponds to a wireguard interface
type Peer struct {
	PublicKey     string `validate:"required,base64"`               // required, the public key of the peer
	AddressFamily string `validate:"required,oneof=ip4 ip6"`        // required, [ip4]/ip6, the address family of this node
	Endpoint      string `validate:"omitempty,ip|hostname_rfc1123"` // the endpoint ip address or resolvable hostname
	SendPort      int    `validate:"required,min=1,max=65535"`      // required, the sending port of the peer
}

func PeersFromPath(path string) ([]*Peer, error) {
	var peers struct {
		Peers []*Peer
	}
	if err := misc.DecodeTOMLFromPath(path, &peers); err != nil {
		return nil, err
	}

	validate := validator.New()
	n := 0
	for _, x := range peers.Peers {
		if x.AddressFamily == "" {
			x.AddressFamily = "ip4"
		}
		if validate.Struct(x) == nil {
			peers.Peers[n] = x
			n++
		} else {
			zap.S().Debugf("peer with public key %s is invalid, discarding", x.PublicKey)
		}
	}
	peers.Peers = peers.Peers[:n]
	return peers.Peers, nil
}
