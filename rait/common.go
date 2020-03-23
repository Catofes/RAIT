package rait

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"math/rand"
	"net"
	"os/exec"
	"time"
)

var _, IP4NetAll, _ = net.ParseCIDR("0.0.0.0/0")
var _, IP6NetAll, _ = net.ParseCIDR("::/0")

func RandomLinklocal() *netlink.Addr {
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
	a, _ := netlink.ParseAddr("fe80:" + parts + "/64")
	return a
}

func SynthesisWireguardConfig(r *RAIT, p *Peer) *wgtypes.Config {
	// Be aware!
	if r.PrivateKey.Key.PublicKey() == p.PublicKey.Key {
		return nil
	}
	listenPort := int(p.SendPort)
	fwMark := int(r.FwMark)
	var endpoint *net.UDPAddr
	if p.Endpoint != nil {
		endpoint = &net.UDPAddr{
			IP:   p.Endpoint,
			Port: int(r.SendPort),
		}
	}
	return &wgtypes.Config{
		PrivateKey:   &r.PrivateKey.Key,
		ListenPort:   &listenPort,
		FirewallMark: &fwMark,
		ReplacePeers: true,
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:         p.PublicKey.Key,
				Remove:            false,
				UpdateOnly:        false,
				PresharedKey:      nil,
				Endpoint:          endpoint,
				ReplaceAllowedIPs: true,
				AllowedIPs:        []net.IPNet{*IP4NetAll, *IP6NetAll},
			}},
	}
}

// TODO: further eliminate the iproute2 dependency
func CreateNamedNamespace(name string) error {
	return exec.Command("ip", "netns", "add", name).Run()
}

func DestroyNamedNamespace(name string) error {
	return exec.Command("ip", "netns", "delete", name).Run()
}

