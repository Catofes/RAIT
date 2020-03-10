package rait

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

type RAIT struct {
	PrivateKey wgtypes.Key
	PublicKey  wgtypes.Key
	SendPort   int
	IFPrefix   string
	DummyName  string
	DummyIP    []*netlink.Addr
}

type RAITConfig struct {
	PrivateKey string
	SendPort   int
	IFPrefix   string
	DummyName  string
	DummyIP    []string
}

func NewRAIT(config *RAITConfig) (*RAIT, error) {
	var r RAIT
	var err error
	r.PrivateKey, err = wgtypes.ParseKey(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private privatekey: %w", err)
	}
	r.PublicKey = r.PrivateKey.PublicKey()
	r.SendPort = config.SendPort
	r.IFPrefix = config.IFPrefix
	r.DummyName = config.DummyName
	for _, RawDummyIP := range config.DummyIP {
		var ip *netlink.Addr
		var err error
		ip, err = netlink.ParseAddr(RawDummyIP)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ip: %w", err)
		}
		r.DummyIP = append(r.DummyIP, ip)
	}
	return &r, nil
}

func NewRAITFromToml(filepath string) (*RAIT, error) {
	var config RAITConfig
	var err error
	_, err = toml.DecodeFile(filepath, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode rait config: %w", err)
	}
	return NewRAIT(&config)
}

func (r *RAIT) generateWireguardConfigs(peers []*Peer) []*wgtypes.Config {
	var configs []*wgtypes.Config
	for _, peer := range peers {
		if r.PublicKey == peer.PublicKey {
			continue // ignore self quick path
		}
		var endpoint *net.UDPAddr
		if peer.EndpointIP != nil {
			endpoint = &net.UDPAddr{
				IP:   *peer.EndpointIP,
				Port: r.SendPort,
			}
		} else {
			// This is to make babeld happy
			fakeip := net.ParseIP("127.0.0.1")
			endpoint = &net.UDPAddr{
				IP:   fakeip,
				Port: r.SendPort,
			}
		}
		peerconfig := wgtypes.PeerConfig{
			PublicKey:         peer.PublicKey,
			Remove:            false,
			UpdateOnly:        false,
			PresharedKey:      nil,
			Endpoint:          endpoint,
			ReplaceAllowedIPs: true,
			AllowedIPs:        []net.IPNet{*IP4NetAll, *IP6NetAll},
		}
		mark := 54 //Just a randomly generated number
		config := &wgtypes.Config{
			PrivateKey:   &r.PrivateKey,
			ListenPort:   &peer.SendPort,
			FirewallMark: &mark,
			ReplacePeers: true,
			Peers:        []wgtypes.PeerConfig{peerconfig},
		}
		configs = append(configs, config)
	}
	return configs
}

func (r *RAIT) SetupWireguardLinks(peers []*Peer) error {
	var handle *netlink.Handle
	var err error
	handle, err = netlink.NewHandle()
	if err != nil {
		return fmt.Errorf("failed to get netlink handle: %w", err)
	}
	defer handle.Delete()

	var client *wgctrl.Client
	client, err = wgctrl.New()
	if err != nil {
		return fmt.Errorf("failed to get wireguard client: %w", err)
	}
	defer client.Close()

	configs := r.generateWireguardConfigs(peers)
	var counter = 0
	for _, config := range configs {
		ifname := r.IFPrefix + strconv.Itoa(counter)
		// relying on iproute2 before native wireguard support
		err = exec.Command("ip", "link", "add", ifname, "type", "wireguard").Run()
		if err != nil {
			return fmt.Errorf("failed to add link: %w", err)
		}
		var link netlink.Link
		link, err = netlink.LinkByName(ifname)
		if err != nil {
			return fmt.Errorf("failed to retrive link by name: %w", err)
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

func (r *RAIT) DestroyWireguardLinks() error {
	var handle *netlink.Handle
	var err error
	handle, err = netlink.NewHandle()
	if err != nil {
		return fmt.Errorf("failed to get netlink handle: %w", err)
	}
	defer handle.Delete()

	var links []netlink.Link
	links, err = handle.LinkList()
	if err != nil {
		return fmt.Errorf("failed to list links: %w", err)
	}
	for _, link := range links {
		if strings.HasPrefix(link.Attrs().Name, r.IFPrefix) && link.Type() == "wireguard" {
			err = handle.LinkDel(link)
			if err != nil {
				return fmt.Errorf("failed to delete link: %w", err)
			}
		}
	}
	return nil
}

func (r *RAIT) SetupDummyInterface() error {
	var handle *netlink.Handle
	var err error
	handle, err = netlink.NewHandle()
	if err != nil {
		return fmt.Errorf("failed to get netlink handle: %w", err)
	}
	defer handle.Delete()

	var link = &netlink.Dummy{
		LinkAttrs: netlink.LinkAttrs{
			Name: r.DummyName,
		},
	}
	err = handle.LinkAdd(link)
	if err != nil {
		return fmt.Errorf("failed to add dummy interface: %w", err)
	}
	err = handle.LinkSetUp(link)
	if err != nil {
		return fmt.Errorf("failed to set dummy interface up: %w", err)
	}
	for _, addr := range r.DummyIP {
		err = handle.AddrAdd(link, addr)
		if err != nil {
			return fmt.Errorf("failed to add addr to dummy interface: %w", err)
		}
	}
	return nil
}

func (r *RAIT) DestroyDummyInterface() error {
	var handle *netlink.Handle
	var err error
	handle, err = netlink.NewHandle()
	if err != nil {
		return fmt.Errorf("failed to get netlink handle: %w", err)
	}
	defer handle.Delete()

	var link netlink.Link
	link, err = handle.LinkByName(r.DummyName)
	if err != nil{
		return fmt.Errorf("failed to get dummy interface: %w", err)
	}

	err = handle.LinkDel(link)
	if err != nil{
		return fmt.Errorf("failed to remove dummy interface: %w", err)
	}
	return nil
}
