package rait

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"runtime"
	"strconv"
)

var RAITFwMark = 54

const IFPrefix = "rait"

func (r *RAIT) WireguardConfig(peers []*Peer) []*wgtypes.Config {
	var configs []*wgtypes.Config
	for _, peer := range peers {
		if r.PublicKey == peer.PublicKey {
			continue // ignore self quick path
		}
		configs = append(configs, &wgtypes.Config{
			PrivateKey:   &r.PrivateKey,
			ListenPort:   &peer.SendPort,
			FirewallMark: &RAITFwMark,
			ReplacePeers: true,
			Peers: []wgtypes.PeerConfig{
				{
					PublicKey:    peer.PublicKey,
					Remove:       false,
					UpdateOnly:   false,
					PresharedKey: nil,
					Endpoint: &net.UDPAddr{
						IP:   peer.EndpointIP,
						Port: r.SendPort,
					},
					ReplaceAllowedIPs: true,
					AllowedIPs:        []net.IPNet{*IP4NetAll, *IP6NetAll},
				}},
		})
	}
	return configs
}

func (r *RAIT) Setup(peers []*Peer) error {
	// Just trying
	_ = CreateNetNSFromName(r.Namespace)
	OutsideHandle, InsideHandle, OutsideNS, InsideNS, err := GetHandles(r.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get netlink handles: %w", err)
	}
	defer InsideNS.Close()
	defer OutsideNS.Close()
	defer InsideHandle.Delete()
	defer OutsideHandle.Delete()

	// Setup wireguard
	// Before wgctrl supports net InsideNS natively, we have to do this
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	err = netns.Set(InsideNS)
	if err != nil {
		return fmt.Errorf("failed to enter netns: %w", err)
	}
	defer netns.Set(OutsideNS)

	var client *wgctrl.Client
	client, err = wgctrl.New()
	if err != nil {
		return fmt.Errorf("failed to get wireguard client: %w", err)
	}
	defer client.Close()

	configs := r.WireguardConfig(peers)

	for index, config := range configs {
		ifname := IFPrefix + strconv.Itoa(index)
		link := &netlink.Wireguard{
			LinkAttrs: netlink.LinkAttrs{
				Name: ifname,
			},
		}
		err = OutsideHandle.LinkAdd(link)
		if err != nil {
			return fmt.Errorf("failed to create link: %w", err)
		}
		err = OutsideHandle.LinkSetNsFd(link, int(InsideNS))
		if err != nil {
			return fmt.Errorf("failed to move link to netns: %w", err)
		}
		err = InsideHandle.LinkSetUp(link)
		if err != nil {
			return fmt.Errorf("failed to bring up link: %w", err)
		}
		err = InsideHandle.AddrAdd(link, RandomLinklocal())
		if err != nil {
			return fmt.Errorf("failed to add linklocal address: %w", err)
		}
		err = client.ConfigureDevice(ifname, *config)
		if err != nil {
			return fmt.Errorf("failed to configure link: %w", err)
		}
	}

	// Setup veth pair
	link := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: IFPrefix + "local",
		},
		PeerName: r.Interface,
	}
	err = InsideHandle.LinkAdd(link)
	if err != nil {
		return fmt.Errorf("failed to add veth pair: %w", err)
	}
	err = InsideHandle.LinkSetUp(link)
	if err != nil {
		return fmt.Errorf("failed to bring up veth inside: %w", err)
	}
	var peer netlink.Link
	peer, err = InsideHandle.LinkByName(r.Interface)
	if err != nil {
		return fmt.Errorf("failed to get peer: %w", err)
	}
	err = InsideHandle.LinkSetNsFd(peer, int(OutsideNS))
	if err != nil {
		return fmt.Errorf("failed to move peer to ns: %w", err)
	}

	// Fetch it again
	peer, err = OutsideHandle.LinkByName(r.Interface)
	if err != nil {
		return fmt.Errorf("failed to get peer: %w", err)
	}
	err = OutsideHandle.LinkSetUp(peer)
	if err != nil {
		return fmt.Errorf("failed to bring up peer: %w", err)
	}
	for _, addr := range r.Addresses {
		err = OutsideHandle.AddrAdd(peer, addr)
		if err != nil {
			return fmt.Errorf("failed to add addr to peer: %w", err)
		}
	}
	err = r.WriteBabeldConfig(len(peers))
	if err != nil {
		return fmt.Errorf("failed to write babeld conf: %w", err)
	}
	return nil
}

func (r *RAIT) Destroy() error {
	return DestroyNetNSFromName(r.Namespace)
}
