package utils

import "github.com/vishvananda/netlink"

func LinkListDiff(current, target []netlink.Link) []netlink.Link {
	t := map[string]bool{}
	for _, l := range target {
		t[l.Attrs().Name] = true
	}
	var diff []netlink.Link
	for _, l := range current {
		if _, ok := t[l.Attrs().Name]; !ok {
			diff = append(diff, l)
		}
	}
	return diff
}