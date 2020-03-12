package rait

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"path"
	"strings"
)

type RAITConfig struct {
	PrivateKey string
	SendPort   int
	Interface  string
	Addresses  []string
	Namespace  string
}

func NewRAITFromFile(filepath string) (*RAIT, error) {
	var config RAITConfig
	var err error
	_, err = toml.DecodeFile(filepath, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode rait config: %w", err)
	}
	return NewRAIT(&config)
}

type PeerConfig struct {
	PublicKey string
	Endpoint  string
	SendPort  int
}

func NewPeerFromFile(filepath string) (*Peer, error) {
	var config PeerConfig
	var err error
	_, err = toml.DecodeFile(filepath, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode peer config: %w", err)
	}
	return NewPeer(&config)
}

func NewPeersFromDirectory(dirpath string, suffix string) ([]*Peer, error) {
	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return nil, fmt.Errorf("failed to list peer configs: %w", err)
	}

	var peers []*Peer
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), suffix) || file.IsDir() {
			continue
		}
		filepath := path.Join(dirpath, file.Name())
		peer, err := NewPeerFromFile(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to load peer from file: %w", err)
		}
		peers = append(peers, peer)
	}
	return peers, nil
}
