package rait

import (
	"fmt"
	"net"

	"github.com/Catofes/RAIT/pkg/isolation"
	"github.com/Catofes/RAIT/pkg/misc"
	"github.com/Catofes/netlink"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func (r *RAIT) List() ([]misc.Link, error) {
	iso, err := isolation.NewIsolation(r.Isolation.IFGroup, r.Isolation.Transit, r.Isolation.Target)
	if err != nil {
		return nil, err
	}
	return iso.LinkList()
}

func (r *RAIT) Load() ([]misc.Link, error) {
	privateKeys := make([]wgtypes.Key, 0)
	for _, t := range r.Transport {
		privateKey, err := wgtypes.ParseKey(t.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %s", err)
		}
		privateKeys = append(privateKeys, privateKey)
	}

	peers, err := NewPeers(r.Peers, privateKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to load peers: %s", err)
	}
	var links []misc.Link
	for _, t := range r.Transport {
		transport := t
		transport.AddressFamily = misc.NewAF(transport.AddressFamily)
		privKey, _ := wgtypes.ParseKey(transport.PrivateKey)

		if transport.InnerAddress == "" {
			transport.InnerAddress = misc.NewLLAddrFromKey(privKey.PublicKey().String() + transport.AddressFamily + "wireguard").String()
		}

		innerIP, _, err := net.ParseCIDR(transport.InnerAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to parse inner address in %s : %s", transport.InnerAddress, err)
		}

		wgPeers := make([]wgtypes.PeerConfig, 0)
		fdb := make([]netlink.Neigh, 0)

		for _, p := range peers {
			peer := p
			pubKey, err := wgtypes.ParseKey(peer.PublicKey)
			if err != nil {
				zap.S().Warnf("failed to parse peer public key: %s, ignoring peer", err)
				continue
			}
			endpoint := peer.Endpoint
			if transport.AddressFamily != misc.NewAF(endpoint.AddressFamily) {
				continue
			}
			resolved, err := net.ResolveIPAddr(transport.AddressFamily, endpoint.Address)
			var wgEndpoint *net.UDPAddr
			if err != nil || resolved.IP == nil {
				zap.S().Debugf("peer address %s resolve failed in address family %s", endpoint.Address, transport.AddressFamily)
			} else {
				zap.S().Debugf("peer address %s resolved as %s in address family %s",
					endpoint.Address, resolved.IP, transport.AddressFamily)
				wgEndpoint = &net.UDPAddr{
					IP:   resolved.IP,
					Port: endpoint.Port,
				}
			}
			peer.GenerateInnerAddress()
			peerInnerAddress, _, err := net.ParseCIDR(peer.Endpoint.InnerAddress)
			var allowedIPs net.IPNet
			if peerInnerAddress.To4() == nil {
				allowedIPs = net.IPNet{
					IP:   peerInnerAddress,
					Mask: net.CIDRMask(128, 128),
				}
			}
			if err != nil {
				zap.S().Debugf("peer %s parse inner address failed: %s, %s, ignore peer", endpoint.Address, endpoint.InnerAddress, err)
				continue
			}
			p := wgtypes.PeerConfig{
				PublicKey:         pubKey,
				Remove:            false,
				UpdateOnly:        false,
				PresharedKey:      nil,
				Endpoint:          wgEndpoint,
				ReplaceAllowedIPs: true,
				AllowedIPs:        []net.IPNet{allowedIPs},
			}
			wgPeers = append(wgPeers, p)
			n := netlink.Neigh{
				Family:       unix.AF_BRIDGE,
				IP:           peerInnerAddress,
				HardwareAddr: peer.GenerateMac(),
				Flags:        netlink.NTF_SELF,
				State:        netlink.NUD_PERMANENT,
			}
			fdb = append(fdb, n)
			n = netlink.Neigh{
				Family:       unix.AF_BRIDGE,
				IP:           peerInnerAddress,
				HardwareAddr: net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				Flags:        netlink.NTF_SELF,
				State:        netlink.NUD_PERMANENT,
			}
			fdb = append(fdb, n)
		}
		port := transport.Port
		link := misc.Link{
			Name: transport.IFPrefix + "wg",
			Type: "wireguard",
			MTU:  transport.MTU,
			Config: wgtypes.Config{
				PrivateKey:   &privKey,
				ListenPort:   &port,
				BindAddress:  misc.ResolveBindAddress(transport.AddressFamily, transport.BindAddress),
				FirewallMark: &transport.FwMark,
				ReplacePeers: true,
				Peers:        wgPeers,
			},
			Address: transport.InnerAddress,
		}
		if transport.Mac == "" {
			transport.Mac = misc.NewMacFromKey(privKey.PublicKey().String() + transport.AddressFamily).String()
		}
		zap.S().Debugf("local mac: %s from %s", transport.Mac, privKey.PublicKey().String()+transport.AddressFamily)
		vxlink := misc.Link{
			Name:    transport.IFPrefix + "vxlan",
			Type:    "vxlan",
			MTU:     transport.MTU - 70,
			Mac:     transport.Mac,
			Address: innerIP.String(),
			VNI:     transport.VNI,
			FDB:     fdb,
		}
		links = append(links, link, vxlink)

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
		if link.Type == "wireguard" {
			err = iso.LinkEnsure(link)
			if err != nil {
				zap.S().Warnf("failed to ensure wireguard link %s: %s, skipping", link.Name, err)
				continue
			}
			targetLinkList = append(targetLinkList, link)
		}
	}

	for _, link := range links {
		if link.Type == "vxlan" {
			err = iso.LinkEnsure(link)
			if err != nil {
				zap.S().Warnf("failed to ensure vxlan link %s: %s, skipping", link.Name, err)
				continue
			}
			targetLinkList = append(targetLinkList, link)
		}
	}

	currentLinkList, err := iso.LinkList()
	if err != nil {
		return err
	}

	for _, link := range currentLinkList {
		if link.Type == "vxlan" && !misc.LinkIn(targetLinkList, link) {
			err = iso.LinkAbsent(link)
			if err != nil {
				zap.S().Warnf("failed to remove link %s: %s, skipping", link.Name, err)
				continue
			}
		}
	}

	for _, link := range currentLinkList {
		if link.Type == "wireguard" && !misc.LinkIn(targetLinkList, link) {
			err = iso.LinkAbsent(link)
			if err != nil {
				zap.S().Warnf("failed to remove link %s: %s, skipping", link.Name, err)
				continue
			}
		}
	}
	return nil
}
