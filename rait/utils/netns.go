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

// GetNetNS returns the handle to the specified netns if it exists
func GetNetNS(name string) (netns.NsHandle, error) {
	var ns netns.NsHandle
	var err error
	switch name {
	case "":
		ns, err = netns.Get()
	default:
		ns, err = netns.GetFromName(name)
	}
	if err != nil {
		return 0, fmt.Errorf("GetNetNS: failed to get netns: %w", err)
	}
	return ns, nil
}

// EnsureNetNS creates the specified netns if it does not exist
func EnsureNetNS(name string) (netns.NsHandle, error) {
	// Short path for existing netns
	ns, err := GetNetNS(name)
	if err == nil {
		return ns, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return 0, fmt.Errorf("EnsureNetNS: unexpected error when getting netns handle: %w", err)
	}

	// Create the runtime dir if it does not exist, try to replicate the behavior of iproute2
	var runtimeDir = "/run/netns"
	_, err = os.Stat(runtimeDir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return 0, fmt.Errorf("EnsureNetNS: unexpected error when opening runtime dir: %w", err)
		}
		err = os.MkdirAll(runtimeDir, 0755)
		if err != nil {
			return 0, fmt.Errorf("EnsureNetNS: failed to create runtime dir: %w", err)
		}
		err = unix.Mount("tmpfs", runtimeDir, "tmpfs", unix.MS_NOSUID|unix.MS_NODEV, "mode=755")
		if err != nil {
			return 0, fmt.Errorf("EnsureNetNS: failed to mount tmpfs onto runtime dir: %w", err)
		}
	}

	var nsPath = path.Join(runtimeDir, name)
	nsFd, err := os.Create(nsPath)
	if err != nil {
		return 0, fmt.Errorf("EnsureNetNS: failed to create fd for netns: %w", err)
	}
	err = nsFd.Close()
	if err != nil {
		return 0, fmt.Errorf("EnsureNetNS: failed to close fd for netns: %w", err)
	}
	defer os.RemoveAll(nsPath)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		err = unix.Unshare(unix.CLONE_NEWNET)
		if err != nil {
			err = fmt.Errorf("EnsureNetNS: failed to unshare netns: %w", err)
			return
		}
		threadNsPath := fmt.Sprintf("/proc/%d/task/%d/ns/net", os.Getpid(), unix.Gettid())
		err = unix.Mount(threadNsPath, nsPath, "none", unix.MS_BIND|unix.MS_SHARED|unix.MS_REC, "")
		if err != nil {
			err = fmt.Errorf("EnsureNetNS: failed to bind mount netns at %s: %w", nsPath, err)
			return
		}
	}()
	wg.Wait()
	if err != nil {
		return 0, fmt.Errorf("EnsureNetNS: failed to create namespace: %w", err)
	}
	return GetNetNS(name)
}

// WithNetNS executes the given closure in specified netns
func WithNetNS(name string, fn func(handle *netlink.Handle) error) error {
	var err error
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		var ns netns.NsHandle
		ns, err = EnsureNetNS(name)
		if err != nil {
			return
		}
		defer ns.Close()
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
