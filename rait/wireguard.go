package rait

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"strconv"
	"strings"
)

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
