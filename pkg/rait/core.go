package rait

import (
	"fmt"
	"gitlab.com/NickCao/RAIT/v2/pkg/isolation"
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"strconv"
)

func (r *RAIT) List() ([]misc.Link, error) {
	iso, err := isolation.NewIsolation(r.Isolation.Type, r.Isolation.IFGroup, r.Isolation.Transit, r.Isolation.Target)
	if err != nil {
		return nil, err
	}

	links, err := iso.LinkList()
	if err != nil {
		return nil, err
	}
	return links, nil
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
	for _, transport := range r.Transport {
		transport.AddressFamily = misc.NewAF(transport.AddressFamily)
		for _, peer := range peers {
			if transport.AddressFamily != misc.NewAF(peer.AddressFamily) {
				continue
			}

			pubKey, err := wgtypes.ParseKey(peer.PublicKey)
			if err != nil {
				zap.S().Warnf("failed to parse peer public key: %s", err)
				continue
			}

			var endpoint net.IP
			resolved, err := net.ResolveIPAddr(transport.AddressFamily, peer.Endpoint)
			if err != nil || resolved.IP == nil {
				zap.S().Debugf("peer endpoint %s resolve failed in address family %s, falling back to localhost",
					peer.Endpoint, transport.AddressFamily)
				switch transport.AddressFamily {
				case "ip4":
					endpoint = net.ParseIP("127.0.0.1")
				default:
					endpoint = net.ParseIP("::1")
				}
			} else {
				zap.S().Debugf("peer endpoint %s resolved as %s in address family %s",
					peer.Endpoint, resolved.IP, transport.AddressFamily)
				endpoint = resolved.IP
			}

			tmpPort := peer.SendPort
			listenPort := &tmpPort
			if transport.DynamicListenPort {
				listenPort = nil
			}

			links = append(links, misc.Link{
				Name: transport.IFPrefix + strconv.Itoa(peer.SendPort),
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
								IP:   endpoint,
								Port: transport.SendPort,
							},
							ReplaceAllowedIPs: true,
							AllowedIPs:        misc.IPNetAll,
						},
					},
				},
			})
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

	iso, err := isolation.NewIsolation(r.Isolation.Type, r.Isolation.IFGroup, r.Isolation.Transit, r.Isolation.Target)
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
				return err
			}
		}
	}
	return nil
}
