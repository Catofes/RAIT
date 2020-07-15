package isolation

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"runtime"
	"sync"
)

// TODO: clean up all these mess

func init() {
	Register("netns", NewNetnsIsolation)
}

// NetnsIsolation is the recommended implementation as by the wireguard developers
// It keeps the wireguard sockets and interfaces in different netns to facilitate isolation
type NetnsIsolation struct {
	transitNamespace   string
	interfaceNamespace string
}

// NewNetnsIsolation takes two arguments: transit and interface namespace
// the creation of netns is handled internally
// the links and sockets will be created in the transit namespace
// and the links will be moved into the interface namespace
func NewNetnsIsolation(transitNamespace, interfaceNamespace string) (Isolation, error) {
	return &NetnsIsolation{
		transitNamespace:   transitNamespace,
		interfaceNamespace: interfaceNamespace,
	}, nil
}

func (i *NetnsIsolation) LinkEnsure(attrs *LinkAttrs, config wgtypes.Config) (err error) {
	interfaceHandle, err := NetlinkFromName(i.interfaceNamespace)
	if err != nil {
		return err
	}
	defer interfaceHandle.Delete()

	interfaceNetns, err := NetNSFromName(i.interfaceNamespace)
	if err != nil {
		return err
	}
	defer interfaceNetns.Close()

	var link netlink.Link
	link, err = interfaceHandle.LinkByName(attrs.Name)
	if err == nil {
		zap.S().Debugf("link %s already exists, skipping creation", attrs.Name)
	} else if _, ok := err.(netlink.LinkNotFoundError); !ok {
		return fmt.Errorf("failed to get link %s: %w", attrs.Name, err)
	} else {
		transitHandle, err := NetlinkFromName(i.transitNamespace)
		if err != nil {
			return err
		}
		transitHandle.Delete()

		link = &netlink.Wireguard{LinkAttrs: netlink.LinkAttrs{Name: attrs.Name}}
		err = transitHandle.LinkAdd(link)
		if err != nil {
			return fmt.Errorf("failed to create link %s: %w", attrs.Name, err)
		}
		zap.S().Debugf("link %s created in transit namespace", attrs.Name)

		err = transitHandle.LinkSetNsFd(link, int(interfaceNetns))
		if err != nil {
			_ = transitHandle.LinkDel(link)
			return fmt.Errorf("failed to move link %s: %w", attrs.Name, err)
		}
		zap.S().Debugf("link %s moved into interface namespace", attrs.Name)
	}

	err = interfaceHandle.LinkSetMTU(link, attrs.MTU)
	if err != nil {
		_ = interfaceHandle.LinkDel(link)
		return fmt.Errorf("failed to set mtu on link %s: %w", attrs.Name, err)
	}
	zap.S().Debugf("link %s mtu set to %d", attrs.Name, attrs.MTU)

	err = interfaceHandle.LinkSetGroup(link, attrs.Group)
	if err != nil {
		_ = interfaceHandle.LinkDel(link)
		return fmt.Errorf("failed to set group on link %s: %w", attrs.Name, err)
	}
	zap.S().Debugf("link %s ifgroup set to %d", attrs.Name, attrs.Group)

	err = interfaceHandle.LinkSetUp(link)
	if err != nil {
		_ = interfaceHandle.LinkDel(link)
		return fmt.Errorf("failed to set up link %s: %w", attrs.Name, err)
	}
	zap.S().Debugf("link %s set up", attrs.Name)

	var addrs []netlink.Addr
	addrs, err = interfaceHandle.AddrList(link, unix.AF_INET6)
	if err != nil {
		_ = interfaceHandle.LinkDel(link)
		return fmt.Errorf("failed to list addr on link %s: %w", attrs.Name, err)
	}

	if len(addrs) == 0 {
		err = interfaceHandle.AddrAdd(link, misc.LinkLocalAddr())
		if err != nil {
			_ = interfaceHandle.LinkDel(link)
			return fmt.Errorf("failed to add addr to link %s: %w", attrs.Name, err)
		}
		zap.S().Debugf("link %s linklocal address configured", attrs.Name)
	} else {
		zap.S().Debugf("link %s already has address configured, skipping configuration", attrs.Name)
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		runtime.LockOSThread()
		err = netns.Set(interfaceNetns)
		if err != nil {
			_ = interfaceHandle.LinkDel(link)
			err = fmt.Errorf("failed to move into netns: %w", err)
			return
		}

		var wg *wgctrl.Client
		wg, err = wgctrl.New()
		if err != nil {
			_ = interfaceHandle.LinkDel(link)
			err = fmt.Errorf("failed to get wireguard control socket: %w", err)
			return
		}
		defer wg.Close()

		err = wg.ConfigureDevice(attrs.Name, config)
		if err != nil {
			_ = interfaceHandle.LinkDel(link)
			err = fmt.Errorf("failed to configure wireguard interface %s: %w", attrs.Name, err)
			return
		}
		zap.S().Debugf("link %s wireguard configuration set", attrs.Name)
	}()
	waitGroup.Wait()

	zap.S().Debugf("link %s ready", attrs.Name)
	return err
}

func (i *NetnsIsolation) LinkAbsent(attrs *LinkAttrs) error {
	interfaceHandle, err := NetlinkFromName(i.interfaceNamespace)
	if err != nil {
		return err
	}
	defer interfaceHandle.Delete()

	link, err := interfaceHandle.LinkByName(attrs.Name)
	if err != nil {
		return fmt.Errorf("failed to get link %s: %w", attrs.Name, err)
	}

	err = interfaceHandle.LinkDel(link)
	if err != nil {
		return fmt.Errorf("failed to delete link %s: %w", attrs.Name, err)
	}
	zap.S().Debugf("link %s removed", attrs.Name)
	return nil
}

func (i *NetnsIsolation) LinkList() ([]*LinkAttrs, error) {
	interfaceHandle, err := NetlinkFromName(i.interfaceNamespace)
	if err != nil {
		return nil, err
	}
	defer interfaceHandle.Delete()

	rawList, err := interfaceHandle.LinkList()
	if err != nil {
		return nil, fmt.Errorf("failed to list link: %w", err)
	}

	var list []*LinkAttrs
	for _, link := range rawList {
		if link.Type() == "wireguard" {
			list = append(list, &LinkAttrs{
				Name:  link.Attrs().Name,
				MTU:   link.Attrs().MTU,
				Group: int(link.Attrs().Group),
			})
		}
	}
	return list, nil
}
