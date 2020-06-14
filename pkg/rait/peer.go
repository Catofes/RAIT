package rait

import (
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
	"gitlab.com/NickCao/RAIT/v2/pkg/types"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
)

// Peer represents a single wireguard peer
// which corresponds to an interface in RAIT
type Peer struct {
	PublicKey     wgtypes.Key // mandatory, the public key of the peer
	AddressFamily string      // optional, default ip4, the address family of the specified endpoint, ip4 or ip6
	Endpoint      net.IP      // optional, default localhost, the endpoint ip address or resolvable domain name
	SendPort      int         // mandatory, the sending port of the peer
}

func PeerFromMap(data map[string]interface{}) (*Peer, error) {
	logger := zap.S().Named("rait.PeerFromMap")
	var peer Peer
	var err error
	peer.PublicKey, err = wgtypes.ParseKey(misc.Fallback(data["PublicKey"], "").(string))
	if err != nil {
		logger.Errorf("failed to parse wireguard public key, error %s", err)
		return nil, err
	}
	peer.AddressFamily = types.ParseAddressFamily(data["AddressFamily"])
	peer.Endpoint = types.ParseEndpoint(data["Endpoint"], data["AddressFamily"])
	peer.SendPort = types.ParseInt64(data["SendPort"], 0)
	return &peer, nil
}

func PeersFromPath(path string) ([]*Peer, error) {
	var peersMap struct {
		Peers []map[string]interface{}
	}
	err := misc.DecodeTOMLFromPath(path, &peersMap)
	if err != nil {
		return nil, err
	}
	var peers []*Peer
	for _, peerMap := range peersMap.Peers {
		peer, err := PeerFromMap(peerMap)
		if err != nil {
			return nil, err
		}
		peers = append(peers, peer)
	}
	return peers, nil
}
