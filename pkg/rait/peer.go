package rait

import (
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
	"gitlab.com/NickCao/RAIT/v2/pkg/types"
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

func PeerFromMap(data map[string]string) (*Peer, error) {
	var peer Peer
	var err error
	peer.PublicKey, err = wgtypes.ParseKey(data["PublicKey"])
	if err != nil {
		return nil, NewErrDecode("Peer", "PublicKey", err)
	}
	peer.AddressFamily, err = types.ParseAddressFamily(misc.OrDefault(data["AddressFamily"], "ip4"))
	if err != nil {
		return nil, NewErrDecode("Peer", "AddressFamily", err)
	}
	peer.Endpoint, err = types.ParseEndpoint(data["Endpoint"], peer.AddressFamily)
	if err != nil {
		return nil, NewErrDecode("Peer", "Endpoint", err)
	}
	peer.SendPort, err = types.ParseUint16(data["SendPort"])
	if err != nil {
		return nil, NewErrDecode("Peer", "SendPort", err)
	}
	return &peer, nil
}

func PeersFromPath(path string) ([]*Peer, error) {
	var peersMap struct {
		Peers []map[string]string
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
