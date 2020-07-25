package rait

import (
	"fmt"
	"gitlab.com/NickCao/RAIT/v3/pkg/isolation"
	"gitlab.com/NickCao/RAIT/v3/pkg/misc"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"strconv"
)

func (r *RAIT) List() ([]misc.Link, error) {
	iso, err := isolation.NewIsolation(r.Isolation.IFGroup, r.Isolation.Transit, r.Isolation.Target)
	if err != nil {
		return nil, err
	}
	return iso.LinkList()
}

func (r *RAIT) Load() ([]misc.Link, error) {
	privKey, err := wgtypes.ParseKey(r.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %s", err)
	}

	peers, err := NewPeers(r.Peers, privKey.PublicKey().String())
	if err != nil {
		return nil, fmt.Errorf("failed to load peers: %s", err)
	}

	var links []misc.Link
	for _, t := range r.Transport {
		transport := t
		transport.AddressFamily = misc.NewAF(transport.AddressFamily)
		for _, p := range peers {
			peer := p
			pubKey, err := wgtypes.ParseKey(peer.PublicKey)
			if err != nil {
				zap.S().Warnf("failed to parse peer public key: %s, ignoring peer", err)
				continue
			}
			for _, e := range peer.Endpoint {
				endpoint := e
				if transport.AddressFamily != misc.NewAF(endpoint.AddressFamily) {
					continue
				}
				var address net.IP
				resolved, err := net.ResolveIPAddr(transport.AddressFamily, endpoint.Address)
				if err != nil || resolved.IP == nil {
					zap.S().Debugf("peer address %s resolve failed in address family %s, falling back to localhost",
						endpoint.Address, transport.AddressFamily)
					switch transport.AddressFamily {
					case "ip4":
						address = net.ParseIP("127.0.0.1")
					default:
						address = net.ParseIP("::1")
					}
				} else {
					zap.S().Debugf("peer address %s resolved as %s in address family %s",
						endpoint.Address, resolved.IP, transport.AddressFamily)
					address = resolved.IP
				}

				var listenPort *int
				if transport.RandomPort {
					listenPort = nil
				} else {
					listenPort = &endpoint.SendPort
				}

				link := misc.Link{
					Name: transport.IFPrefix + strconv.Itoa(endpoint.SendPort),
					MTU:  transport.MTU,
					Config: wgtypes.Config{
						PrivateKey:   &privKey,
						ListenPort:   listenPort,
						BindAddress:  misc.ResolveBindAddress(transport.AddressFamily, transport.BindAddress),
						FirewallMark: &transport.FwMark,
						ReplacePeers: true,
						Peers: []wgtypes.PeerConfig{
							{
								PublicKey:    pubKey,
								Remove:       false,
								UpdateOnly:   false,
								PresharedKey: nil,
								Endpoint: &net.UDPAddr{
									IP:   address,
									Port: transport.SendPort,
								},
								ReplaceAllowedIPs: true,
								AllowedIPs:        misc.IPNetAll,
							},
						},
					},
				}
				links = append(links, link)
			}
		}
	}
	return links, nil
}

func (r *RAIT) Sync(up bool) error {
	var links []misc.Link
	var err error
	if up {
		links, err = r.Load()
		if err != nil {
			return err
		}
	}

	iso, err := isolation.NewIsolation(r.Isolation.IFGroup, r.Isolation.Transit, r.Isolation.Target)
	if err != nil {
		return err
	}

	var targetLinkList []misc.Link
	for _, link := range links {
		err = iso.LinkEnsure(link)
		if err != nil {
			zap.S().Warnf("failed to ensure link %s: %s, skipping", link.Name, err)
			continue
		}
		targetLinkList = append(targetLinkList, link)
	}

	currentLinkList, err := iso.LinkList()
	if err != nil {
		return err
	}

	for _, link := range currentLinkList {
		if !misc.LinkIn(targetLinkList, link) {
			err = iso.LinkAbsent(link)
			if err != nil {
				zap.S().Warnf("failed to remove link %s: %s, skipping", link.Name, err)
				continue
			}
		}
	}
	return nil
}
