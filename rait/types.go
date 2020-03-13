package rait

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
)

// TODO: wait for them to satisfy the TextUnmarshaler interface
type Key struct {
	wgtypes.Key
}

func (k *Key) UnmarshalText(text []byte) error {
	var err error
	k.Key, err = wgtypes.ParseKey(string(text))
	return err
}

type Addr struct {
	*netlink.Addr
}

func (a *Addr) UnmarshalText(text []byte) error {
	var err error
	a.Addr, err = netlink.ParseAddr(string(text))
	return err
}

type Peer struct {
	PublicKey Key
	Endpoint  net.IP
	SendPort  uint16
}

func PeerFromFile(filePath string) (*Peer, error) {
	var p Peer
	var err error
	_, err = toml.DecodeFile(filePath, &p)
	if err != nil {
		return nil, fmt.Errorf("failed to decode peer config at %v: %w", filePath, err)
	}
	return &p, nil
}

func PeersFromDirectory(dirPath string) ([]*Peer, error) {
	var ps []*Peer
	var fileList []os.FileInfo
	var err error
	fileList, err = ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory at %v: %w", dirPath, err)
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

type RAIT struct {
	PrivateKey Key
	SendPort   uint16

	Interface string
	Addresses []Addr
	Namespace string
	IFPrefix  string
	MTU       uint16
	FwMark    uint16
}

func RAITFromFile(filePath string) (*RAIT, error) {
	var r RAIT
	var err error
	_, err = toml.DecodeFile(filePath, &r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode rait config at %v: %w", filePath, err)
	}
	return &r, nil
}

type NamespaceHelper struct {
	SrcNamespace netns.NsHandle
	DstNamespace netns.NsHandle
	SrcHandle    *netlink.Handle
	DstHandle    *netlink.Handle
}

func (h *NamespaceHelper) Destroy() {
	h.SrcHandle.Delete()
	h.DstHandle.Delete()
	h.SrcNamespace.Close()
	h.DstNamespace.Close()
}

func NamespaceHelperFromName(name string) (*NamespaceHelper, error) {
	_ = CreateNamedNamespace(name) // It won't hurt
	var h NamespaceHelper
	var err error
	h.SrcNamespace, err = netns.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get src ns: %w", err)
	}
	h.DstNamespace, err = netns.GetFromName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get dst ns: %w", err)
	}
	h.SrcHandle, err = netlink.NewHandle()
	if err != nil {
		return nil, fmt.Errorf("failed to get src handle: %w", err)
	}
	h.DstHandle, err = netlink.NewHandleAt(h.DstNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get dst handle: %w", err)
	}
	return &h, nil
}
