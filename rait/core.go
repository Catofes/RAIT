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
	"strconv"
	"strings"
)

var ErrExpected = errors.New("some expected error has occurred")

// GenerateWireguardConfig generates wireguard interface configuration for specified peer
func (client *Client) GenerateWireguardConfig(peer *Peer) (name string, config *wgtypes.Config, err error) {
	if client.SendPort == peer.SendPort || !client.AddressFamily.Equal(peer.AddressFamily) {
		err = ErrExpected
		return
	}
	var endpoint *net.IPAddr
	endpoint, err = peer.AddressFamily.ResolveIP(peer.Endpoint)
	if err != nil {
		err = fmt.Errorf("GenerateWireguardConfig: failed to resolve endpoint address %s: %w", peer.Endpoint, err)
		return
	}
	name = client.InterfacePrefix + strconv.Itoa(peer.SendPort)
	_pk := wgtypes.Key(client.PrivateKey)
	config = &wgtypes.Config{
		PrivateKey:   &_pk,
		ListenPort:   &peer.SendPort,
		FirewallMark: &client.FwMark,
		ReplacePeers: true,
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:    wgtypes.Key(peer.PublicKey),
				Remove:       false,
				UpdateOnly:   false,
				PresharedKey: nil,
				Endpoint: &net.UDPAddr{
					IP:   endpoint.IP,
					Port: client.SendPort,
				},
				ReplaceAllowedIPs: true,
				AllowedIPs:        consts.IPNetAll,
			},
		},
	}
	return
}

// SetupWireguardInterface setups wireguard interface for specified peer
func (client *Client) SetupWireguardInterface(peer *Peer) (link netlink.Link, err error) {
	var name string
	var config *wgtypes.Config
	name, config, err = client.GenerateWireguardConfig(peer)
	if err != nil {
		return
	}

	err = utils.WithNetns(client.InterfaceNamespace, func(handle *netlink.Handle) (err error) {
		link, err = handle.LinkByName(name)
		return
	})

	if err == nil {
		// TODO: check whether the link is in desired state
	} else if _, ok := err.(netlink.LinkNotFoundError); ok {
		err = utils.WithNetns(client.TransitNamespace, func(handle *netlink.Handle) (err error) {
			link = &netlink.Wireguard{
				LinkAttrs: netlink.LinkAttrs{
					Name:  name,
					MTU:   client.MTU,
					Group: uint32(client.InterfaceGroup),
				},
			}
			err = handle.LinkAdd(link)
			if err != nil {
				err = fmt.Errorf("SetupWireguardInterface: failed to create interface in transit netns: %w", err)
				return
			}
			var ns netns.NsHandle
			ns, err = utils.GetNetns(client.TransitNamespace)
			if err != nil {
				return
			}
			defer ns.Close()
			err = handle.LinkSetNsFd(link, int(ns))
			if err != nil {
				_ = handle.LinkDel(link)
				err = fmt.Errorf("SetupWireguardInterface: failed to create interface in transit netns: %w", err)
				return
			}
			return
		})
		if err != nil {
			return
		}

		err = utils.WithNetns(client.InterfaceNamespace, func(handle *netlink.Handle) (err error) {
			err = handle.LinkSetUp(link)
			if err != nil {
				_ = handle.LinkDel(link)
				err = fmt.Errorf("SetupWireguardInterface: failed to bring interface up: %w", err)
				return
			}
			err = handle.AddrAdd(link, utils.GenerateLinklocalAddress())
			if err != nil {
				_ = handle.LinkDel(link)
				err = fmt.Errorf("SetupWireguardInterface: failed to add link local address to interface: %w", err)
				return
			}
			return
		})
		if err != nil {
			return
		}
	} else {
		return
	}

	err = utils.WithNetns(client.InterfaceNamespace, func(handle *netlink.Handle) (err error) {
		var wg *wgctrl.Client
		wg, err = wgctrl.New()
		if err != nil {
			err = fmt.Errorf("SetupWireguardInterface: failed to get wireguard control socket: %w", err)
			return
		}
		defer wg.Close()
		err = wg.ConfigureDevice(link.Attrs().Name, *config)
		if err != nil {
			err = fmt.Errorf("SetupWireguardInterface: failed to configure wireguard interface: %w", err)
			return
		}
		return
	})
	return
}

func (client *Client) ListInterfaces() (linkList []netlink.Link, err error) {
	err = utils.WithNetns(client.InterfaceNamespace, func(handle *netlink.Handle) (err error) {
		var unfilteredLinkList []netlink.Link
		unfilteredLinkList, err = handle.LinkList()
		if err != nil {
			err = fmt.Errorf("ListInterfaces: failed to list links: %w", err)
			return
		}
		for _, link := range unfilteredLinkList {
			if link.Type() == "wireguard" && strings.HasPrefix(link.Attrs().Name, client.InterfacePrefix) &&
				link.Attrs().Group == uint32(client.InterfaceGroup) {
				linkList = append(linkList, link)
			}
		}
		return
	})
	return
}

// SyncWireguardInterfaces ensures the state of the interfaces matches the configuration files
func (client *Client) SyncWireguardInterfaces(peers []*Peer) (err error) {
	var targetLinkList []netlink.Link
	for _, peer := range peers {
		var link netlink.Link
		link, err = client.SetupWireguardInterface(peer)
		if err != nil {
			if errors.Is(err, ErrExpected) {
				continue
			}
			return
		}
		targetLinkList = append(targetLinkList, link)
	}

	var currentLinkList []netlink.Link
	currentLinkList, err = client.ListInterfaces()
	if err != nil {
		return
	}

	var diff []netlink.Link
	diff = utils.LinkListDiff(currentLinkList, targetLinkList)

	err = utils.WithNetns(client.InterfaceNamespace, func(handle *netlink.Handle) (err error) {
		for _, link := range diff {
			err = handle.LinkDel(link)
			if err != nil {
				err = fmt.Errorf("SyncWireguardInterfaces: failed to delete interface: %w", err)
				return
			}
		}
		return
	})
	return
}
