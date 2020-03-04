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

type PeerConfig struct {
	PublicKey string
	Endpoint  string `json:",omitempty"`
	SendPort  int
	Tags      map[string]string `json:",omitempty"`
}

func NewPeer(config *PeerConfig) (*Peer, error) {
	var publickey wgtypes.Key
	var err error
	publickey, err = wgtypes.ParseKey(config.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse peer publickey: %w", err)
	}
	var ip *net.IP
	if config.Endpoint != "" {
		var ips []net.IP
		ips, err = net.LookupIP(config.Endpoint)
		if err != nil || len(ips) < 1 {
			return nil, fmt.Errorf("failed to resolve peer endpoint: %w", err)
		}
		ip = &ips[0]
	}
	return &Peer{
		PublicKey:  publickey,
		EndpointIP: ip,
		SendPort:   config.SendPort,
		Tags:       config.Tags,
	}, nil
}

func NewPeerFromToml(filepath string) (*Peer, error) {
	var config PeerConfig
	var err error
	_, err = toml.DecodeFile(filepath, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode peer config: %w", err)
	}
	return NewPeer(&config)
}

func LoadPeersFromTomls(dirpath string) ([]*Peer, error) {
	var peers []*Peer
	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return nil, fmt.Errorf("failed to list peer configs: %w", err)
	}
	for _, file := range files {
		filepath := path.Join(dirpath, file.Name())
		peer, err := NewPeerFromToml(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to load peer from toml file: %w", err)
		}
		peers = append(peers, peer)
	}
	return peers, nil
}
