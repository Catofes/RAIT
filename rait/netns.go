package rait

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"os/exec"
)

func CreateNetNSFromName(name string) error {
	cmd := exec.Command("ip", "netns", "add", name)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create named netns: %w", err)
	}
	return nil
}

func DestroyNetNSFromName(name string) error {
	cmd := exec.Command("ip", "netns", "delete", name)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to destroy named netns: %w", err)
	}
	return nil
}

func GetHandles(name string) (*netlink.Handle, *netlink.Handle, netns.NsHandle, netns.NsHandle, error) {
	var err error
	var InsideNS netns.NsHandle
	var OutsideNS netns.NsHandle
	var InsideHandle *netlink.Handle
	var OutsideHandle *netlink.Handle
	OutsideNS, err = netns.Get()
	if err != nil {
		return nil, nil, netns.None(), netns.None(), err
	}
	InsideNS, err = netns.GetFromName(name)
	if err != nil {
		return nil, nil, netns.None(), netns.None(), err
	}
	OutsideHandle, err = netlink.NewHandle()
	if err != nil {
		return nil, nil, netns.None(), netns.None(), err
	}
	InsideHandle, err = netlink.NewHandleAt(InsideNS)
	if err != nil {
		return nil, nil, netns.None(), netns.None(), err
	}
	return OutsideHandle, InsideHandle, OutsideNS, InsideNS, nil
}
