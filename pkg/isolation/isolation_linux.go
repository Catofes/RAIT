package isolation

import (
	"github.com/Catofes/RAIT/v4/pkg/isolation/netns"
)

func NewIsolation(ifgroup int, transit, target string) (Isolation, error) {
	return netns.NewNetnsIsolation(ifgroup, transit, target)
}
