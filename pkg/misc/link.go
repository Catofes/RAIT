package misc

import "golang.zx2c4.com/wireguard/wgctrl/wgtypes"

// Link represents a single link managed by isolation
type Link struct {
	Name   string
	MTU    int
	Config wgtypes.Config
}

func LinkFilter(links []*Link, fn func(*Link) bool) (filtered []*Link) {
	for _, link := range links {
		if fn(link) {
			filtered = append(filtered, link)
		}
	}
	return
}

func LinkString(links []*Link) (stringed []string) {
	for _, link := range links {
		stringed = append(stringed, link.Name)
	}
	return
}

func LinkIn(links []*Link, link *Link) bool {
	for _, item := range links {
		if item.Name == link.Name {
			return true
		}
	}
	return false
}
