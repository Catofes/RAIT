package namespace

import (
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"go.uber.org/zap"
	"runtime"
	"sync"
)

// WithNetlink executes the given closure within the specified namespace
// and passes in a netlink handle created in the namespace
func WithNetlink(ns netns.NsHandle, fn func(handle *netlink.Handle) error) error {
	logger := zap.S().Named("namespace.WithNetlink")
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		err = netns.Set(ns)
		if err != nil {
			logger.Errorf("failed to move into netns, error %s", err)
			return
		}
		var handle *netlink.Handle
		handle, err = netlink.NewHandle()
		if err != nil {
			logger.Errorf("failed to get netlink handle, error %s", err)
			return
		}
		defer handle.Delete()
		err = fn(handle)
	}()
	wg.Wait()
	return err
}
