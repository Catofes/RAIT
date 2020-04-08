package consts

import "net"

var _, IP4NetAll, _ = net.ParseCIDR("0.0.0.0/0")
var _, IP6NetAll, _ = net.ParseCIDR("::/0")
