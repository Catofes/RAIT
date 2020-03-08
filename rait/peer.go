package rait

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"io/ioutil"
	"log"
	"net"
	"path"
	"strings"
)

type Peer struct {
	PublicKey  wgtypes.Key
	EndpointIP *net.IP
	SendPort   int
}

type PeerConfig struct {
	PublicKey string
	Endpoint  string `json:",omitempty"`
	SendPort  int
}

func NewPeer(config *PeerConfig) (*Peer, error) {
	var publickey wgtypes.Key
	var ip *net.IP
	var err error
	publickey, err = wgtypes.ParseKey(config.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse peer publickey: %w", err)
	}
	if config.Endpoint != "" {
		tmp := net.ParseIP(config.Endpoint)
		if tmp != nil {
			ip = &tmp
		}
	}
	return &Peer{
		PublicKey:  publickey,
		EndpointIP: ip,
		SendPort:   config.SendPort,
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

func LoadPeersFromTomls(dirpath string, suffix string) ([]*Peer, error) {
	var peers []*Peer
	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return nil, fmt.Errorf("failed to list peer configs: %w", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), suffix) && !file.IsDir() {
			filepath := path.Join(dirpath, file.Name())
			peer, err := NewPeerFromToml(filepath)
			if err != nil {
				log.Println(fmt.Errorf("failed to load peer from toml file: %w", err))
				continue
			}
			peers = append(peers, peer)
		}
	}
	return peers, nil
}
