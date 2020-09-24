package misc

import (
	"github.com/Catofes/netlink"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Link represents a single link managed by isolation
type Link struct {
	Name    string
	Type    string
	MTU     int
	Address string
	Mac     string
	VNI     int
	FDB     []netlink.Neigh
	Config  wgtypes.Config
}

func LinkString(links []Link) (stringed []string) {
	for _, link := range links {
		stringed = append(stringed, link.Name)
	}
	return
}

func LinkIn(links []Link, link Link) bool {
	return StringIn(LinkString(links), link.Name)
}

func StringIn(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}
