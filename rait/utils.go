package rait

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"os"
	"os/exec"
	"strings"
)

var _, IP4NetAll, _ = net.ParseCIDR("0.0.0.0/0")
var _, IP6NetAll, _ = net.ParseCIDR("::/0")

func ConnectStdIO(cmd *exec.Cmd) *exec.Cmd {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd
}

func RandomLinklocal() *netlink.Addr {
	a, _ := netlink.ParseAddr("fe80::fc54:ff:fe37:87aa/64")
	return a
}

func DestroyLinks(IFPrefix string) error {
	handle, err := netlink.NewHandle()
	if err != nil {
		return fmt.Errorf("failed to get netlink handle: %w", err)
	}
	defer handle.Delete()
	links, err := handle.LinkList()
	if err != nil {
		return fmt.Errorf("failed to list links: %w", err)
	}
	for _, link := range links {
		if strings.HasPrefix(link.Attrs().Name, IFPrefix) {
			err = handle.LinkDel(link)
			if err != nil {
				return fmt.Errorf("failed to delete link: %w", err)
			}
		}
	}
	return nil
}