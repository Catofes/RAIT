package rait

import (
	"encoding/json"
	"fmt"
	"github.com/vishvananda/netlink"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

var _, IP4NetAll, _ = net.ParseCIDR("0.0.0.0/0")
var _, IP6NetAll, _ = net.ParseCIDR("::/0")

func ConnectStdIO(cmd * exec.Cmd) *exec.Cmd{
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd
}

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

func DestroyLinks(ifprefix string) error {
	var handle *netlink.Handle
	var err error
	handle, err = netlink.NewHandle()
	if err != nil {
		return fmt.Errorf("failed to get netlink handle: %w", err)
	}
	defer handle.Delete()
	var links []netlink.Link
	links, err = handle.LinkList()
	if err != nil {
		return fmt.Errorf("failed to list links: %w", err)
	}
	for _, link := range links {
		if strings.HasPrefix(link.Attrs().Name, ifprefix) && link.Type() == "wireguard" {
			err = handle.LinkDel(link)
			if err != nil {
				return fmt.Errorf("failed to delete link: %w", err)
			}
		}
	}
	return nil
}

type JSONConfig struct {
	RAIT  RAITConfig   `json:"rait"`
	Peers []PeerConfig `json:"peers"`
}

func LoadFromJSON(data []byte) (*RAIT, []*Peer, error) {
	var config JSONConfig
	var err error
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode json: %w", err)
	}
	var peers []*Peer
	for _, peerconfig := range config.Peers {
		peer, err := NewPeer(&peerconfig)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load peer: %w", err)
		}
		peers = append(peers, peer)
	}
	var rait *RAIT
	rait, err = NewRAIT(&config.RAIT)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load rait: %w", err)
	}
	return rait, peers, nil
}
