package rait

import (
	"fmt"
	"github.com/vishvananda/netlink"
)

func (r *RAIT) SetupVethPair() error {
	if r.Veth == "off" || r.Namespace == "off" {
		return nil
	}
	helper, err := NamespaceHelperFromName(r.Namespace)
	if err != nil {
		return fmt.Errorf("SetupVethPair: failed to get netns helper: %w", err)
	}
	defer helper.Destroy()

	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: "host",
			MTU:  int(r.MTU),
		},
		PeerName: r.Veth,
	}
	err = helper.DstHandle.LinkAdd(veth)
	if err != nil {
		return fmt.Errorf("SetupVethPair: failed to add veth pair to specified netns: %w", err)
	}
	err = helper.DstHandle.LinkSetUp(veth)
	if err != nil {
		return fmt.Errorf("SetupVethPair: failed to bring up veth peer in specified netns: %w", err)
	}

	peer, err := helper.DstHandle.LinkByName(veth.PeerName)
	if err != nil {
		return fmt.Errorf("SetupVethPair: failed to get veth peer in specified netns: %w", err)
	}
	err = helper.DstHandle.LinkSetNsFd(peer, int(helper.SrcNamespace))
	if err != nil {
		return fmt.Errorf("SetupVethPair: failed to move veth peer to calling netns: %w", err)
	}
	peer, err = helper.SrcHandle.LinkByName(veth.PeerName)
	if err != nil {
		return fmt.Errorf("SetupVethPair: failed to get veth peer in calling netns: %w", err)
	}
	err = helper.SrcHandle.LinkSetUp(peer)
	if err != nil {
		return fmt.Errorf("SetupVethPair: failed to bring up peer in calling netns: %w", err)
	}
	return nil
}

func (r *RAIT) DestroyVethPair() error {
	if r.Veth == "off" || r.Namespace == "off" {
		return nil
	}
	helper, err := NamespaceHelperFromName(r.Namespace)
	if err != nil {
		return fmt.Errorf("DestroyVethPair: failed to get netns helper: %w", err)
	}
	defer helper.Destroy()

	veth, err := helper.SrcHandle.LinkByName(r.Veth)
	if err != nil {
		return fmt.Errorf("DestroyVethPair: failed to get veth pair in calling netns: %w")
	}
	err = helper.SrcHandle.LinkDel(veth)
	if err != nil {
		return fmt.Errorf("DestroyVethPair: failed to delete veth pair in calling netns: %w")
	}
	return nil
}
