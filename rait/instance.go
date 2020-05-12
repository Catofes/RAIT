package rait

// Instance represents the control structure of RAIT
type Instance struct {
	PrivateKey         string // mandatory, the private key of the client
	AddressFamily      string // the address family of the client, ip4 or ip6
	SendPort           int    // mandatory, the sending port of the client
	InterfacePrefix    string // the common prefix to name the wireguard interfaces
	InterfaceGroup     int    // the ifgroup for the wireguard interfaces
	InterfaceNamespace string // the netns to move wireguard interface into
	TransitNamespace   string // the netns to create wireguard sockets in
	MTU                int    // the MTU of the wireguard interfaces
	FwMark             int    // the fwmark on packets sent by wireguard
	Peers              string // the url of the peer list
}

// DefaultInstance returns a sane default for most fields in Instance
func DefaultInstance() *Instance {
	return &Instance{
		AddressFamily:   "ip4",
		InterfacePrefix: "rait",
		InterfaceGroup:  0x36,
		MTU:             1400,
		FwMark:          0x36,
		Peers:           "/etc/rait/peers.conf",
	}
}
