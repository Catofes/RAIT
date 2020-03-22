package rait

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
)

type Peer struct {
	PublicKey Key
	Endpoint  net.IP
	SendPort  uint16
}

func PeerFromFile(filePath string) (*Peer, error) {
	var p = Peer{
		Endpoint: net.ParseIP("127.0.0.1"),
	}
	var err error
	_, err = toml.DecodeFile(filePath, &p)
	if err != nil {
		return nil, fmt.Errorf("PeerFromFile: failed to decode peer config at %v: %w", filePath, err)
	}
	return &p, nil
}

func PeersFromDirectory(dirPath string) ([]*Peer, error) {
	var ps []*Peer
	var fileList []os.FileInfo
	var err error
	fileList, err = ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("PeersFromDirectory: failed to list directory at %v: %w", dirPath, err)
	}
	for _, fileInfo := range fileList {
		var p *Peer
		var err error
		if fileInfo.IsDir() || !strings.HasSuffix(fileInfo.Name(), ".conf") {
			continue
		}
		p, err = PeerFromFile(path.Join(dirPath, fileInfo.Name()))
		if err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}
	return ps, nil
}
