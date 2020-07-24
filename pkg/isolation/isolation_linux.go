package isolation

import (
	"gitlab.com/NickCao/RAIT/v3/pkg/isolation/netns"
)

func NewIsolation(ifgroup int, transit, target string) (Isolation, error) {
	return netns.NewNetnsIsolation(ifgroup, transit, target)
}