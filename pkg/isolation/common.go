package isolation

import (
	"github.com/Catofes/RAIT/pkg/misc"
)

// Isolation represents a management interface for wireguard links
// together with the isolation technique employed to isolate overlay from underlay
type Isolation interface {
	// LinkEnsure ensures the existence and state of the given link is as expected
	// this method should be idempotent as it's also used to sync the state of links
	LinkEnsure(link misc.Link) error
	// LinkAbsent ensures the absence of the given link, only name should be used to identify the link
	LinkAbsent(link misc.Link) error
	// LinkList returns the wireguard links as seen by the isolation, aka managed interfaces
	LinkList() ([]misc.Link, error)
}
