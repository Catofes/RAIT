package rait

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
	"gitlab.com/NickCao/RAIT/v2/pkg/namespace"
	"gitlab.com/NickCao/RAIT/v2/pkg/types"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"strconv"
	"strings"
)

func (instance *Instance) LoadPeers() ([]*Peer, error) {
	peers, err := PeersFromPath(instance.Peers)
	if err != nil {
		return nil, err
	}
	n := 0
	for _, x := range peers {
		if x.AddressFamily == instance.AddressFamily &&
			x.PublicKey.String() != instance.PrivateKey.PublicKey().String() {
			peers[n] = x
			n++
		}
	}
	peers = peers[:n]
	return peers, nil
}

func (instance *Instance) WireguardConfig(peer *Peer) (string, *wgtypes.Config, error) {
	return instance.InterfacePrefix + strconv.Itoa(peer.SendPort), &wgtypes.Config{
		PrivateKey:   &instance.PrivateKey,
		ListenPort:   &peer.SendPort,
		FirewallMark: &instance.FwMark,
		ReplacePeers: true,
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:    peer.PublicKey,
				Remove:       false,
				UpdateOnly:   false,
				PresharedKey: nil,
				Endpoint: &net.UDPAddr{
					IP:   peer.Endpoint,
					Port: instance.SendPort,
				},
				ReplaceAllowedIPs: true,
				AllowedIPs:        types.IPNetAll,
			},
		},
	}, nil
}

func (instance *Instance) EnsureInterface(peer *Peer) (netlink.Link, error) {
	name, config, err := instance.WireguardConfig(peer)
	if err != nil {
		return nil, err
	}
	var link netlink.Link
	err = namespace.WithNetlink(instance.InterfaceNamespace, func(handle *netlink.Handle) error {
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

	err = namespace.WithNetlink(instance.TransitNamespace, func(handle *netlink.Handle) error {
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

		err = handle.LinkSetNsFd(link, int(instance.InterfaceNamespace))
		if err != nil {
			_ = handle.LinkDel(link)
			return fmt.Errorf("EnsureInterface: failed to move interface into netns: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = namespace.WithNetlink(instance.InterfaceNamespace, func(handle *netlink.Handle) error {
		err = handle.LinkSetUp(link)
		if err != nil {
			_ = handle.LinkDel(link)
			return fmt.Errorf("EnsureInterface: failed to bring interface up: %w", err)
		}
		err = handle.AddrAdd(link, misc.LinkLocalAddr())
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
	err = namespace.WithNetlink(instance.InterfaceNamespace, func(handle *netlink.Handle) error {
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

func (instance *Instance) ListInterfaces() ([]netlink.Link, error) {
	var err error
	var linkList []netlink.Link
	err = namespace.WithNetlink(instance.InterfaceNamespace, func(handle *netlink.Handle) error {
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

func (instance *Instance) IsManagedInterface(link netlink.Link) bool {
	return link.Type() == "wireguard" &&
		link.Attrs().Group == uint32(instance.InterfaceGroup) &&
		strings.HasPrefix(link.Attrs().Name, instance.InterfacePrefix)
}

func (instance *Instance) SyncInterfaces(up bool) error {
	var peers []*Peer
	var err error
	if up {
		peers, err = instance.LoadPeers()
		if err != nil {
			return err
		}
	}

	var targetLinkList []netlink.Link
	for _, peer := range peers {
		link, err := instance.EnsureInterface(peer)
		if err != nil {
			return err
		}
		if link != nil {
			targetLinkList = append(targetLinkList, link)
		}
	}

	currentLinkList, err := instance.ListInterfaces()
	if err != nil {
		return err
	}

	err = namespace.WithNetlink(instance.InterfaceNamespace, func(handle *netlink.Handle) error {
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
