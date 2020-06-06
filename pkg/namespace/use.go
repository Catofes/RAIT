package namespace

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
	"sync"
)

// WithNetlink executes the given closure within the specified namespace
// and passes in a netlink handle created in the namespace
func WithNetlink(ns netns.NsHandle, fn func(handle *netlink.Handle) error) error {
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		err = netns.Set(ns)
		if err != nil {
			err = fmt.Errorf("WithNetlink: failed to set netns: %w", err)
			return
		}
		var handle *netlink.Handle
		handle, err = netlink.NewHandle()
		if err != nil {
			err = fmt.Errorf("WithNetlink: failed to get netlink handle: %w", err)
			return
		}
		defer handle.Delete()
		err = fn(handle)
	}()
	wg.Wait()
	return err
}
