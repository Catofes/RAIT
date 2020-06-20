package rait

import (
	"gitlab.com/NickCao/RAIT/v2/pkg/isolation"
	"gitlab.com/NickCao/RAIT/v2/pkg/types"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"strconv"
)

func (instance *Instance) LoadPeers() ([]*Peer, error) {
	peers, err := PeersFromPath(instance.Peers)
	if err != nil {
		return nil, err
	}
	key, err := wgtypes.ParseKey(instance.PrivateKey)
	if err != nil {
		return nil, err
	}

	n := 0
	for _, x := range peers {
		if x.AddressFamily == instance.AddressFamily &&
			x.PublicKey != key.PublicKey().String() {
			peers[n] = x
			n++
		}
	}
	peers = peers[:n]
	return peers, nil
}

func (instance *Instance) WireguardConfig(peer *Peer) (string, *wgtypes.Config, error) {
	privKey, err := wgtypes.ParseKey(instance.PrivateKey)
	if err != nil {
		return "", nil, err
	}
	pubKey, err := wgtypes.ParseKey(peer.PublicKey)
	if err != nil {
		return "", nil, err
	}
	var endpoint net.IP
	resolved, err := net.ResolveIPAddr(instance.AddressFamily, peer.Endpoint)
	if err != nil || resolved.IP == nil {
		switch instance.AddressFamily {
		case "ip4":
			endpoint = net.ParseIP("127.0.0.1")
		case "ip6":
			endpoint = net.ParseIP("::1")
		}
	} else {
		endpoint = resolved.IP
	}
	return instance.InterfacePrefix + strconv.Itoa(peer.SendPort), &wgtypes.Config{
		PrivateKey:   &privKey,
		ListenPort:   &peer.SendPort,
		FirewallMark: &instance.FwMark,
		ReplacePeers: true,
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:    pubKey,
				Remove:       false,
				UpdateOnly:   false,
				PresharedKey: nil,
				Endpoint: &net.UDPAddr{
					IP:   endpoint,
					Port: instance.SendPort,
				},
				ReplaceAllowedIPs: true,
				AllowedIPs:        types.IPNetAll,
			},
		},
	}, nil
}

func (instance *Instance) SyncInterfaces(up bool) error {
	var peers []*Peer
	var err error
	if up {
		peers, err = instance.LoadPeers()
		if err != nil {
			return err
		}
	}

	gi, err := isolation.NewGenericIsolation(instance.Isolation, instance.TransitNamespace, instance.InterfaceNamespace)
	if err != nil {
		return err
	}

	var targetLinkList []string
	for _, peer := range peers {
		name, config, err := instance.WireguardConfig(peer)
		if err != nil {
			return err
		}
		err = gi.LinkEnsure(name, instance.MTU, instance.InterfaceGroup, *config)
		if err != nil {
			return err
		}
		targetLinkList = append(targetLinkList, name)
	}

	err = gi.LinkConstrain(targetLinkList, instance.InterfacePrefix, instance.InterfaceGroup)
	if err != nil {
		return err
	}
	return nil
}
