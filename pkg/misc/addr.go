package misc

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/Catofes/netlink"
	"go.uber.org/zap"
)

var Bind bool

var _4b = [4]byte{}
var _16b = [16]byte{}

// IPNetALL is simply 0/0 plus ::/0
var IPNetAll = []net.IPNet{{IP: _4b[:], Mask: _4b[:]}, {IP: _16b[:], Mask: _16b[:]}}

// NewLLAddr generates a RFC 4862 compliant linklocal address, from a random mac address
func NewLLAddr() *netlink.Addr {
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

func NewLLAddrFromMac(mac net.HardwareAddr) *netlink.Addr {
	if len(mac) != 6 {
		return NewLLAddr()
	}
	digits := make([]byte, 8)
	copy(digits[:], mac[:])
	copy(digits[5:], digits[3:])
	digits[3] = 0xff
	digits[4] = 0xfe
	digits[0] = digits[0] ^ 2
	//var parts string
	part := hex.EncodeToString(digits[:])
	addrString := fmt.Sprintf("fe80::%s:%s:%s:%s/64", part[0:4], part[4:8], part[8:12], part[12:16])
	// for i := 0; i < 8; i += 2 {
	// 	parts += ":"
	// 	parts += fmt.Sprintf("%x", digits[i])
	// 	parts += fmt.Sprintf("%x", digits[i+1])
	// }
	//addr, _ := netlink.ParseAddr("fe80:" + parts + "/64")
	addr, _ := netlink.ParseAddr(addrString)
	return addr
}

func NewMacFromKey(key string) net.HardwareAddr {
	hash := md5.Sum([]byte(key))
	digits := []byte{hash[0], hash[1], hash[2], hash[3], hash[4], hash[5]}
	digits[0] = digits[0] & 0b11111110
	return digits
}

func NewLLAddrFromKey(key string) *netlink.Addr {
	return NewLLAddrFromMac(NewMacFromKey(key))
}

func NewAF(af string) string {
	switch af {
	case "ip4", "ip6":
		return af
	case "":
		return "ip4"
	default:
		zap.S().Warnf("unrecognized address family %s, falling back to ip4", af)
		return "ip4"
	}
}

func ResolveBindAddress(af string, addrSpec string) net.IP {
	if !Bind {
		return nil
	}

	if addrSpec == "" {
		switch af {
		case "ip4":
			return net.ParseIP("0.0.0.0")
		case "ip6":
			return net.ParseIP("::")
		}
	}
	return net.ParseIP(addrSpec)
}
