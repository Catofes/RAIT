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
		ifname := r.IFPrefix + strconv.Itoa(index)
		config := SynthesisWireguardConfig(r, p)
		if config == nil {
			continue
		}

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

func DestroyWireguardWithPrefix(handle *netlink.Handle, prefix string) error {
	linkList, err := handle.LinkList()
	if err != nil {
		return fmt.Errorf("DestroyWireguardWithPrefix: failed to list interfaces: %w", err)
	}
	for _, link := range linkList {
		if link.Type() == "wireguard" && strings.HasPrefix(link.Attrs().Name, prefix) {
			err = handle.LinkDel(link)
			if err != nil {
				return fmt.Errorf("DestroyWireguardWithPrefix: failed to delete interface: %w", err)
			}
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
	err = DestroyWireguardWithPrefix(helper.SrcHandle, r.IFPrefix)
	if err != nil {
		return err
	}
	err = DestroyWireguardWithPrefix(helper.DstHandle, r.IFPrefix)
	if err != nil {
		return err
	}
	return nil
}
