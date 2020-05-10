package consts

import "net"

var _, ip4NetAll, _ = net.ParseCIDR("0.0.0.0/0")
var _, ip6NetAll, _ = net.ParseCIDR("::/0")
var IPNetAll = []net.IPNet{*ip4NetAll, *ip6NetAll}
