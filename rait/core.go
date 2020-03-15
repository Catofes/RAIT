package rait

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"strconv"
	"strings"
)

func (r *RAIT) Setup(peers []*Peer) error {
	var helper *NamespaceHelper
	var err error
	helper, err = NamespaceHelperFromName(r.Namespace)
	if err != nil {
		return err
	}
	defer helper.Destroy()

	var client *wgctrl.Client
	client, err = wgctrl.New()
	if err != nil {
		return fmt.Errorf("failed to get wireguard client: %w", err)
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
			return fmt.Errorf("failed to create wireguard interface: %w", err)
		}
		err = client.ConfigureDevice(ifname, *config)
		if err != nil {
			return fmt.Errorf("failed to configure wireguard interface: %w", err)
		}
		err = helper.SrcHandle.LinkSetNsFd(link, int(helper.DstNamespace))
		if err != nil {
			return fmt.Errorf("failed to move interface to netns: %w", err)
		}
		err = helper.DstHandle.LinkSetUp(link)
		if err != nil {
			return fmt.Errorf("failed to bring up wireguard interface: %w", err)
		}
		err = helper.DstHandle.AddrAdd(link, RandomLinklocal())
		if err != nil {
			return fmt.Errorf("failed to add linklocal address: %w", err)
		}
	}

	peerName := r.IFPrefix + "local"
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: r.Interface,
			MTU: 65535,
		},
		PeerName: peerName,
	}
	err = helper.SrcHandle.LinkAdd(veth)
	if err != nil {
		return fmt.Errorf("failed to add veth pair: %w", err)
	}
	err = helper.SrcHandle.LinkSetUp(veth)
	if err != nil {
		return fmt.Errorf("failed to bring up veth: %w", err)
	}
	for _, addr := range r.Addresses {
		err = helper.SrcHandle.AddrAdd(veth, addr.Addr)
		if err != nil {
			return fmt.Errorf("failed to add addr to peer: %w", err)
		}
	}

	var peer netlink.Link
	peer, err = helper.SrcHandle.LinkByName(peerName)
	if err != nil {
		return fmt.Errorf("failed to get peer: %w", err)
	}
	err = helper.SrcHandle.LinkSetNsFd(peer, int(helper.DstNamespace))
	if err != nil {
		return fmt.Errorf("failed to move peer to ns: %w", err)
	}
	err = helper.DstHandle.LinkSetUp(peer)
	if err != nil {
		return fmt.Errorf("failed to bring up peer: %w", err)
	}
	err = helper.DstHandle.AddrAdd(peer, SynthesisAddress(r.Name))
	if err != nil {
		return fmt.Errorf("failed to add synthesised address: %w", err)
	}
	return nil
}

func (r *RAIT) Destroy() error {
	var helper *NamespaceHelper
	var err error
	helper, err = NamespaceHelperFromName(r.Namespace)
	if err != nil {
		return err
	}
	defer helper.Destroy()

	var veth netlink.Link
	veth, err = helper.SrcHandle.LinkByName(r.Interface)
	if err == nil {
		_ = helper.SrcHandle.LinkDel(veth)
	}

	linkList, err := helper.DstHandle.LinkList()
	if err != nil {
		return err
	}
	for _, link := range linkList {
		if link.Type() == "wireguard" && strings.HasPrefix(link.Attrs().Name, r.IFPrefix) {
			_ = helper.DstHandle.LinkDel(link)
		}
	}

	linkList, err = helper.SrcHandle.LinkList()
	if err != nil {
		return err
	}
	for _, link := range linkList {
		if link.Type() == "wireguard" && strings.HasPrefix(link.Attrs().Name, r.IFPrefix) {
			_ = helper.SrcHandle.LinkDel(link)
		}
	}

	return nil
}
