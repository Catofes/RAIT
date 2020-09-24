package netns

import (
	"fmt"
	"net"
	"runtime"
	"strings"
	"sync"

	"github.com/Catofes/RAIT/pkg/misc"
	"github.com/Catofes/netlink"
	"github.com/vishvananda/netns"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/wgctrl"
)

// NetnsIsolation is the recommended implementation as by the wireguard developers
// It keeps the wireguard sockets and interfaces in different netns to facilitate isolation
type NetnsIsolation struct {
	group   int
	transit string
	target  string
}

// NewNetnsIsolation takes two arguments: transit and interface namespace
// the creation of netns is handled internally
// the links and sockets will be created in the transit namespace
// and the links will be moved into the interface namespace
func NewNetnsIsolation(group int, transit, target string) (*NetnsIsolation, error) {
	return &NetnsIsolation{
		group:   group,
		transit: transit,
		target:  target,
	}, nil
}

func (i *NetnsIsolation) create(attrs misc.Link, h *netlink.Handle, ns netns.NsHandle, t *netlink.Handle) error {
	switch attrs.Type {
	case "wireguard":
		link := &netlink.Wireguard{
			LinkAttrs: netlink.LinkAttrs{
				Name:  attrs.Name,
				MTU:   attrs.MTU,
				Group: uint32(i.group)}}

		if err := t.LinkAdd(link); err != nil {
			return fmt.Errorf("failed to create link %s: %s", attrs.Name, err)
		}
		zap.S().Debugf("link %s created", attrs.Name)

		if ns != 0 {
			if err := t.LinkSetNsFd(link, int(ns)); err != nil {
				_ = t.LinkDel(link)
				return fmt.Errorf("failed to move link %s: %s", attrs.Name, err)
			}
			zap.S().Debugf("link %s moved into target namespace", attrs.Name)
		}

		if err := i.updateWireguardIP(attrs, h, ns, t); err != nil {
			return err
		}

		if err := i.updateWireguardConf(attrs, h, ns, t); err != nil {
			return err
		}
		return nil

	case "vxlan":
		mac, _ := net.ParseMAC(attrs.Mac)
		src := net.ParseIP(attrs.Address)
		parent := 0
		if src.To4() == nil && src[0] == 0xfe && src[1] == 0x80 {
			link, err := h.LinkByName(strings.TrimRight(attrs.Name, "vxlan") + "wg")
			if err != nil {
				zap.S().Debugf("find parent for %s failed, err", attrs.Name, err)
			}
			parent = link.Attrs().Index
		}
		link := &netlink.Vxlan{
			LinkAttrs: netlink.LinkAttrs{
				Name:         attrs.Name,
				MTU:          attrs.MTU,
				Group:        uint32(i.group),
				HardwareAddr: mac,
				NetNsID:      int(ns)},
			Learning:     false,
			VxlanId:      attrs.VNI,
			SrcAddr:      src,
			VtepDevIndex: parent,
		}
		if err := h.LinkAdd(link); err != nil {
			return fmt.Errorf("failed to create vxlan %s: %s", attrs.Name, err)
		}
		zap.S().Debugf("link %s created", attrs.Name)
		return nil
	}
	return nil
}

func (i *NetnsIsolation) updateVXLANMac(attrs misc.Link, h *netlink.Handle, ns netns.NsHandle, t *netlink.Handle) error {
	link, _ := h.LinkByName(attrs.Name)
	if link.Attrs().HardwareAddr.String() != attrs.Mac {
		mac, err := net.ParseMAC(attrs.Mac)
		if err != nil {
			return fmt.Errorf("failed to parse %s mac %s: %s", attrs.Name, attrs.Mac, err)
		}
		if err := h.LinkSetHardwareAddr(link, mac); err != nil {
			return fmt.Errorf("failed to set %s mac %s: %s", attrs.Name, attrs.Mac, err)
		}
	}
	return nil
}

func (i *NetnsIsolation) updateVXLANNeigh(attrs misc.Link, h *netlink.Handle, ns netns.NsHandle, t *netlink.Handle) error {
	link, _ := h.LinkByName(attrs.Name)
	for _, neigh := range attrs.FDB {
		neigh.LinkIndex = link.Attrs().Index
		if neigh.IP.To4() == nil && neigh.IP[0] == 0xfe && neigh.IP[1] == 0x80 {
			viaIf, err := h.LinkByName(strings.TrimRight(attrs.Name, "vxlan") + "wg")
			if err != nil {
				zap.S().Debugf("find fdb viaIf for %s failed, err", neigh.HardwareAddr, err)
			}
			neigh.ViaIfIndex = viaIf.Attrs().Index
		}
		if err := h.NeighAppend(&neigh); err != nil {
			zap.S().Warnf("neigh %s add failed, %s, ignore", neigh.IP, err)
			continue
		}
		zap.S().Debugf("neigh %s %s add succes at %s", neigh.HardwareAddr.String(), neigh.IP.String(), link.Attrs().Name)
	}
	return nil
}

func (i *NetnsIsolation) updateWireguardIP(attrs misc.Link, h *netlink.Handle, ns netns.NsHandle, t *netlink.Handle) error {
	var err error
	link, _ := h.LinkByName(attrs.Name)
	var addrs []netlink.Addr
	addrs, err = h.AddrList(link, unix.AF_INET|unix.AF_INET6)
	if err != nil {
		_ = h.LinkDel(link)
		return fmt.Errorf("failed to list addr on link %s: %s", attrs.Name, err)
	}
	flag := false
	innerAddress, _, _ := net.ParseCIDR(attrs.Address)
	for _, addr := range addrs {
		if addr.IP.Equal(innerAddress) {
			flag = true
			break
		}
	}
	if !flag {
		zap.S().Debugf(attrs.Address)
		k, _ := netlink.ParseAddr(attrs.Address)
		if err = h.AddrAdd(link, k); err != nil {
			_ = h.LinkDel(link)
			return fmt.Errorf("failed to add addr to link %s: %s", attrs.Name, err)
		}
		zap.S().Debugf("link %s inner address configured", attrs.Name)
	}
	return nil
}

func (i *NetnsIsolation) update(attrs misc.Link, h *netlink.Handle, ns netns.NsHandle, t *netlink.Handle) error {
	link, _ := h.LinkByName(attrs.Name)
	if attrs.MTU != 0 {
		if err := h.LinkSetMTU(link, attrs.MTU); err != nil {
			_ = h.LinkDel(link)
			return fmt.Errorf("failed to set mtu on link %s: %s", attrs.Name, err)
		}
	}
	zap.S().Debugf("link %s mtu set to %d", attrs.Name, attrs.MTU)

	if err := h.LinkSetGroup(link, i.group); err != nil {
		_ = h.LinkDel(link)
		return fmt.Errorf("failed to set group on link %s: %s", attrs.Name, err)
	}
	zap.S().Debugf("link %s ifgroup set to %d", attrs.Name, i.group)

	if err := h.LinkSetUp(link); err != nil {
		_ = h.LinkDel(link)
		return fmt.Errorf("failed to set up link %s: %s", attrs.Name, err)
	}
	zap.S().Debugf("link %s set up", attrs.Name)

	switch link.Type() {
	case "wireguard":
		if err := i.updateWireguardIP(attrs, h, ns, t); err != nil {
			return err
		}
		if err := i.updateWireguardConf(attrs, h, ns, t); err != nil {
			return err
		}
	case "vxlan":
		if err := i.updateVXLANNeigh(attrs, h, ns, t); err != nil {
			return err
		}
		if err := i.updateVXLANMac(attrs, h, ns, t); err != nil {
			return err
		}
	}
	zap.S().Debugf("link %s ready", attrs.Name)
	return nil
}

func (i *NetnsIsolation) delete(attrs misc.Link, h *netlink.Handle, ns netns.NsHandle, t *netlink.Handle) error {
	link, _ := h.LinkByName(attrs.Name)
	if err := h.LinkDel(link); err != nil {
		return fmt.Errorf("failed to remove link %s: %s", attrs.Name, err)
	}
	return nil
}

func (i *NetnsIsolation) updateWireguardConf(attrs misc.Link, h *netlink.Handle, ns netns.NsHandle, t *netlink.Handle) error {
	link, _ := h.LinkByName(attrs.Name)
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		runtime.LockOSThread()
		err := netns.Set(ns)
		if err != nil {
			_ = h.LinkDel(link)
			err = fmt.Errorf("failed to move into netns: %s", err)
			return
		}

		var wg *wgctrl.Client
		wg, err = wgctrl.New()
		if err != nil {
			_ = h.LinkDel(link)
			err = fmt.Errorf("failed to get wireguard control socket: %s", err)
			return
		}
		defer wg.Close()

		err = wg.ConfigureDevice(attrs.Name, attrs.Config)
		if err != nil {
			_ = h.LinkDel(link)
			err = fmt.Errorf("failed to configure wireguard interface %s: %s", attrs.Name, err)
			return
		}
		zap.S().Debugf("link %s wireguard configuration set", attrs.Name)
	}()
	waitGroup.Wait()
	return nil
}

func (i *NetnsIsolation) LinkEnsure(attrs misc.Link) (err error) {
	targetHandle, err := NewNetlink(i.target)
	if err != nil {
		return err
	}
	defer targetHandle.Delete()

	targetNetns, err := NewNetns(i.target)
	if err != nil {
		return err
	}
	defer targetNetns.Close()

	transitHandle, err := NewNetlink(i.transit)
	if err != nil {
		return err
	}
	defer transitHandle.Delete()

	link, err := targetHandle.LinkByName(attrs.Name)
	if err == nil {
		if link.Type() == "wireguard" || link.Type() == "vxlan" {
			zap.S().Debugf("link %s already exists, skipping creation", attrs.Name)
			return i.update(attrs, targetHandle, targetNetns, transitHandle)
		} else {
			zap.S().Debugf("link %s already exists but is of wrong type, removing", attrs.Name)
			return i.delete(attrs, targetHandle, targetNetns, transitHandle)
		}
	} else if _, ok := err.(netlink.LinkNotFoundError); !ok {
		return fmt.Errorf("failed to get link %s: %s", attrs.Name, err)
	} else {
		err = i.create(attrs, targetHandle, targetNetns, transitHandle)
		if err != nil {
			return err
		}
		return i.update(attrs, targetHandle, targetNetns, transitHandle)
	}
}

func (i *NetnsIsolation) LinkAbsent(attrs misc.Link) error {
	targetHandle, err := NewNetlink(i.target)
	if err != nil {
		return err
	}
	defer targetHandle.Delete()

	link, err := targetHandle.LinkByName(attrs.Name)
	if err != nil {
		return fmt.Errorf("failed to get link %s: %s", attrs.Name, err)
	}

	err = targetHandle.LinkDel(link)
	if err != nil {
		return fmt.Errorf("failed to delete link %s: %s", attrs.Name, err)
	}
	zap.S().Debugf("link %s removed", attrs.Name)
	return nil
}

func (i *NetnsIsolation) LinkList() ([]misc.Link, error) {
	targetHandle, err := NewNetlink(i.target)
	if err != nil {
		return nil, err
	}
	defer targetHandle.Delete()

	rawList, err := targetHandle.LinkList()
	if err != nil {
		return nil, fmt.Errorf("failed to list link: %s", err)
	}

	var list []misc.Link
	for _, link := range rawList {
		if (link.Type() == "wireguard" || link.Type() == "vxlan") && int(link.Attrs().Group) == i.group {
			list = append(list, misc.Link{
				Name: link.Attrs().Name,
				Type: link.Type(),
				MTU:  link.Attrs().MTU,
			})
		}
	}
	return list, nil
}
