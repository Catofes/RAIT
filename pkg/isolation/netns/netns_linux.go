package netns

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"gitlab.com/NickCao/RAIT/v3/pkg/misc"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/wgctrl"
	"runtime"
	"sync"
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
		if link.Type() == "wireguard" {
			zap.S().Debugf("link %s already exists, skipping creation", attrs.Name)
			goto created
		} else {
			zap.S().Debugf("link %s already exists but is of wrong type, removing", attrs.Name)
			goto removal
		}
	} else if _, ok := err.(netlink.LinkNotFoundError); !ok {
		return fmt.Errorf("failed to get link %s: %s", attrs.Name, err)
	} else {
		goto create
	}

removal:
	err = targetHandle.LinkDel(link)
	if err != nil {
		return fmt.Errorf("failed to remove link %s: %s", attrs.Name, err)
	}

create:
	link = &netlink.Wireguard{LinkAttrs: netlink.LinkAttrs{Name: attrs.Name, MTU: attrs.MTU, Group: uint32(i.group)}}
	err = transitHandle.LinkAdd(link)
	if err != nil {
		return fmt.Errorf("failed to create link %s: %s", attrs.Name, err)
	}
	zap.S().Debugf("link %s created in transit namespace", attrs.Name)

	err = transitHandle.LinkSetNsFd(link, int(targetNetns))
	if err != nil {
		_ = transitHandle.LinkDel(link)
		return fmt.Errorf("failed to move link %s: %s", attrs.Name, err)
	}
	zap.S().Debugf("link %s moved into target namespace", attrs.Name)

created:
	err = targetHandle.LinkSetMTU(link, attrs.MTU)
	if err != nil {
		_ = targetHandle.LinkDel(link)
		return fmt.Errorf("failed to set mtu on link %s: %s", attrs.Name, err)
	}
	zap.S().Debugf("link %s mtu set to %d", attrs.Name, attrs.MTU)

	err = targetHandle.LinkSetGroup(link, i.group)
	if err != nil {
		_ = targetHandle.LinkDel(link)
		return fmt.Errorf("failed to set group on link %s: %s", attrs.Name, err)
	}
	zap.S().Debugf("link %s ifgroup set to %d", attrs.Name, i.group)

	err = targetHandle.LinkSetUp(link)
	if err != nil {
		_ = targetHandle.LinkDel(link)
		return fmt.Errorf("failed to set up link %s: %s", attrs.Name, err)
	}
	zap.S().Debugf("link %s set up", attrs.Name)

	var addrs []netlink.Addr
	addrs, err = targetHandle.AddrList(link, unix.AF_INET6)
	if err != nil {
		_ = targetHandle.LinkDel(link)
		return fmt.Errorf("failed to list addr on link %s: %s", attrs.Name, err)
	}

	for _, addr := range addrs {
		if addr.IP.IsLinkLocalUnicast() {
			zap.S().Debugf("link %s already has lladdr, skipping configuration", attrs.Name)
			goto llfound
		}
	}

	err = targetHandle.AddrAdd(link, misc.NewLLAddr())
	if err != nil {
		_ = targetHandle.LinkDel(link)
		return fmt.Errorf("failed to add addr to link %s: %s", attrs.Name, err)
	}
	zap.S().Debugf("link %s linklocal address configured", attrs.Name)

llfound:
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		runtime.LockOSThread()
		err = netns.Set(targetNetns)
		if err != nil {
			_ = targetHandle.LinkDel(link)
			err = fmt.Errorf("failed to move into netns: %s", err)
			return
		}

		var wg *wgctrl.Client
		wg, err = wgctrl.New()
		if err != nil {
			_ = targetHandle.LinkDel(link)
			err = fmt.Errorf("failed to get wireguard control socket: %s", err)
			return
		}
		defer wg.Close()

		err = wg.ConfigureDevice(attrs.Name, attrs.Config)
		if err != nil {
			_ = targetHandle.LinkDel(link)
			err = fmt.Errorf("failed to configure wireguard interface %s: %s", attrs.Name, err)
			return
		}
		zap.S().Debugf("link %s wireguard configuration set", attrs.Name)
	}()
	waitGroup.Wait()

	if err != nil {
		return err
	}
	zap.S().Debugf("link %s ready", attrs.Name)
	return nil
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
		if link.Type() == "wireguard" && int(link.Attrs().Group) == i.group {
			list = append(list, misc.Link{
				Name: link.Attrs().Name,
				MTU:  link.Attrs().MTU,
			})
		}
	}
	return list, nil
}
