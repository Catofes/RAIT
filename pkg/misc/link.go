package misc

import (
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Link represents a single link managed by isolation
type Link struct {
	Name   string
	MTU    int
	Config wgtypes.Config
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

