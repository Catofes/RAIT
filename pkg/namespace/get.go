package namespace

import (
	"errors"
	"fmt"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
	"os"
	"path"
	"runtime"
	"sync"
)

// GetFromName creates named network namespace,
// and returns current namespace if no name is specified
func GetFromName(name string) (netns.NsHandle, error) {
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
		return 0, fmt.Errorf("GetFromName: error getting netns handle: %w", err)
	}

	// create the runtime dir if it does not exist
	// also try to replicate the behavior of iproute2 by mounting tmpfs onto it
	var runtimeDir = "/run/netns"
	_, err = os.Stat(runtimeDir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return 0, fmt.Errorf("GetFromName: error opening runtime dir: %w", err)
		}
		err = os.MkdirAll(runtimeDir, 0755)
		if err != nil {
			return 0, fmt.Errorf("GetFromName: error creating runtime dir: %w", err)
		}
		err = unix.Mount("tmpfs", runtimeDir, "tmpfs", unix.MS_NOSUID|unix.MS_NODEV, "mode=755")
		if err != nil {
			return 0, fmt.Errorf("GetFromName: error mounting tmpfs onto runtime dir: %w", err)
		}
	}

	// create the fd for the new namespace
	var nsPath = path.Join(runtimeDir, name)
	nsFd, err := os.Create(nsPath)
	if err != nil {
		return 0, fmt.Errorf("GetFromName: error creating fd for new netns: %w", err)
	}
	err = nsFd.Close()
	if err != nil {
		return 0, fmt.Errorf("GetFromName: error closing fd for new netns: %w", err)
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
			err = fmt.Errorf("GetFromName: error unsharing netns: %w", err)
			return
		}
		threadNsPath := fmt.Sprintf("/proc/%d/task/%d/ns/net", os.Getpid(), unix.Gettid())
		err = unix.Mount(threadNsPath, nsPath, "none", unix.MS_BIND|unix.MS_SHARED|unix.MS_REC, "")
		if err != nil {
			err = fmt.Errorf("GetFromName: error bind mounting netns at %s: %w", nsPath, err)
			return
		}
	}()
	wg.Wait()
	if err != nil {
		return 0, fmt.Errorf("GetFromName: failed to create namespace: %w", err)
	}

	return netns.GetFromName(name)
}
