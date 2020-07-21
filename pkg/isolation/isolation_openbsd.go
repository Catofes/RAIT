package isolation

import (
	"fmt"
)

// NewIsolation provides a unified constructor for concrete implementations
// current supported isolation types are netns and vrf
func NewIsolation(kind string, group int, transit, target string) (Isolation, error) {
	switch kind {
	default:
		return nil, fmt.Errorf("unsupported isolation type %s", kind)
	}
}
