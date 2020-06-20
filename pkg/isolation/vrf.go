package isolation

import (
	"github.com/vishvananda/netlink"
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"strings"
)

type VrfIsolation struct {
	TransitVrf   string
	InterfaceVrf string
}

func NewVrfIsolation(transitVrf, interfaceVrf string) *VrfIsolation {
	return &VrfIsolation{
		TransitVrf:   transitVrf,
		InterfaceVrf: interfaceVrf,
	}
}

func (i *VrfIsolation) LinkEnsure(name string, mtu, group int, config wgtypes.Config) (err error) {
	logger := zap.S().Named("VrfIsolation.LinkEnsure")

	var interfaceHandle *netlink.Handle
	if interfaceHandle, err = netlink.NewHandle(); err != nil {
		return err
	}
	defer interfaceHandle.Delete()

	var link netlink.Link
	link, err = interfaceHandle.LinkByName(name)
	if err == nil {
		logger.Debugf("link %s already exists, skipping creation", name)
	} else if _, ok := err.(netlink.LinkNotFoundError); !ok {
		logger.Errorf("failed to get link %s: %s", name, err)
		return err
	} else {
		var transitVrfIndex int
		if transitVrfIndex, err = VrfIndexFromName(i.TransitVrf); err != nil {
			return err
		}

		link = &netlink.Wireguard{LinkAttrs: netlink.LinkAttrs{Name: name, MasterIndex: transitVrfIndex}}
		err = interfaceHandle.LinkAdd(link)
		if err != nil {
			logger.Errorf("failed to create link %s: %s", name, err)
			return err
		}
	}

	var interfaceVrfIndex int
	if interfaceVrfIndex, err = VrfIndexFromName(i.InterfaceVrf); err != nil {
		return err
	}

	err = interfaceHandle.LinkSetMasterByIndex(link, interfaceVrfIndex)
	if err != nil {
		logger.Errorf("failed to set vrf on link %s: %s", name, err)
		_ = interfaceHandle.LinkDel(link)
		return err
	}

	err = interfaceHandle.LinkSetMTU(link, mtu)
	if err != nil {
		logger.Errorf("failed to set mtu on link %s: %s", name, err)
		_ = interfaceHandle.LinkDel(link)
		return err
	}

	err = interfaceHandle.LinkSetGroup(link, group)
	if err != nil {
		logger.Errorf("failed to set group on link %s: %s", name, err)
		_ = interfaceHandle.LinkDel(link)
		return err
	}

	err = interfaceHandle.LinkSetUp(link)
	if err != nil {
		logger.Errorf("failed to set up link %s: %s", name, err)
		_ = interfaceHandle.LinkDel(link)
		return err
	}

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
	}

	var wg *wgctrl.Client
	wg, err = wgctrl.New()
	if err != nil {
		logger.Errorf("failed to get wireguard control socket: %s", err)
		return err
	}
	defer wg.Close()

	err = wg.ConfigureDevice(name, config)
	if err != nil {
		logger.Errorf("failed to configure wireguard interface %s: %s", name, err)
		return err
	}

	return nil
}

func (i *VrfIsolation) LinkAbsent(name string) error {
	logger := zap.S().Named("VrfIsolation.LinkAbsent")

	var err error
	var interfaceHandle *netlink.Handle
	if interfaceHandle, err = netlink.NewHandle(); err != nil {
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

	return nil
}

func (i *VrfIsolation) LinkList(prefix string, group int) ([]string, error) {
	logger := zap.S().Named("VrfIsolation.LinkList")

	var err error
	var interfaceHandle *netlink.Handle
	if interfaceHandle, err = netlink.NewHandle(); err != nil {
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
