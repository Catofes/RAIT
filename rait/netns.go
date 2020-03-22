package rait

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type NamespaceHelper struct {
	SrcHandle    *netlink.Handle
	DstHandle    *netlink.Handle
	DstNamespace netns.NsHandle
}

func (h *NamespaceHelper) Destroy() {
	h.SrcHandle.Delete()
	h.DstHandle.Delete()
}

func NamespaceHelperFromName(name string) (*NamespaceHelper, error) {
	_ = CreateNamedNamespace(name) // It won't hurt
	var h NamespaceHelper
	var err error
	h.SrcHandle, err = netlink.NewHandle()
	if err != nil {
		return nil, fmt.Errorf("NamespaceHelperFromName: failed to get src ns handle: %w", err)
	}
	h.DstNamespace, err = netns.GetFromName(name)
	if err != nil {
		return nil, fmt.Errorf("NamespaceHelperFromName: failed to get dst ns: %w", err)
	}
	h.DstHandle, err = netlink.NewHandleAt(h.DstNamespace)
	if err != nil {
		return nil, fmt.Errorf("NamespaceHelperFromName: failed to get dst ns handle: %w", err)
	}
	return &h, nil
}
