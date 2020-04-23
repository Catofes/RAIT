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

// GetNetlinkHelpers returns a pair of netlink helpers for both interface netns and transit netns
func (client *Client) GetNetlinkHelpers() (*utils.NetlinkHelper, *utils.NetlinkHelper, error) {
	var helperInterface *utils.NetlinkHelper
	var err error
	helperInterface, err = utils.NetlinkHelperFromName(client.InterfaceNamespace)
	if err != nil {
		return nil, nil, fmt.Errorf("failed too get netlink helper in interface namespace: %w", err)
	}
	var helperTransit *utils.NetlinkHelper
	helperTransit, err = utils.NetlinkHelperFromName(client.TransitNamespace)
	if err != nil {
		defer helperInterface.Destroy()
		return nil, nil, fmt.Errorf("failed too get netlink helper in transit namespace: %w", err)
	}
	return helperInterface, helperTransit, nil
}

// GenerateWireguardConfig takes a peer configuration then generates the corresponding wireguard interface configuration
func (client *Client) GenerateWireguardConfig(peer *Peer) *wgtypes.Config {
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
					IP:   peer.Endpoint.IP,
					Port: client.SendPort,
				},
				ReplaceAllowedIPs: true,
				AllowedIPs:        []net.IPNet{*consts.IP4NetAll, *consts.IP6NetAll},
			},
		},
	}
}

func (client *Client) GenerateInterfaceName(peer *Peer) string {
	return client.InterfacePrefix + strconv.Itoa(peer.SendPort)
}

// PrepareWireguardInterface gets the wireguard interface for the specific peer ready
func (client *Client) PrepareWireguardInterface(peer *Peer) (netlink.Link, error) {
	var ifName string
	ifName = client.GenerateInterfaceName(peer)

	var helperInterface *utils.NetlinkHelper
	var helperTransit *utils.NetlinkHelper
	var err error
	helperInterface, helperTransit, err = client.GetNetlinkHelpers()
	if err != nil {
		return nil, err
	}
	defer helperInterface.Destroy()
	defer helperTransit.Destroy()

	var link netlink.Link
	link, err = helperInterface.NetlinkHandle.LinkByName(ifName)
	if err == nil {
		// TODO: check whether the link is in desired state
		return link, nil
	}

	if _, ok := err.(netlink.LinkNotFoundError); ok {
		link = &netlink.Wireguard{
			LinkAttrs: netlink.LinkAttrs{
				Name: ifName,
				MTU:  client.MTU,
			},
		}
		err = helperTransit.NetlinkHandle.LinkAdd(link)
		if err != nil {
			return nil, fmt.Errorf("failed to create interface in transit netns: %w", err)
		}
		err = helperTransit.NetlinkHandle.LinkSetNsFd(link, int(helperInterface.NamespaceHandle))
		if err != nil {
			_ = helperTransit.NetlinkHandle.LinkDel(link)
			return nil, fmt.Errorf("failed to move interface into interface netns: %w", err)
		}
		err = helperInterface.NetlinkHandle.LinkSetUp(link)
		if err != nil {
			_ = helperInterface.NetlinkHandle.LinkDel(link)
			return nil, fmt.Errorf("failed to bring interface up: %w", err)
		}
		err = helperInterface.NetlinkHandle.AddrAdd(link, utils.GenerateLinklocalAddress())
		if err != nil {
			_ = helperInterface.NetlinkHandle.LinkDel(link)
			return nil, fmt.Errorf("failed to add link local address to interface: %w", err)
		}
	} else {
		return nil, err
	}
	return link, nil
}

// SetupWireguardInterface setups a single wireguard interface according to peer configuration
func (client *Client) ConfigureWireguardInterface(peer *Peer, link netlink.Link) error {
	var config *wgtypes.Config
	config = client.GenerateWireguardConfig(peer)
	// Since wgctrl does not currently support specifying a netns, we have to move into interface netns to configure it
	// Borrowed some magic from libpod
	var wtg sync.WaitGroup
	wtg.Add(1)
	var err error
	go (func() {
		defer wtg.Done()
		runtime.LockOSThread()
		var helper *utils.NetlinkHelper
		helper,err = utils.NetlinkHelperFromName(client.InterfaceNamespace)
		if err != nil {
			err = fmt.Errorf("failed to get interface netns netlink helper: %w", err)
			return
		}
		defer helper.Destroy()
		err = netns.Set(helper.NamespaceHandle)
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
		err = wg.ConfigureDevice(link.Attrs().Name, *config)
		if err != nil {
			err = fmt.Errorf("failed to configure wireguard interface: %w", err)
			return
		}
	})()
	wtg.Wait()
	if err != nil {
		return fmt.Errorf("failed to configure interface: %w", err)
	}
	return nil
}

// SyncWireguardInterfaces ensures the state of the interfaces matches the configuration files
func (client *Client) SyncWireguardInterfaces(peers []*Peer) error {
	var err error
	var link netlink.Link
	var targetLinkList []netlink.Link
	for _, peer := range peers {
		if client.SendPort == peer.SendPort {
			continue
		}
		link, err = client.PrepareWireguardInterface(peer)
		if err != nil {
			return err
		}
		err = client.ConfigureWireguardInterface(peer, link)
		if err != nil {
			return err
		}
		targetLinkList = append(targetLinkList, link)
	}

	var helperInterface *utils.NetlinkHelper
	var helperTransit *utils.NetlinkHelper
	helperInterface, helperTransit, err = client.GetNetlinkHelpers()
	if err != nil {
		return err
	}
	defer helperInterface.Destroy()
	defer helperTransit.Destroy()

	var currentLinkList []netlink.Link
	currentLinkList, err = helperInterface.NetlinkHandle.LinkList()
	if err != nil {
		return fmt.Errorf("failed to list links: %w", err)
	}

	var diff []netlink.Link
	diff = utils.LinkListDiff(currentLinkList, targetLinkList)

	for _, link := range diff {
		if link.Type() == "wireguard" && strings.HasPrefix(link.Attrs().Name, client.InterfacePrefix) {
			err = helperInterface.NetlinkHandle.LinkDel(link)
			if err != nil {
				return fmt.Errorf("failed to delete interface: %w", err)
			}
		}
	}
	return nil
}
