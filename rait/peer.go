package rait

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"io/ioutil"
	"net"
	"path"
)

type Peer struct {
	PublicKey  wgtypes.Key
	EndpointIP *net.IP
	SendPort   int
	Tags       map[string]string
}

func NewPeer(publicKey string, host string, port int, tags map[string]string) (*Peer, error) {
	var key wgtypes.Key
	var err error
	key, err = wgtypes.ParseKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse publickey: %v: %w", publicKey, err)
	}
	var ip *net.IP
	if host != "" {
		var ips []net.IP
		ips, err = net.LookupIP(host)
		if err != nil || len(ips) == 0 {
			return nil, fmt.Errorf("failed to resolve host: %v: %w", host, err)
		}
		ip = &ips[0]
	}
	return &Peer{
		PublicKey:  key,
		EndpointIP: ip,
		SendPort:   port,
		Tags:       tags,
	}, nil
}

type PeerFile struct {
	PublicKey string
	Endpoint  string
	SendPort  int
	Tags      map[string]string
}

func NewPeerFromFile(file string) (*Peer, error) {
	var peerFile PeerFile
	var err error
	_, err = toml.DecodeFile(file, &peerFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode peer file: %v: %w", file, err)
	}
	return NewPeer(peerFile.PublicKey, peerFile.Endpoint, peerFile.SendPort, peerFile.Tags)
}

func NewPeersFromDir(dir string) ([]*Peer, error) {
	var peers []*Peer
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to list peers dir: %v: %w", dir, err)
	}
	for _, file := range files {
		fullpath := path.Join(dir, file.Name())
		peer, err := NewPeerFromFile(fullpath)
		if err != nil {
			return nil, fmt.Errorf("failed to load peer from file: %v: %w", fullpath, err)
		}
		peers = append(peers, peer)
	}
	return peers, nil
}
