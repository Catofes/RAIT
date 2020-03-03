package rait

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"path"
	"strconv"
)

type RAIT struct {
	IFPrefix   string
	PrivateKey *wgtypes.Key
	PublicKey  *wgtypes.Key
	SendPort   int
	Tags       map[string]string
	TagPolicy  string
	Peers      []*Peer
}

func NewRAIT(ifPrefix string, privateKey string, port int, tags map[string]string, policy string, peers []*Peer) (*RAIT, error) {
	var key wgtypes.Key
	var err error
	key, err = wgtypes.ParseKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	var publicKey wgtypes.Key
	publicKey = key.PublicKey()
	return &RAIT{
		IFPrefix:   ifPrefix,
		PrivateKey: &key,
		PublicKey:  &publicKey,
		SendPort:   port,
		Tags:       tags,
		TagPolicy:  policy,
		Peers:      peers,
	}, nil
}

func (r *RAIT) EvaluatePolicy(peer *Peer) bool {
	//TODO: Decide the policy format
	return true
}

func (r *RAIT) GetConfigs() []*wgtypes.Config {
	var configs []*wgtypes.Config
	for _, peer := range r.Peers {
		if *r.PublicKey == peer.PublicKey {
			continue
		}
		if !r.EvaluatePolicy(peer) {
			continue
		}
		var endpoint *net.UDPAddr
		if peer.EndpointIP != nil {
			endpoint = &net.UDPAddr{
				IP:   *peer.EndpointIP,
				Port: r.SendPort,
			}
		}
		peerConfig := wgtypes.PeerConfig{
			PublicKey:         peer.PublicKey,
			Remove:            false,
			UpdateOnly:        false,
			PresharedKey:      nil,
			Endpoint:          endpoint,
			ReplaceAllowedIPs: true,
			AllowedIPs:        []net.IPNet{*IP4NetAll, *IP6NetAll},
		}
		config := &wgtypes.Config{
			PrivateKey:   r.PrivateKey,
			ListenPort:   &peer.SendPort,
			FirewallMark: nil,
			ReplacePeers: true,
			Peers:        []wgtypes.PeerConfig{peerConfig},
		}
		configs = append(configs, config)
	}
	return configs
}

func (r *RAIT) SetupLinks() error {
	handle, err := netlink.NewHandle()
	if err != nil {
		return fmt.Errorf("failed to get netlink handle: %w", err)
	}
	defer handle.Delete()
	client, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("failed to get wireguard client: %w", err)
	}
	defer client.Close()
	configs := r.GetConfigs()

	var counter = 0
	for _, config := range configs {
		ifname := r.IFPrefix + strconv.Itoa(counter)
		attrs := netlink.NewLinkAttrs()
		attrs.Name = ifname
		link := &netlink.Wireguard{attrs}
		err := handle.LinkAdd(link)
		if err != nil {
			return fmt.Errorf("failed to add link: %w", err)
		}
		err = handle.LinkSetUp(link)
		if err != nil {
			return fmt.Errorf("failed to bring up link: %w", err)
		}
		err = handle.AddrAdd(link, RandomLinklocal())
		if err != nil {
			return fmt.Errorf("failed to add linklocal address: %w", err)
		}
		err = client.ConfigureDevice(ifname, *config)
		if err != nil {
			return fmt.Errorf("failed to configure link: %w", err)
		}
		counter++
	}
	return nil
}

type RAITFile struct {
	IFPrefix   string
	PrivateKey string
	SendPort   int
	Tags       map[string]string
	TagPolicy  string
	PeerDir    string
}

func NewRAITFromFile(file string) (*RAIT, error) {
	var raitFile RAITFile
	var err error
	_, err = toml.DecodeFile(file, &raitFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode conf file: %v: %w", file, err)
	}
	var peers []*Peer
	peers, err = NewPeersFromDir(path.Join(path.Dir(file), raitFile.PeerDir))
	if err != nil {
		return nil, fmt.Errorf("failed to load peers: %w", err)
	}
	return NewRAIT(raitFile.IFPrefix, raitFile.PrivateKey, raitFile.SendPort, raitFile.Tags, raitFile.TagPolicy, peers)
}
