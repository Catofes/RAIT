package isolation

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"strings"
)

func init(){
	Register("vrf", NewVrfIsolation)
}

// VrfIndexFromName returns the index of the specified vrf interface
func VrfIndexFromName(name string) (int, error) {
	logger := zap.S().Named("isolation.VrfIndexFromName")
	if name == "" {
		return 0, nil
	}
	link, err := netlink.LinkByName(name)
	if err != nil {
		logger.Errorf("failed to get link with name: %s", err)
		return 0, err
	}
	vrf, ok := link.(*netlink.Vrf)
	if !ok {
		logger.Errorf("the link named %s is not a vrf", name)
		return 0, fmt.Errorf("the link named %s is not a vrf", name)
	}
	return vrf.Index, nil
}

// VrfIsolation is an emerging technology as it provides less isolation than netns
// and eases the pain of cross namespace routing by eliminating the need of veth pairs
type VrfIsolation struct {
	TransitVrf   string
	InterfaceVrf string
}

// NewVrfIsolation takes two arguments: transit and interface vrf
// the links and sockets will be created in the transit vrf
// and the links will be moved into the interface vrf
// however, due to the technical limitation of wireguard
// the transit vrf param may not function as intended
// the vrf interfaces should be created in advance
func NewVrfIsolation(transitVrf, interfaceVrf string) (Isolation, error) {
	return &VrfIsolation{
		TransitVrf:   transitVrf,
		InterfaceVrf: interfaceVrf,
	}, nil
}

func (i *VrfIsolation) LinkEnsure(name string, config wgtypes.Config, mtu, ifgroup int) (err error) {
	logger := zap.S().Named("isolation.VrfIsolation.LinkEnsure")

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
		logger.Debugf("link %s created in transit vrf", name)
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
	logger.Debugf("link %s moved into interface vrf", name)

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

	var wg *wgctrl.Client
	wg, err = wgctrl.New()
	if err != nil {
		logger.Errorf("failed to get wireguard control socket: %s", err)
		_ = interfaceHandle.LinkDel(link)
		return err
	}
	defer wg.Close()

	err = wg.ConfigureDevice(name, config)
	if err != nil {
		logger.Errorf("failed to configure wireguard interface %s: %s", name, err)
		_ = interfaceHandle.LinkDel(link)
		return err
	}
	logger.Debugf("link %s wireguard configuration set", name)

	logger.Debugf("link %s ready", name)
	return nil
}

func (i *VrfIsolation) LinkAbsent(name string) error {
	logger := zap.S().Named("isolation.VrfIsolation.LinkAbsent")

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
	logger.Debugf("link %s removed", name)

	return nil
}

func (i *VrfIsolation) LinkFilter(prefix string, group int) ([]string, error) {
	logger := zap.S().Named("isolation.VrfIsolation.LinkFilter")

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
