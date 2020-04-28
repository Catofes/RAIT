package utils

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
	"os"
	"path"
	"runtime"
	"sync"
)

// NetlinkHelper handles the lifecycle of NetlinkHandle and NamespaceHandle
type NetlinkHelper struct {
	NetlinkHandle   *netlink.Handle
	NamespaceHandle netns.NsHandle
}

func (helper *NetlinkHelper) Destroy() {
	helper.NetlinkHandle.Delete()
	_ = helper.NamespaceHandle.Close()
}

// NetlinkHelperFromName creates a NetlinkHelper from the specified netns, or empty from the original netns
func NetlinkHelperFromName(name string) (*NetlinkHelper, error) {
	var helper NetlinkHelper
	var err error
	switch name {
	case "":
		helper.NamespaceHandle, err = netns.Get()
		if err != nil {
			return nil, fmt.Errorf("failed to get namespace handle: %w", err)
		}
		helper.NetlinkHandle, err = netlink.NewHandle()
		if err != nil {
			return nil, fmt.Errorf("failed to get netlink handle: %w", err)
		}
	default:
		err = CreateNamespaceFromNameIfNotExist(name)
		if err != nil {
			return nil, fmt.Errorf("failed to create named namespace: %w", err)
		}
		helper.NamespaceHandle, err = netns.GetFromName(name)
		if err != nil {
			return nil, fmt.Errorf("failed to get namespace handle: %w", err)
		}
		helper.NetlinkHandle, err = netlink.NewHandleAt(helper.NamespaceHandle)
		if err != nil {
			return nil, fmt.Errorf("failed to get netlink handle: %w", err)
		}
	}
	return &helper, nil
}

// CreateNamespaceFromNameIfNotExist creates named netns in the runtime dir
// It serves as a replacement of ip netns add
func CreateNamespaceFromNameIfNotExist(name string) error {
	var runtimeDir = "/run/netns"
	var nsPath = path.Join(runtimeDir, name)
	var handle netns.NsHandle
	var err error
	handle, err = netns.GetFromName(name)
	if err == nil {
		handle.Close()
		return nil
	}
	// Don't touch it if the runtime dir exists
	_, err = os.Stat(runtimeDir)
	if err != nil {
		err = os.MkdirAll(runtimeDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create runtime dir: %w", err)
		}
		err = unix.Mount("tmpfs", runtimeDir, "tmpfs", unix.MS_NOSUID|unix.MS_NODEV, "mode=755")
		if err != nil {
			return fmt.Errorf("failed to mount tmpfs onto runtime dir: %w", err)
		}
	}

	var nsFd *os.File
	nsFd, err = os.Create(nsPath)
	if err != nil {
		return fmt.Errorf("failed to create fd for ns: %w", err)
	}
	err = nsFd.Close()
	if err != nil {
		return fmt.Errorf("failed to close fd for ns: %w", err)
	}
	defer os.RemoveAll(nsPath)

	var wg sync.WaitGroup
	wg.Add(1)
	go (func() {
		defer wg.Done()
		runtime.LockOSThread()
		threadNsPath := fmt.Sprintf("/proc/%d/task/%d/ns/net", os.Getpid(), unix.Gettid())
		err = unix.Unshare(unix.CLONE_NEWNET)
		if err != nil {
			err = fmt.Errorf("failed to unshare netns: %w", err)
			return
		}
		err = unix.Mount(threadNsPath, nsPath, "none", unix.MS_BIND|unix.MS_SHARED|unix.MS_REC, "")
		if err != nil {
			err = fmt.Errorf("failed to bind mount ns at %s: %w", nsPath, err)
			return
		}
	})()
	wg.Wait()
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}
	return nil
}
