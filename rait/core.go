package rait

import (
	"errors"
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"gitlab.com/NickCao/RAIT/rait/types"
	"gitlab.com/NickCao/RAIT/rait/utils"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"log"
	"net"
	"strconv"
	"strings"
)

var ErrExpected = errors.New("expected error")

// WireguardConfig generates wireguard interface configuration for specified peer
func (instance *Instance) WireguardConfig(peer *Peer) (string, *wgtypes.Config, error) {
	if instance.SendPort == peer.SendPort || instance.AddressFamily != peer.AddressFamily {
		return "", nil, ErrExpected
	}
	endpoint, err := net.ResolveIPAddr(peer.AddressFamily, peer.Endpoint)
	if err != nil {
		return "", nil, fmt.Errorf("WireguardConfig: failed to resolve endpoint address %s: %w", peer.Endpoint, err)
	}
	privKey, err := wgtypes.ParseKey(instance.PrivateKey)
	if err != nil {
		return "", nil, fmt.Errorf("WireguardConfig: failed to parse private key: %w", err)
	}
	pubKey, err := wgtypes.ParseKey(peer.PublicKey)
	if err != nil {
		return "", nil, fmt.Errorf("WireguardConfig: failed to parse public key of peer: %w", err)
	}
	return instance.InterfacePrefix + strconv.Itoa(peer.SendPort), &wgtypes.Config{
		PrivateKey:   &privKey,
		ListenPort:   &peer.SendPort,
		FirewallMark: &instance.FwMark,
		ReplacePeers: true,
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:    pubKey,
				Remove:       false,
				UpdateOnly:   false,
				PresharedKey: nil,
				Endpoint: &net.UDPAddr{
					IP:   endpoint.IP,
					Port: instance.SendPort,
				},
				ReplaceAllowedIPs: true,
				AllowedIPs:        types.IPNetAll,
			},
		},
	}, nil
}

// EnsureInterface setups wireguard interface for specified peer
func (instance *Instance) EnsureInterface(peer *Peer) (netlink.Link, error) {
	name, config, err := instance.WireguardConfig(peer)
	if err != nil {
		return nil, err
	}
	var link netlink.Link
	err = utils.WithNetNS(instance.InterfaceNamespace, func(handle *netlink.Handle) error {
		link, err = handle.LinkByName(name)
		return err
	})

	if err == nil {
		// TODO: check whether the interface is in desired state
		goto configure
	}

	if _, ok := err.(netlink.LinkNotFoundError); !ok {
		return nil, fmt.Errorf("EnsureInterface: unexpected error when trying to get link: %w", err)
	}

	err = utils.WithNetNS(instance.TransitNamespace, func(handle *netlink.Handle) error {
		link = &netlink.Wireguard{
			LinkAttrs: netlink.LinkAttrs{
				Name:  name,
				MTU:   instance.MTU,
				Group: uint32(instance.InterfaceGroup),
			},
		}
		err = handle.LinkAdd(link)
		if err != nil {
			return fmt.Errorf("EnsureInterface: failed to create interface in transit netns: %w", err)
		}
		var ns netns.NsHandle
		ns, err = utils.EnsureNetNS(instance.InterfaceNamespace)
		if err != nil {
			return err
		}
		defer ns.Close()
		err = handle.LinkSetNsFd(link, int(ns))
		if err != nil {
			_ = handle.LinkDel(link)
			return fmt.Errorf("EnsureInterface: failed to move interface into netns: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = utils.WithNetNS(instance.InterfaceNamespace, func(handle *netlink.Handle) error {
		err = handle.LinkSetUp(link)
		if err != nil {
			_ = handle.LinkDel(link)
			return fmt.Errorf("EnsureInterface: failed to bring interface up: %w", err)
		}
		err = handle.AddrAdd(link, utils.LinkLocalAddr())
		if err != nil {
			_ = handle.LinkDel(link)
			return fmt.Errorf("EnsureInterface: failed to add link local address to interface: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

configure:
	err = utils.WithNetNS(instance.InterfaceNamespace, func(handle *netlink.Handle) error {
		var wg *wgctrl.Client
		wg, err = wgctrl.New()
		if err != nil {
			return fmt.Errorf("EnsureInterface: failed to get wireguard control socket: %w", err)
		}
		defer wg.Close()
		err = wg.ConfigureDevice(link.Attrs().Name, *config)
		if err != nil {
			return fmt.Errorf("EnsureInterface: failed to configure wireguard interface: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return link, nil
}

// ListInterfaces lists the interfaces managed by rait
func (instance *Instance) ListInterfaces() ([]netlink.Link, error) {
	var err error
	var linkList []netlink.Link
	err = utils.WithNetNS(instance.InterfaceNamespace, func(handle *netlink.Handle) error {
		var unfilteredLinkList []netlink.Link
		unfilteredLinkList, err = handle.LinkList()
		if err != nil {
			return fmt.Errorf("ListInterfaces: failed to list links: %w", err)
		}
		for _, link := range unfilteredLinkList {
			if instance.IsManagedInterface(link) {
				linkList = append(linkList, link)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return linkList, nil
}

// IsManagedInterface indicates whether the specified link is managed by the current rait instance
func (instance *Instance) IsManagedInterface(link netlink.Link) bool {
	return link.Type() == "wireguard" &&
		link.Attrs().Group == uint32(instance.InterfaceGroup) &&
		strings.HasPrefix(link.Attrs().Name, instance.InterfacePrefix)
}

// SyncInterfaces ensures the state of the interfaces matches the configuration files
func (instance *Instance) SyncInterfaces(up bool) error {
	var err error
	var plist PeerList
	if up {
		plist, err = LoadPeersFromPath(instance.Peers)
		if err != nil {
			return err
		}
	}

	var targetLinkList []netlink.Link
	for _, peer := range plist.Peers {
		link, err := instance.EnsureInterface(peer)
		if err != nil && !errors.Is(err, ErrExpected) {
			log.Println(err)
			continue // Don't let a single peer fail the whole process
		}
		if link != nil {
			targetLinkList = append(targetLinkList, link)
		}
	}

	currentLinkList, err := instance.ListInterfaces()
	if err != nil {
		return err
	}

	err = utils.WithNetNS(instance.InterfaceNamespace, func(handle *netlink.Handle) error {
		for _, link := range currentLinkList {

			var unneeded = true
			for _, otherLink := range targetLinkList {
				if link.Attrs().Name == otherLink.Attrs().Name {
					unneeded = false
				}
			}
			if unneeded {
				err = handle.LinkDel(link)
				if err != nil {
					return fmt.Errorf("SyncInterfaces: failed to delete interface: %w", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
