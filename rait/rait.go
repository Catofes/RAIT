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
)

type RAIT struct {
	PrivateKey *wgtypes.Key
	PublicKey  *wgtypes.Key
	SendPort   int
}

type RAITConfig struct {
	PrivateKey string
	SendPort   int
}

func NewRAIT(config *RAITConfig) (*RAIT, error) {
	var privatekey wgtypes.Key
	var publickey wgtypes.Key
	var err error
	privatekey, err = wgtypes.ParseKey(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private privatekey: %w", err)
	}
	publickey = privatekey.PublicKey()
	return &RAIT{
		PrivateKey: &privatekey,
		PublicKey:  &publickey,
		SendPort:   config.SendPort,
	}, nil
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

func (r *RAIT) WireguardConfigs(peers []*Peer) []*wgtypes.Config {
	var configs []*wgtypes.Config
	for _, peer := range peers {
		if *r.PublicKey == peer.PublicKey {
			continue
		}
		var endpoint *net.UDPAddr
		if peer.EndpointIP != nil {
			endpoint = &net.UDPAddr{
				IP:   *peer.EndpointIP,
				Port: r.SendPort,
			}
		} else {
			// This is to make babeld happy
			// fakeip := net.ParseIP("1.1.1.1")
			// endpoint = &net.UDPAddr{
			// 	   IP:   fakeip,
			//     Port: r.SendPort,
			// }
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
			PrivateKey:   r.PrivateKey,
			ListenPort:   &peer.SendPort,
			FirewallMark: &mark,
			ReplacePeers: true,
			Peers:        []wgtypes.PeerConfig{peerconfig},
		}
		configs = append(configs, config)
	}
	return configs
}

func (r *RAIT) SetupLinks(ifprefix string, peers []*Peer) error {
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

	configs := r.WireguardConfigs(peers)
	var counter = 0
	for _, config := range configs {
		ifname := ifprefix + strconv.Itoa(counter)
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
