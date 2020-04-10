package rait

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"gitlab.com/NickCao/RAIT/rait/internal/consts"
	"gitlab.com/NickCao/RAIT/rait/internal/types"
	"gitlab.com/NickCao/RAIT/rait/internal/utils"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// Client represents the control structure of rait
type Client struct {
	PrivateKey         types.Key // The common private key of wireguard interfaces managed by rait
	SendPort           int       // The listen port of all peers this client connects to
	InterfaceNamespace string    // The netns to move wireguard interface into, "current" to keep them in the original netns
	TransitNamespace   string    // The netns to create wireguard interface in, "current" to create them in the original netns
	InterfacePrefix    string    // The common prefix to name the created wireguard interfaces
	MTU                int
	FwMark             int
}

// ClientFromURL loads client configuration from a toml file
func ClientFromURL(url string) (*Client, error) {
	var data []byte
	var err error
	data, err = utils.FileFromURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get client from url: %w", err)
	}
	var client = Client{
		InterfaceNamespace: "current",
		TransitNamespace:   "current",
		InterfacePrefix:    "rait",
		MTU:                1400,
	}
	_, err = toml.Decode(string(data), &client)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

// GenerateWireguardConfig takes a peer configuration then generates the corresponding wireguard interface configuration
// Notice that if the peer and client share the same public key, the peer will be ignored and nil will be returned
func (client *Client) GenerateWireguardConfig(peer *Peer) *wgtypes.Config {
	if client.PrivateKey.Key.PublicKey() == peer.PublicKey.Key {
		return nil
	}
	return &wgtypes.Config{
		PrivateKey:   &client.PrivateKey.Key,
		ListenPort:   &peer.SendPort,
		FirewallMark: &client.FwMark,
		ReplacePeers: true,
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:    peer.PublicKey.Key,
				Remove:       false,
				UpdateOnly:   false,
				PresharedKey: nil,
				Endpoint: &net.UDPAddr{
					IP:   peer.Endpoint,
					Port: client.SendPort,
				},
				ReplaceAllowedIPs: true,
				AllowedIPs:        []net.IPNet{*consts.IP4NetAll, *consts.IP6NetAll},
			},
		},
	}
}

// SetupWireguardInterface setups a single wireguard interface according to peer configuration
func (client *Client) SetupWireguardInterface(peer *Peer, interfaceSuffix string) error {
	var config *wgtypes.Config
	config = client.GenerateWireguardConfig(peer)
	if config == nil {
		return nil
	}

	var ifName string
	ifName = client.InterfacePrefix + interfaceSuffix

	var helperInterface *utils.NetlinkHelper
	var err error
	helperInterface, err = utils.NetlinkHelperFromName(client.InterfaceNamespace)
	if err != nil {
		return err
	}
	defer helperInterface.Destroy()
	var helperTransit *utils.NetlinkHelper
	helperTransit, err = utils.NetlinkHelperFromName(client.TransitNamespace)
	if err != nil {
		return err
	}
	defer helperTransit.Destroy()

	var link *netlink.Wireguard
	link = &netlink.Wireguard{
		LinkAttrs: netlink.LinkAttrs{
			Name: ifName,
			MTU:  client.MTU,
		},
	}

	err = helperTransit.NetlinkHandle.LinkAdd(link)
	if err != nil {
		return fmt.Errorf("failed to create interface in transit netns: %w", err)
	}
	err = helperTransit.NetlinkHandle.LinkSetNsFd(link, int(helperInterface.NamespaceHandle))
	if err != nil {
		_ = helperTransit.NetlinkHandle.LinkDel(link)
		return fmt.Errorf("failed to move interface into interface netns: %w", err)
	}
	err = helperInterface.NetlinkHandle.LinkSetUp(link)
	if err != nil {
		_ = helperInterface.NetlinkHandle.LinkDel(link)
		return fmt.Errorf("failed to bring interface up: %w", err)
	}
	err = helperInterface.NetlinkHandle.AddrAdd(link, utils.GenerateLinklocalAddress())
	if err != nil {
		_ = helperInterface.NetlinkHandle.LinkDel(link)
		return fmt.Errorf("failed to add link local address to interface: %w", err)
	}
	// Since wgctrl does not currently support specifying a netns, we have to move into interface netns to configure it
	// Borrowed some magic from libpod
	var wtg sync.WaitGroup
	wtg.Add(1)
	go (func() {
		defer wtg.Done()
		runtime.LockOSThread()
		err = netns.Set(helperInterface.NamespaceHandle)
		if err != nil {
			err = fmt.Errorf("failed to move into interface netns: %w", err)
			return
		}
		var wg *wgctrl.Client
		wg, err = wgctrl.New()
		if err != nil {
			err = fmt.Errorf("failed to get wireguard control socket: %w", err)
			return
		}
		defer wg.Close()
		err = wg.ConfigureDevice(ifName, *config)
		if err != nil {
			err = fmt.Errorf("failed to configure wireguard interface: %w", err)
			return
		}
	})()
	wtg.Wait()
	if err != nil {
		_ = helperInterface.NetlinkHandle.LinkDel(link)
		return fmt.Errorf("failed to configure interface: %w", err)
	}
	return nil
}

// SetupWireguardInterfaces calls SetupWireguardInterface for a list of peers
func (client *Client) SetupWireguardInterfaces(peers []*Peer) error {
	var err error
	for index, peer := range peers {
		err = client.SetupWireguardInterface(peer, strconv.Itoa(index))
		if err != nil {
			return err
		}
	}
	return nil
}

// DestroyWireguardInterfaces destroys all the wireguard interface with specified prefix in interface netns
func (client *Client) DestroyWireguardInterfaces() error {
	var helper *utils.NetlinkHelper
	var err error
	helper, err = utils.NetlinkHelperFromName(client.InterfaceNamespace)
	if err != nil {
		return fmt.Errorf("failed to get netlink helper: %w", err)
	}
	defer helper.Destroy()

	var linkList []netlink.Link
	linkList, err = helper.NetlinkHandle.LinkList()
	if err != nil {
		return fmt.Errorf("failed to list links: %w", err)
	}
	for _, link := range linkList {
		if link.Type() == "wireguard" && strings.HasPrefix(link.Attrs().Name, client.InterfacePrefix) {
			err = helper.NetlinkHandle.LinkDel(link)
			if err != nil {
				return fmt.Errorf("failed to delete interface: %w", err)
			}
		}
	}
	return nil
}
