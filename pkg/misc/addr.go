package misc

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"math/rand"
	"net"
	"time"
)

var IPNetAll = []net.IPNet{{IP: net.IP{0, 0, 0, 0}, Mask: net.IPMask{0, 0, 0, 0}},
	{IP: net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Mask: net.IPMask{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}}

// LinkLocalAddr generates a RFC 4862 compliant linklocal address, from a random mac address
func LinkLocalAddr() *netlink.Addr {
	rand.Seed(time.Now().UnixNano())
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
	addr, _ := netlink.ParseAddr("fe80:" + parts + "/64")
	return addr
}
