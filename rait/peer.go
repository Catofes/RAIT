package rait

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"gitlab.com/NickCao/RAIT/rait/internal/types"
	"gitlab.com/NickCao/RAIT/rait/internal/utils"
	"net"
)

// Peer represents a peer, which Client connects to
type Peer struct {
	PublicKey types.Key // The public key of the peer
	Endpoint  net.IP    // The ip address of the peer, support for domain name is deliberately removed to avoid choosing between multiple address
	SendPort  int       // The listen port of client for this peer
}

// PeerList represents a list of peers, to workaround the absence of top level array in toml
type PeerList struct {
	Peers []*Peer
}

// PeersFromURL load peer configurations from the given URL
func PeersFromURL(url string) ([]*Peer, error) {
	var data []byte
	var err error
	data, err = utils.FileFromURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get peers from url: %w", err)
	}
	var peerList PeerList
	_, err = toml.Decode(string(data), &peerList)
	if err != nil {
		return nil, err
	}
	for _, peer := range peerList.Peers {
		if peer.Endpoint == nil {
			peer.Endpoint = net.ParseIP("127.0.0.1") // Dirty but working
		}
	}
	return peerList.Peers, nil
}
