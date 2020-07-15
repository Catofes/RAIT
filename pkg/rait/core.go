package rait

import (
	"fmt"
	"gitlab.com/NickCao/RAIT/v2/pkg/isolation"
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"strconv"
	"strings"
)

func (instance *Instance) IsManagedInterface(attrs *isolation.LinkAttrs) bool {
	return strings.HasPrefix(attrs.Name, instance.InterfacePrefix) && attrs.Group == instance.InterfaceGroup
}

func (instance *Instance) ListInterfaceName() ([]string, error) {
	iso, err := isolation.NewIsolation(instance.Isolation, instance.TransitNamespace, instance.InterfaceNamespace)
	if err != nil {
		return nil, err
	}

	unfiltered, err := iso.LinkList()
	if err != nil {
		return nil, err
	}

	return isolation.LinkString(isolation.LinkFilter(unfiltered, instance.IsManagedInterface)), nil
}

func (instance *Instance) LoadPeers() ([]*Peer, error) {
	peers, err := PeersFromPath(instance.Peers)
	if err != nil {
		return nil, err
	}

	key, err := wgtypes.ParseKey(instance.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse instance private key: %w", err)
	}

	// some in place array filtering magic
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

func (instance *Instance) InterfaceConfig(peer *Peer) (*isolation.LinkAttrs, *wgtypes.Config, error) {
	privKey, err := wgtypes.ParseKey(instance.PrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse instance private key: %w", err)
	}

	pubKey, err := wgtypes.ParseKey(peer.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse peer public key: %w", err)
	}

	var endpoint net.IP
	resolved, err := net.ResolveIPAddr(instance.AddressFamily, peer.Endpoint)
	if err != nil || resolved.IP == nil {
		zap.S().Debugf("peer endpoint %s resolve failed in address family %s, falling back to localhost", peer.Endpoint, instance.AddressFamily)
		switch instance.AddressFamily {
		case "ip4":
			endpoint = net.ParseIP("127.0.0.1")
		case "ip6":
			endpoint = net.ParseIP("::1")
		}
	} else {
		zap.S().Debugf("peer endpoint %s resolved as %s in address family %s", peer.Endpoint, resolved.IP, instance.AddressFamily)
		endpoint = resolved.IP
	}

	listenPort := &peer.SendPort
	if instance.DynamicListenPort {
		listenPort = nil
	}

	return &isolation.LinkAttrs{
			Name:  instance.InterfacePrefix + strconv.Itoa(peer.SendPort),
			MTU:   instance.MTU,
			Group: instance.InterfaceGroup,
		}, &wgtypes.Config{
			PrivateKey:   &privKey,
			ListenPort:   listenPort,
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
					AllowedIPs:        misc.IPNetAll,
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

	iso, err := isolation.NewIsolation(instance.Isolation, instance.TransitNamespace, instance.InterfaceNamespace)
	if err != nil {
		return err
	}

	var targetLinkList []*isolation.LinkAttrs
	for _, peer := range peers {
		attrs, config, err := instance.InterfaceConfig(peer)
		if err != nil {
			return err
		}
		err = iso.LinkEnsure(attrs, *config)
		if err != nil {
			return err
		}
		targetLinkList = append(targetLinkList, attrs)
	}

	unfiltered, err := iso.LinkList()
	if err != nil {
		return err
	}
	currentLinkList := isolation.LinkFilter(unfiltered, instance.IsManagedInterface)

	for _, link := range currentLinkList {
		if !isolation.LinkIn(targetLinkList, link) {
			err = iso.LinkAbsent(link)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
