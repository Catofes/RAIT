package isolation

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

func VrfIndexFromName(name string) (int, error) {
	logger := zap.S().Named("isolation.VrfFromName")

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
