package rait

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"strconv"
	"strings"
)

func (r *RAIT) SetupVethPair() error {
	helper, err := NamespaceHelperFromName(r.Namespace)
	if err != nil {
		return fmt.Errorf("SetupVethPair: failed to get netns helper: %w", err)
	}
	defer helper.Destroy()
	peerName := r.IFPrefix + "local"
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: r.Interface,
			MTU:  int(r.MTU),
		},
		PeerName: peerName,
	}
	err = helper.SrcHandle.LinkAdd(veth)
	if err != nil {
		return fmt.Errorf("SetupVethPair: failed to add veth pair to calling netns: %w", err)
	}
	err = helper.SrcHandle.LinkSetUp(veth)
	if err != nil {
		return fmt.Errorf("SetupVethPair: failed to bring up veth peer in calling netns: %w", err)
	}
	for _, addr := range r.Addresses {
		err = helper.SrcHandle.AddrAdd(veth, addr.Addr)
		if err != nil {
			return fmt.Errorf("SetupVethPair: failed to add addr to veth peer in calling netns: %w", err)
		}
	}
	peer, err := helper.SrcHandle.LinkByName(peerName)
	if err != nil {
		return fmt.Errorf("SetupVethPair: failed to get veth peer in calling netns: %w", err)
	}
	err = helper.SrcHandle.LinkSetNsFd(peer, int(helper.DstNamespace))
	if err != nil {
		return fmt.Errorf("SetupVethPair: failed to move veth peer to specified netns: %w", err)
	}
	err = helper.DstHandle.LinkSetUp(peer)
	if err != nil {
		return fmt.Errorf("SetupVethPair: failed to bring up peer in specified netns: %w", err)
	}
	err = helper.DstHandle.AddrAdd(peer, SynthesisAddress(r.Name))
	if err != nil {
		return fmt.Errorf("SetupVethPair: failed to add synthesised addr to veth peer in specified netns: %w", err)
	}
	return nil
}

func (r *RAIT) DestroyVethPair() error {
	helper, err := NamespaceHelperFromName(r.Namespace)
	if err != nil {
		return fmt.Errorf("DestroyVethPair: failed to get netns helper: %w", err)
	}
	defer helper.Destroy()
	veth, err := helper.SrcHandle.LinkByName(r.Interface)
	if err != nil {
		return fmt.Errorf("DestroyVethPair: failed to get veth pair in calling netns: %w")
	}
	err = helper.SrcHandle.LinkDel(veth)
	if err != nil {
		return fmt.Errorf("DestroyVethPair: failed to delete veth pair in calling netns: %w")
	}
	return nil
}

func (r *RAIT) SetupWireguard(peers []*Peer) error {
	helper, err := NamespaceHelperFromName(r.Namespace)
	if err != nil {
		return fmt.Errorf("SetupWireguard: failed to get netns helper: %w", err)
	}
	defer helper.Destroy()
	client, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("SetupWireguard: failed to get wireguard client: %w", err)
	}
	defer client.Close()
	for index, p := range peers {
		config := SynthesisWireguardConfig(r, p)
		if config == nil {
			continue
		}
		ifname := r.IFPrefix + strconv.Itoa(index)
		link := &netlink.Wireguard{
			LinkAttrs: netlink.LinkAttrs{
				Name: ifname,
				MTU:  int(r.MTU),
			},
		}
		err = helper.SrcHandle.LinkAdd(link)
		if err != nil {
			return fmt.Errorf("SetupWireguard: failed to create wireguard interface: %w", err)
		}
		err = client.ConfigureDevice(ifname, *config)
		if err != nil {
			return fmt.Errorf("SetupWireguard: failed to configure wireguard interface: %w", err)
		}
		err = helper.SrcHandle.LinkSetNsFd(link, int(helper.DstNamespace))
		if err != nil {
			return fmt.Errorf("SetupWireguard: failed to move wireguard interface to specified netns: %w", err)
		}
		err = helper.DstHandle.LinkSetUp(link)
		if err != nil {
			return fmt.Errorf("SetupWireguard: failed to bring up wireguard interface in specified netns: %w", err)
		}
		err = helper.DstHandle.AddrAdd(link, RandomLinklocal())
		if err != nil {
			return fmt.Errorf("SetupWireguard: failed to add linklocal addr to wireguard interface in specified netns: %w", err)
		}
	}
	return nil
}

func (r *RAIT) DestroyWireguard() error {
	helper, err := NamespaceHelperFromName(r.Namespace)
	if err != nil {
		return fmt.Errorf("DestroyWireguard: failed to get netns helper: %w", err)
	}
	defer helper.Destroy()
	linkList, err := helper.DstHandle.LinkList()
	if err != nil {
		return fmt.Errorf("DestroyWireguard: failed to list interfaces in specified netns: %w", err)
	}
	for _, link := range linkList {
		if link.Type() == "wireguard" && strings.HasPrefix(link.Attrs().Name, r.IFPrefix) {
			// Never hard fail, I mean, it's just destroying links.
			_ = helper.DstHandle.LinkDel(link)
		}
	}
	linkList, err = helper.SrcHandle.LinkList()
	if err != nil {
		return fmt.Errorf("DestroyWireguard: failed to list interfaces in calling netns: %w", err)
	}
	for _, link := range linkList {
		if link.Type() == "wireguard" && strings.HasPrefix(link.Attrs().Name, r.IFPrefix) {
			_ = helper.SrcHandle.LinkDel(link)
		}
	}
	return nil
}
