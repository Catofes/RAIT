package rait

import (
	"github.com/vishvananda/netns"
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
	"gitlab.com/NickCao/RAIT/v2/pkg/namespace"
	"gitlab.com/NickCao/RAIT/v2/pkg/types"
	"go.uber.org/zap"
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

func InstanceFromMap(data map[string]interface{}) (*Instance, error) {
	logger := zap.S().Named("rait.InstanceFromMap")
	var instance Instance
	var err error
	instance.PrivateKey, err = wgtypes.ParseKey(misc.Fallback(data["PrivateKey"], "").(string))
	if err != nil {
		logger.Errorf("failed to parse wireguard private key, error %s", err)
		return nil, err
	}
	instance.AddressFamily = types.ParseAddressFamily(data["AddressFamily"])
	instance.SendPort = types.ParseInt64(data["SendPort"], 0)
	instance.InterfacePrefix = misc.Fallback(data["InterfacePrefix"], "rait").(string)
	instance.InterfaceGroup = types.ParseInt64(data["InterfaceGroup"], 54)
	instance.InterfaceNamespace, err = namespace.GetFromName(misc.Fallback(data["InterfaceNamespace"], "").(string))
	if err != nil {
		logger.Errorf("failed to parse interface namespace, error %s", err)
		return nil, err
	}
	instance.TransitNamespace, err = namespace.GetFromName(misc.Fallback(data["TransitNamespace"], "").(string))
	if err != nil {
		logger.Errorf("failed to parse transit namespace, error %s", err)
		return nil, err
	}
	instance.MTU = types.ParseInt64(data["MTU"], 1400)
	instance.FwMark = types.ParseInt64(data["FwMark"], 0)
	instance.Peers = misc.Fallback(data["Peers"], "/etc/rait/peers.conf").(string)
	return &instance, nil
}

func InstanceFromPath(path string) (*Instance, error) {
	var instanceMap map[string]interface{}
	err := misc.DecodeTOMLFromPath(path, &instanceMap)
	if err != nil {
		return nil, err
	}
	return InstanceFromMap(instanceMap)
}
