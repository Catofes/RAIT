package utils

import (
	"errors"
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
	"os"
	"path"
	"runtime"
	"sync"
)

type NetnsFn func(handle *netlink.Handle) error

// WithNetns executes the given closure in specified netns
func WithNetns(name string, fn NetnsFn) (err error) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread() // We are not unlocking it, and go runtime will reap the thread
		var ns netns.NsHandle
		ns, err = TryGetNetns(name)
		if err != nil {
			return
		}
		defer ns.Close()
		err = netns.Set(ns)
		if err != nil {
			err = fmt.Errorf("WithNetns: failed to set netns: %w", err)
			return
		}
		var handle *netlink.Handle
		handle, err = netlink.NewHandle()
		if err != nil {
			err = fmt.Errorf("WithNetns: failed to get netlink handle: %w", err)
			return
		}
		defer handle.Delete()
		err = fn(handle)
	}()
	wg.Wait()
	return
}

// TryGetNetns returns the handle to the specified netns
func TryGetNetns(name string) (ns netns.NsHandle, err error) {
	if name == "" {
		ns, err = netns.Get()
	} else {
		ns, err = netns.GetFromName(name)
	}
	if err != nil {
		err = fmt.Errorf("TryGetNetns: failed to get netns: %w", err)
		return
	}
	return
}

// GetNetns creates the specified netns if it does not exist
func GetNetns(name string) (ns netns.NsHandle, err error) {
	ns, err = TryGetNetns(name)
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		return
	}

	var runtimeDir = "/run/netns"
	if _, err = os.Stat(runtimeDir); err != nil {
		if err = os.MkdirAll(runtimeDir, 0755); err != nil {
			err = fmt.Errorf("GetNetns: failed to create runtime dir: %w", err)
			return
		}
		if err = unix.Mount("tmpfs", runtimeDir, "tmpfs", unix.MS_NOSUID|unix.MS_NODEV, "mode=755"); err != nil {
			err = fmt.Errorf("GetNetns: failed to mount tmpfs onto runtime dir: %w", err)
			return
		}
	}

	var nsPath = path.Join(runtimeDir, name)
	var nsFd *os.File
	if nsFd, err = os.Create(nsPath); err != nil {
		err = fmt.Errorf("GetNetns: failed to create fd for ns: %w", err)
		return
	}
	if err = nsFd.Close(); err != nil {
		err = fmt.Errorf("GetNetns: failed to close fd for ns: %w", err)
		return
	}
	defer os.RemoveAll(nsPath)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		if err = unix.Unshare(unix.CLONE_NEWNET); err != nil {
			err = fmt.Errorf("GetNetns: failed to unshare netns: %w", err)
			return
		}
		threadNsPath := fmt.Sprintf("/proc/%d/task/%d/ns/net", os.Getpid(), unix.Gettid())
		if err = unix.Mount(threadNsPath, nsPath, "none", unix.MS_BIND|unix.MS_SHARED|unix.MS_REC, ""); err != nil {
			err = fmt.Errorf("GetNetns: failed to bind mount ns at %s: %w", nsPath, err)
			return
		}
	}()
	wg.Wait()
	if err != nil {
		err = fmt.Errorf("GetNetns: failed to create namespace: %w", err)
		return
	}
	return TryGetNetns(name)
}
