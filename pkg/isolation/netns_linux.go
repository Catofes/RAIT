package isolation

import (
	"errors"
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
)

func init(){
	Register("netns", NewNetnsIsolation)
}

// NetnsFromName creates and returns named network namespace,
// or the current namespace if no name is specified
func NetnsFromName(name string) (netns.NsHandle, error) {
	logger := zap.S().Named("isolation.NetnsFromName")
	// shortcut for current namespace
	if name == "" {
		return netns.Get()
	}

	// shortcut for existing namespace
	ns, err := netns.GetFromName(name)
	if err == nil {
		return ns, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		logger.Errorf("unexpected error when getting netns handle: %s, error %s", name, err)
		return 0, err
	}

	// create the runtime dir if it does not exist
	// also try to replicate the behavior of iproute2 by mounting tmpfs onto it
	var runtimeDir = "/run/netns"
	_, err = os.Stat(runtimeDir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logger.Errorf("unexpected error when stating runtime dir: %s, error %s", runtimeDir, err)
			return 0, err
		}
		err = os.MkdirAll(runtimeDir, 0755)
		if err != nil {
			logger.Errorf("failed to create runtime dir: %s, error %s", runtimeDir, err)
			return 0, err
		}
		err = unix.Mount("tmpfs", runtimeDir, "tmpfs", unix.MS_NOSUID|unix.MS_NODEV, "mode=755")
		if err != nil {
			logger.Errorf("failed to mount tmpfs onto runtime dir: %s, error %s", runtimeDir, err)
			return 0, err
		}
		logger.Debugf("created netns runtime dir: %s", runtimeDir)
	}

	// create the fd for the new namespace
	var nsPath = path.Join(runtimeDir, name)
	nsFd, err := os.Create(nsPath)
	if err != nil {
		logger.Errorf("failed to create ns fd: %s, error %s", nsPath, err)
		return 0, err
	}
	err = nsFd.Close()
	if err != nil {
		logger.Errorf("failed to close ns fd: %s, error %s", nsPath, err)
		return 0, err
	}
	// cleanup the fd file in case of failure
	// this has no effect when the new netns is successfully mounted
	defer os.RemoveAll(nsPath)

	// do the dirty jobs in a locked os thread
	// go runtime will reap it instead of reuse it
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		err = unix.Unshare(unix.CLONE_NEWNET)
		if err != nil {
			logger.Errorf("failed to unshare netns, error %s", err)
			return
		}
		threadNsPath := fmt.Sprintf("/proc/%d/task/%d/ns/net", os.Getpid(), unix.Gettid())
		err = unix.Mount(threadNsPath, nsPath, "none", unix.MS_BIND|unix.MS_SHARED|unix.MS_REC, "")
		if err != nil {
			logger.Errorf("failed to bind mount nsfs: %s, error %s", threadNsPath, err)
			return
		}
	}()
	wg.Wait()
	if err != nil {
		return 0, err
	}

	logger.Debugf("created namespace: %s", name)
	return netns.GetFromName(name)
}

// NetlinkFromName returns netlink handle created in the specified netns
func NetlinkFromName(name string) (*netlink.Handle, error) {
	ns, err := NetnsFromName(name)
	if err != nil {
		return nil, err
	}
	defer ns.Close()
	return netlink.NewHandleAt(ns)
}

// NetnsIsolation is the recommended implementation as by the wireguard developers
// It keeps the wireguard sockets and interfaces in different netns to facilitate isolation
type NetnsIsolation struct {
	TransitNamespace   string
	InterfaceNamespace string
}

// NewNetnsIsolation takes two arguments: transit and interface namespace
// the creation of netns is handled internally
// the links and sockets will be created in the transit namespace
// and the links will be moved into the interface namespace
func NewNetnsIsolation(transitNamespace, interfaceNamespace string) (Isolation, error) {
	return &NetnsIsolation{
		TransitNamespace:   transitNamespace,
		InterfaceNamespace: interfaceNamespace,
	}, nil
}

func (i *NetnsIsolation) LinkEnsure(name string, config wgtypes.Config, mtu, ifgroup int) (err error) {
	logger := zap.S().Named("isolation.NetnsIsolation.LinkEnsure")

	var interfaceHandle *netlink.Handle
	if interfaceHandle, err = NetlinkFromName(i.InterfaceNamespace); err != nil {
		return err
	}
	defer interfaceHandle.Delete()

	var interfaceNetns netns.NsHandle
	if interfaceNetns, err = NetnsFromName(i.InterfaceNamespace); err != nil {
		return err
	}
	defer interfaceNetns.Close()

	var link netlink.Link
	link, err = interfaceHandle.LinkByName(name)
	if err == nil {
		logger.Debugf("link %s already exists, skipping creation", name)
	} else if _, ok := err.(netlink.LinkNotFoundError); !ok {
		logger.Errorf("failed to get link %s: %s", name, err)
		return err
	} else {
		var transitHandle *netlink.Handle
		if transitHandle, err = NetlinkFromName(i.TransitNamespace); err != nil {
			return err
		}
		transitHandle.Delete()

		link = &netlink.Wireguard{LinkAttrs: netlink.LinkAttrs{Name: name}}
		err = transitHandle.LinkAdd(link)
		if err != nil {
			logger.Errorf("failed to create link %s: %s", name, err)
			return err
		}
		logger.Debugf("link %s created in transit namespace", name)

		err = transitHandle.LinkSetNsFd(link, int(interfaceNetns))
		if err != nil {
			logger.Errorf("failed to move link %s: %s", name, err)
			_ = transitHandle.LinkDel(link)
			return err
		}
		logger.Debugf("link %s moved into interface namespace", name)
	}

	err = interfaceHandle.LinkSetMTU(link, mtu)
	if err != nil {
		logger.Errorf("failed to set mtu on link %s: %s", name, err)
		_ = interfaceHandle.LinkDel(link)
		return err
	}
	logger.Debugf("link %s mtu set to %d", name, mtu)

	err = interfaceHandle.LinkSetGroup(link, ifgroup)
	if err != nil {
		logger.Errorf("failed to set group on link %s: %s", name, err)
		_ = interfaceHandle.LinkDel(link)
		return err
	}
	logger.Debugf("link %s ifgroup set to %d", name, ifgroup)

	err = interfaceHandle.LinkSetUp(link)
	if err != nil {
		logger.Errorf("failed to set up link %s: %s", name, err)
		_ = interfaceHandle.LinkDel(link)
		return err
	}
	logger.Debugf("link %s set up", name)

	var addrs []netlink.Addr
	addrs, err = interfaceHandle.AddrList(link, unix.AF_INET6)
	if err != nil {
		logger.Errorf("failed to list addr on link %s: %s", name, err)
		_ = interfaceHandle.LinkDel(link)
		return err
	}

	if len(addrs) == 0 {
		err = interfaceHandle.AddrAdd(link, misc.LinkLocalAddr())
		if err != nil {
			logger.Errorf("failed to add addr to link %s: %s", name, err)
			_ = interfaceHandle.LinkDel(link)
			return err
		}
		logger.Debugf("link %s linklocal address configured", name)
	} else {
		logger.Debugf("link %s already has address configured, skipping configuration", name)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		err = netns.Set(interfaceNetns)
		if err != nil {
			logger.Errorf("failed to move into netns: %s", err)
			_ = interfaceHandle.LinkDel(link)
			return
		}

		var wg *wgctrl.Client
		wg, err = wgctrl.New()
		if err != nil {
			logger.Errorf("failed to get wireguard control socket: %s", err)
			_ = interfaceHandle.LinkDel(link)
			return
		}
		defer wg.Close()

		err = wg.ConfigureDevice(name, config)
		if err != nil {
			logger.Errorf("failed to configure wireguard interface %s: %s", name, err)
			_ = interfaceHandle.LinkDel(link)
			return
		}
		logger.Debugf("link %s wireguard configuration set", name)
	}()
	wg.Wait()

	logger.Debugf("link %s ready", name)
	return err
}

func (i *NetnsIsolation) LinkAbsent(name string) error {
	logger := zap.S().Named("isolation.NetnsIsolation.LinkAbsent")

	var err error
	var interfaceHandle *netlink.Handle
	if interfaceHandle, err = NetlinkFromName(i.InterfaceNamespace); err != nil {
		return err
	}
	defer interfaceHandle.Delete()

	var link netlink.Link
	link, err = interfaceHandle.LinkByName(name)
	if err != nil {
		logger.Errorf("failed to get link %s: %s", name, err)
		return err
	}

	err = interfaceHandle.LinkDel(link)
	if err != nil {
		logger.Errorf("failed to delete link %s: %s", name, err)
		return err
	}
	logger.Debugf("link %s removed", name)

	return nil
}

func (i *NetnsIsolation) LinkFilter(prefix string, group int) ([]string, error) {
	logger := zap.S().Named("isolation.NetnsIsolation.LinkFilter")

	var err error
	var interfaceHandle *netlink.Handle
	if interfaceHandle, err = NetlinkFromName(i.InterfaceNamespace); err != nil {
		return nil, err
	}
	defer interfaceHandle.Delete()

	var rawList []netlink.Link
	rawList, err = interfaceHandle.LinkList()
	if err != nil {
		logger.Errorf("failed to list link: %s", err)
		return nil, err
	}

	var list []string
	for _, link := range rawList {
		if link.Type() == "wireguard" && strings.HasPrefix(link.Attrs().Name, prefix) && int(link.Attrs().Group) == group {
			list = append(list, link.Attrs().Name)
		}
	}

	return list, nil
}
