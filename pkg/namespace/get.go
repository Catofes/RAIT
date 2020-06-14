package namespace

import (
	"errors"
	"fmt"
	"github.com/vishvananda/netns"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"os"
	"path"
	"runtime"
	"sync"
)

// GetFromName creates named network namespace,
// and returns current namespace if no name is specified
func GetFromName(name string) (netns.NsHandle, error) {
	logger := zap.S().Named("namespace.GetFromName")
	// shortcut for current namespace
	if name == "" {
		return netns.Get()
	}

	// shortcut for existing namespace
	ns, err := netns.GetFromName(name)
	if err == nil {
		return ns, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		logger.Errorf("unexpected error when getting netns handle: %s, error %s", name, err)
		return 0, err
	}

	// create the runtime dir if it does not exist
	// also try to replicate the behavior of iproute2 by mounting tmpfs onto it
	var runtimeDir = "/run/netns"
	_, err = os.Stat(runtimeDir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logger.Errorf("unexpected error when stating runtime dir: %s, error %s", runtimeDir, err)
			return 0, err
		}
		err = os.MkdirAll(runtimeDir, 0755)
		if err != nil {
			logger.Errorf("failed to create runtime dir: %s, error %s", runtimeDir, err)
			return 0, err
		}
		err = unix.Mount("tmpfs", runtimeDir, "tmpfs", unix.MS_NOSUID|unix.MS_NODEV, "mode=755")
		if err != nil {
			logger.Errorf("failed to mount tmpfs onto runtime dir: %s, error %s", runtimeDir, err)
			return 0, err
		}
	}

	// create the fd for the new namespace
	var nsPath = path.Join(runtimeDir, name)
	nsFd, err := os.Create(nsPath)
	if err != nil {
		logger.Errorf("failed to create ns fd: %s, error %s", nsPath, err)
		return 0, err
	}
	err = nsFd.Close()
	if err != nil {
		logger.Errorf("failed to close ns fd: %s, error %s", nsPath, err)
		return 0, err
	}
	// cleanup the fd file in case of failure
	// this has no effect when the new netns is successfully mounted
	defer os.RemoveAll(nsPath)

	// do the dirty jobs in a locked os thread
	// go runtime will reap it instead of reuse it
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		err = unix.Unshare(unix.CLONE_NEWNET)
		if err != nil {
			logger.Errorf("failed to unshare netns, error %s", err)
			return
		}
		threadNsPath := fmt.Sprintf("/proc/%d/task/%d/ns/net", os.Getpid(), unix.Gettid())
		err = unix.Mount(threadNsPath, nsPath, "none", unix.MS_BIND|unix.MS_SHARED|unix.MS_REC, "")
		if err != nil {
			logger.Errorf("failed to bind mount nsfs: %s, error %s", threadNsPath, err)
			return
		}
	}()
	wg.Wait()
	if err != nil {
		return 0, err
	}

	return netns.GetFromName(name)
}
