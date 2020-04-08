package rait

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"gitlab.com/NickCao/RAIT/rait/internal/types"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
)

// Peer represents a peer, which Client connects to
type Peer struct {
	PublicKey types.Key // The public key of the peer
	Endpoint  net.IP    // The ip address of the peer, support for domain name is deliberately removed to avoid choosing between multiple address
	SendPort  int       // The listen port of client for this peer
}

// PeerFromFile load peer configuration from a toml file
func PeerFromFile(filePath string) (*Peer, error) {
	var peer = Peer{
		Endpoint: net.ParseIP("127.0.0.1"), // If no endpoint is set, sending packet through this interface would cause failures
	}
	var err error
	_, err = toml.DecodeFile(filePath, &peer)
	if err != nil {
		return nil, fmt.Errorf("failed to decode peer config at %v: %w", filePath, err)
	}
	return &peer, nil
}

// PeersFromDirectory load peer configurations from a directory, in which are toml files with .conf suffix
func PeersFromDirectory(dirPath string) ([]*Peer, error) {
	var fileList []os.FileInfo
	var err error
	fileList, err = ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory content at %v: %w", dirPath, err)
	}
	var peers []*Peer
	for _, fileInfo := range fileList {
		if fileInfo.IsDir() || !strings.HasSuffix(fileInfo.Name(), ".conf") {
			continue
		}
		var peer *Peer
		var err error
		peer, err = PeerFromFile(path.Join(dirPath, fileInfo.Name()))
		if err != nil {
			return nil, err
		}
		peers = append(peers, peer)
	}
	return peers, nil
}
