package rait

import (
	"errors"
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"gitlab.com/NickCao/RAIT/rait/internal/consts"
	"gitlab.com/NickCao/RAIT/rait/internal/utils"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var ErrPeerIsSelf = errors.New("the peer is the same node as client")
var ErrIncompatibleAF = errors.New("the peer has address family different from client")

// GenerateWireguardConfig generates wireguard interface configuration for specified peer
func (client *Client) GenerateWireguardConfig(peer *Peer) (string, *wgtypes.Config, error) {
	if client.SendPort == peer.SendPort {
		return "", nil, ErrPeerIsSelf
	}
	if client.AddressFamily != peer.AddressFamily {
		return "", nil, ErrIncompatibleAF
	}
	var endpoint *net.IPAddr
	var err error
	endpoint, err = net.ResolveIPAddr(peer.AddressFamily.String(), peer.Endpoint)
	if err != nil {
		return "", nil, fmt.Errorf("failed to resolve endpoint endpoint address: %w", err)
	}
	return client.InterfacePrefix + strconv.Itoa(peer.SendPort), &wgtypes.Config{
		PrivateKey:   &client.PrivateKey.Key,
		ListenPort:   &peer.SendPort,
		FirewallMark: client.FwMark,
		ReplacePeers: true,
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:    peer.PublicKey.Key,
				Remove:       false,
				UpdateOnly:   false,
				PresharedKey: nil,
				Endpoint: &net.UDPAddr{
					IP:   endpoint.IP,
					Port: client.SendPort,
				},
				ReplaceAllowedIPs: true,
				AllowedIPs:        []net.IPNet{*consts.IP4NetAll, *consts.IP6NetAll},
			},
		},
	}, nil
}

// SetupWireguardInterface setups wireguard interface for specified peer
func (client *Client) SetupWireguardInterface(peer *Peer) (netlink.Link, error) {
	var name string
	var config *wgtypes.Config
	var err error
	name, config, err = client.GenerateWireguardConfig(peer)
	if err != nil {
		return nil, err
	}

	var interfaceHelper *utils.NetlinkHelper
	var transitHelper *utils.NetlinkHelper
	interfaceHelper, err = utils.NetlinkHelperFromName(client.InterfaceNamespace)
	if err != nil {
		return nil, err
	}
	defer interfaceHelper.Destroy()
	transitHelper, err = utils.NetlinkHelperFromName(client.TransitNamespace)
	if err != nil {
		return nil, err
	}
	defer transitHelper.Destroy()

	var link netlink.Link
	link, err = interfaceHelper.NetlinkHandle.LinkByName(name)
	if err == nil {
		// TODO: check whether the link is in desired state
	} else {
		if _, ok := err.(netlink.LinkNotFoundError); ok {
			link = &netlink.Wireguard{
				LinkAttrs: netlink.LinkAttrs{
					Name: name,
					MTU:  client.MTU,
				},
			}
			err = transitHelper.NetlinkHandle.LinkAdd(link)
			if err != nil {
				return nil, fmt.Errorf("failed to create interface in transit netns: %w", err)
			}
			err = transitHelper.NetlinkHandle.LinkSetNsFd(link, int(interfaceHelper.NamespaceHandle))
			if err != nil {
				_ = transitHelper.NetlinkHandle.LinkDel(link)
				return nil, fmt.Errorf("failed to move interface into interface netns: %w", err)
			}
			err = interfaceHelper.NetlinkHandle.LinkSetUp(link)
			if err != nil {
				_ = interfaceHelper.NetlinkHandle.LinkDel(link)
				return nil, fmt.Errorf("failed to bring interface up: %w", err)
			}
			err = interfaceHelper.NetlinkHandle.AddrAdd(link, utils.GenerateLinklocalAddress())
			if err != nil {
				_ = interfaceHelper.NetlinkHandle.LinkDel(link)
				return nil, fmt.Errorf("failed to add link local address to interface: %w", err)
			}
		} else {
			return nil, err
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go (func() {
		defer wg.Done()
		runtime.LockOSThread()
		err = netns.Set(interfaceHelper.NamespaceHandle)
		if err != nil {
			err = fmt.Errorf("failed to move into interface netns: %w", err)
			return
		}
		var wgc *wgctrl.Client
		wgc, err = wgctrl.New()
		if err != nil {
			err = fmt.Errorf("failed to get wireguard control socket: %w", err)
			return
		}
		defer wgc.Close()
		err = wgc.ConfigureDevice(link.Attrs().Name, *config)
		if err != nil {
			err = fmt.Errorf("failed to configure wireguard interface: %w", err)
			return
		}
	})()
	wg.Wait()
	if err != nil {
		_ = interfaceHelper.NetlinkHandle.LinkDel(link)
		return nil, fmt.Errorf("failed to configure interface: %w", err)
	}
	return link, nil
}

// SyncWireguardInterfaces ensures the state of the interfaces matches the configuration files
func (client *Client) SyncWireguardInterfaces(peers []*Peer) error {
	var targetLinkList []netlink.Link
	for _, peer := range peers {
		var link netlink.Link
		var err error
		link, err = client.SetupWireguardInterface(peer)
		if err != nil {
			if errors.Is(err, ErrPeerIsSelf) || errors.Is(err, ErrIncompatibleAF) {
				continue
			}
			return err
		}
		targetLinkList = append(targetLinkList, link)
	}

	var interfaceHelper *utils.NetlinkHelper
	var err error
	interfaceHelper, err = utils.NetlinkHelperFromName(client.InterfaceNamespace)
	if err != nil {
		return err
	}
	defer interfaceHelper.Destroy()

	var currentLinkList []netlink.Link
	currentLinkList, err = interfaceHelper.NetlinkHandle.LinkList()
	if err != nil {
		return fmt.Errorf("failed to list links: %w", err)
	}

	var diff []netlink.Link
	diff = utils.LinkListDiff(currentLinkList, targetLinkList)

	for _, link := range diff {
		if link.Type() == "wireguard" && strings.HasPrefix(link.Attrs().Name, client.InterfacePrefix) {
			err = interfaceHelper.NetlinkHandle.LinkDel(link)
			if err != nil {
				return fmt.Errorf("failed to delete interface: %w", err)
			}
		}
	}
	return nil
}
