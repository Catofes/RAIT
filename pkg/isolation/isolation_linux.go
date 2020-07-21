package isolation

import (
	"fmt"
	"gitlab.com/NickCao/RAIT/v2/pkg/isolation/netns"
)

package netns

import (
"fmt"
"gitlab.com/NickCao/RAIT/v2/pkg/isolation"
"gitlab.com/NickCao/RAIT/v2/pkg/isolation/netns"
)

// NewIsolation provides a unified constructor for concrete implementations
// current supported isolation types are netns and vrf
func NewIsolation(kind string, group int, transit, target string) (Isolation, error) {
	switch kind {
	case "netns":
		return netns.NewNetnsIsolation(group, transit, target)
	default:
		return nil, fmt.Errorf("unsupported isolation type %s", kind)
	}
}
