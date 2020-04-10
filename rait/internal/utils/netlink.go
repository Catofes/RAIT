package utils

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// NetlinkHelper handles the lifecycle of NetlinkHandle and NamespaceHandle
type NetlinkHelper struct {
	NetlinkHandle   *netlink.Handle
	NamespaceHandle netns.NsHandle
}

func (helper *NetlinkHelper) Destroy() error {
	var err error
	helper.NetlinkHandle.Delete()
	err = helper.NamespaceHandle.Close()
	if err != nil {
		return err
	}
	return nil
}

// NetlinkHelperFromName creates a NetlinkHelper from the specified netns, or "current" from the original netns
func NetlinkHelperFromName(name string) (*NetlinkHelper, error) {
	var helper NetlinkHelper
	var err error
	switch name {
	case "current":
		helper.NamespaceHandle, err = netns.Get()
		if err != nil {
			return nil, fmt.Errorf("failed to get namespace handle: %w", err)
		}
		helper.NetlinkHandle, err = netlink.NewHandle()
		if err != nil {
			return nil, fmt.Errorf("failed to get netlink handle: %w", err)
		}
	default:
		helper.NamespaceHandle, err = netns.GetFromName(name)
		if err != nil {
			return nil, fmt.Errorf("failed to get namespace handle: %w", err)
		}
		helper.NetlinkHandle, err = netlink.NewHandleAt(helper.NamespaceHandle)
		if err != nil {
			return nil, fmt.Errorf("failed to get netlink handle: %w", err)
		}
	}
	return &helper, nil
}
