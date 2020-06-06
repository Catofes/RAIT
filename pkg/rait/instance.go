package rait

import (
	"github.com/vishvananda/netns"
	"gitlab.com/NickCao/RAIT/pkg/types"
	"gitlab.com/NickCao/RAIT/pkg/utils"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Instance represents the control structure of RAIT
type Instance struct {
	PrivateKey         wgtypes.Key    // mandatory, the private key of the client
	AddressFamily      string         // optional, default ip4, the address family of the client, ip4 or ip6
	SendPort           int            // mandatory, the sending port of the client
	InterfacePrefix    string         // optional, default rait, the common prefix to name the wireguard interfaces
	InterfaceGroup     int            // optional, default 54, the ifgroup for the wireguard interfaces
	InterfaceNamespace netns.NsHandle // optional, default current, the netns to move wireguard interface into
	TransitNamespace   netns.NsHandle // optional, default current, the netns to create wireguard sockets in
	MTU                int            // optional, default 1400, the MTU of the wireguard interfaces
	FwMark             int            // optional, default 54, the fwmark on packets sent by wireguard
	Peers              string         // optional, default /etc/rait/peers.conf, the url of the peer list
}

func InstanceFromMap(data map[string]string) (*Instance, error) {
	var instance Instance
	var err error
	instance.PrivateKey, err = wgtypes.ParseKey(data["PrivateKey"])
	if err != nil {
		return nil, NewErrDecode("Instance", "PrivateKey", err)
	}
	instance.AddressFamily, err = types.ParseAddressFamily(types.OrDefault(data["AddressFamily"], "ip4"))
	if err != nil {
		return nil, NewErrDecode("Instance", "AddressFamily", err)
	}
	instance.SendPort, err = types.ParseUint16(data["SendPort"])
	if err != nil {
		return nil, NewErrDecode("Instance", "SendPort", err)
	}
	instance.InterfacePrefix = types.OrDefault(data["InterfacePrefix"], "rait")
	instance.InterfaceGroup, err = types.ParseUint16(types.OrDefault(data["InterfaceGroup"], "54"))
	if err != nil {
		return nil, NewErrDecode("Instance", "InterfaceGroup", err)
	}
	instance.InterfaceNamespace, err = types.ParseNamespace(data["InterfaceNamespace"])
	if err != nil {
		return nil, NewErrDecode("Instance", "InterfaceNamespace", err)
	}
	instance.TransitNamespace, err = types.ParseNamespace(data["TransitNamespace"])
	if err != nil {
		return nil, NewErrDecode("Instance", "TransitNamespace", err)
	}
	instance.MTU, err = types.ParseUint16(types.OrDefault(data["MTU"], "1400"))
	if err != nil {
		return nil, NewErrDecode("Instance", "MTU", err)
	}
	instance.FwMark, err = types.ParseUint16(types.OrDefault(data["FwMark"], "54"))
	if err != nil {
		return nil, NewErrDecode("Instance", "FwMark", err)
	}
	instance.Peers = types.OrDefault(data["Peers"], "/etc/rait/peers.conf")
	return &instance, nil
}

func InstanceFromPath(path string) (*Instance, error) {
	var instanceMap map[string]string
	err := utils.DecodeTOMLFromPath(path, &instanceMap)
	if err != nil {
		return nil, err
	}
	return InstanceFromMap(instanceMap)
}
