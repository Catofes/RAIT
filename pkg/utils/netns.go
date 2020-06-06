package utils

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
	"sync"
)

func WithNetNS(ns netns.NsHandle, fn func(handle *netlink.Handle) error) error {
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		err = netns.Set(ns)
		if err != nil {
			err = fmt.Errorf("WithNetNS: failed to set netns: %w", err)
			return
		}
		var handle *netlink.Handle
		handle, err = netlink.NewHandle()
		if err != nil {
			err = fmt.Errorf("WithNetNS: failed to get netlink handle: %w", err)
			return
		}
		defer handle.Delete()
		err = fn(handle)
	}()
	wg.Wait()
	return err
}
