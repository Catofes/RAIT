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
	"strings"
)

var RAITFwMark = 54

func (r *RAIT) WireguardConfigs(peers []*Peer) []*wgtypes.Config {
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

func (r *RAIT) Init() error {
	var err error
	r.OriginalNSHandle, err = netns.Get()
	if err != nil {
		return fmt.Errorf("failed to get netns handle: %w", err)
	}
	r.OriginalNetlinkHandle, err = netlink.NewHandleAt(r.OriginalNSHandle)
	if err != nil {
		return fmt.Errorf("failed to get netlink handle: %w", err)
	}

	if r.NetNS != "" {
		r.SpecifiedNSHandle, err = netns.GetFromName(r.NetNS)
		if err != nil {
			return fmt.Errorf("failed to get netns handle: %w", err)
		}
		r.SpecifiedNetlinkHandle, err = netlink.NewHandleAt(r.SpecifiedNSHandle)
		if err != nil {
			return fmt.Errorf("failed to get netlink handle: %w", err)
		}
	} else {
		r.SpecifiedNSHandle = r.OriginalNSHandle
		r.SpecifiedNetlinkHandle = r.OriginalNetlinkHandle
	}
	return nil
}

func (r *RAIT) Destroy() {
	r.OriginalNetlinkHandle.Delete()
	r.SpecifiedNetlinkHandle.Delete()
	_ = r.OriginalNSHandle.Close()
	_ = r.SpecifiedNSHandle.Close()
}

func (r *RAIT) NSEnter() error {
	runtime.LockOSThread()
	err := netns.Set(r.SpecifiedNSHandle)
	if err != nil {
		runtime.UnlockOSThread()
	}
	return err
}

func (r *RAIT) NSLeave() error {
	err := netns.Set(r.OriginalNSHandle)
	if err == nil {
		runtime.UnlockOSThread()
	}
	return err
}

func (r *RAIT) SetupWireguardLinks(peers []*Peer) error {
	var err error
	// Before wgctrl supports net ns natively, we have to do this
	err = r.NSEnter()
	if err != nil {
		return fmt.Errorf("failed to enter netns: %w", err)
	}
	defer r.NSLeave()

	var client *wgctrl.Client
	client, err = wgctrl.New()
	if err != nil {
		return fmt.Errorf("failed to get wireguard client: %w", err)
	}
	defer client.Close()

	configs := r.WireguardConfigs(peers)

	for index, config := range configs {
		ifname := r.IFPrefix + strconv.Itoa(index)
		link := &netlink.Wireguard{
			LinkAttrs: netlink.LinkAttrs{
				Name: ifname,
			},
		}
		err = r.OriginalNetlinkHandle.LinkAdd(link)
		if err != nil {
			return fmt.Errorf("failed to create link: %w", err)
		}
		if r.NetNS != "" {
			err = r.OriginalNetlinkHandle.LinkSetNsFd(link, int(r.SpecifiedNSHandle))
			if err != nil {
				return fmt.Errorf("failed to move link to net ns: %w", err)
			}
		}
		err = r.SpecifiedNetlinkHandle.LinkSetUp(link)
		if err != nil {
			return fmt.Errorf("failed to bring up link: %w", err)
		}
		err = r.SpecifiedNetlinkHandle.AddrAdd(link, RandomLinklocal())
		if err != nil {
			return fmt.Errorf("failed to add linklocal address: %w", err)
		}
		err = client.ConfigureDevice(ifname, *config)
		if err != nil {
			return fmt.Errorf("failed to configure link: %w", err)
		}
	}
	return nil
}

func (r *RAIT) DestroyWireguardLinks() error {
	var links []netlink.Link
	var err error
	links, err = r.SpecifiedNetlinkHandle.LinkList()
	if err != nil {
		return fmt.Errorf("failed to list links: %w", err)
	}
	for _, link := range links {
		if strings.HasPrefix(link.Attrs().Name, r.IFPrefix) && link.Type() == "wireguard" {
			err = r.SpecifiedNetlinkHandle.LinkDel(link)
			if err != nil {
				return fmt.Errorf("failed to delete link: %w", err)
			}
		}
	}
	return nil
}

func (r *RAIT) SetupDummyInterface() error {
	var link = &netlink.Dummy{
		LinkAttrs: netlink.LinkAttrs{
			Name: r.DummyName,
		},
	}
	var err error
	err = r.SpecifiedNetlinkHandle.LinkAdd(link)
	if err != nil {
		return fmt.Errorf("failed to add dummy interface: %w", err)
	}
	err = r.SpecifiedNetlinkHandle.LinkSetUp(link)
	if err != nil {
		return fmt.Errorf("failed to set dummy interface up: %w", err)
	}
	for _, addr := range r.DummyIP {
		err = r.SpecifiedNetlinkHandle.AddrAdd(link, addr)
		if err != nil {
			return fmt.Errorf("failed to add addr to dummy interface: %w", err)
		}
	}
	return nil
}

func (r *RAIT) DestroyDummyInterface() error {
	var link netlink.Link
	var err error
	link, err = r.SpecifiedNetlinkHandle.LinkByName(r.DummyName)
	if err != nil {
		return fmt.Errorf("failed to get dummy interface: %w", err)
	}
	err = r.SpecifiedNetlinkHandle.LinkDel(link)
	if err != nil {
		return fmt.Errorf("failed to remove dummy interface: %w", err)
	}
	return nil
}
