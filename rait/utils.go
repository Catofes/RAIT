package rait

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"math/rand"
	"net"
	"strings"
)

var _, IP4NetAll, _ = net.ParseCIDR("0.0.0.0/0")
var _, IP6NetAll, _ = net.ParseCIDR("::/0")

func RandomLinklocal() *netlink.Addr {
	digits := []int{0x00, 0x16, 0x3e, rand.Intn(0x7f + 1), rand.Intn(0xff + 1), rand.Intn(0xff + 1)}
	digits = append(digits, 0, 0)
	copy(digits[5:], digits[3:])
	digits[3] = 0xff
	digits[4] = 0xfe
	digits[0] = digits[0] ^ 2
	var parts string
	for i := 0; i < len(digits); i += 2 {
		parts += ":"
		parts += fmt.Sprintf("%x", digits[i])
		parts += fmt.Sprintf("%x", digits[i+1])
	}
	a, _ := netlink.ParseAddr("fe80:"+parts+"/64")
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
		if strings.HasPrefix(link.Attrs().Name, IFPrefix) && link.Type() == "wireguard" {
			err = handle.LinkDel(link)
			if err != nil {
				return fmt.Errorf("failed to delete link: %w", err)
			}
		}
	}
	return nil
}
