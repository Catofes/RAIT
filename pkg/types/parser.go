package types

import (
	"errors"
	"fmt"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
	"net"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"
)

func OrDefault(value string, def string) string {
	if value == "" {
		return def
	}
	return value
}

func ParseAddressFamily(af string) (string, error) {
	switch af {
	case "ip4":
		return "ip4", nil
	case "ip6":
		return "ip6", nil
	default:
		return "", fmt.Errorf("unsupported address family")
	}
}

func ParseEndpoint(endpoint string, af string) (net.IP, error) {
	addr, err := net.ResolveIPAddr(af, endpoint)
	if err != nil || addr.IP == nil {
		switch af {
		case "ip4":
			return net.ParseIP("127.0.0.1"), nil
		case "ip6":
			return net.ParseIP("::1"), nil
		default:
			return nil, fmt.Errorf("unsupported address family")
		}
	}
	return addr.IP, nil
}

func ParseUint16(num string) (int, error) {
	n, err := strconv.ParseUint(num, 10, 16)
	return int(n), err
}

func ParseNamespace(name string) (netns.NsHandle, error) {
	if name == "" {
		return netns.Get()
	}

	ns, err := netns.GetFromName(name)
	if err == nil {
		return ns, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return 0, fmt.Errorf("ParseNamespace: unexpected error when getting netns handle: %w", err)
	}

	// Create the runtime dir if it does not exist, try to replicate the behavior of iproute2
	var runtimeDir = "/run/netns"
	_, err = os.Stat(runtimeDir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return 0, fmt.Errorf("ParseNamespace: unexpected error when opening runtime dir: %w", err)
		}
		err = os.MkdirAll(runtimeDir, 0755)
		if err != nil {
			return 0, fmt.Errorf("ParseNamespace: failed to create runtime dir: %w", err)
		}
		err = unix.Mount("tmpfs", runtimeDir, "tmpfs", unix.MS_NOSUID|unix.MS_NODEV, "mode=755")
		if err != nil {
			return 0, fmt.Errorf("ParseNamespace: failed to mount tmpfs onto runtime dir: %w", err)
		}
	}

	var nsPath = path.Join(runtimeDir, name)
	nsFd, err := os.Create(nsPath)
	if err != nil {
		return 0, fmt.Errorf("ParseNamespace: failed to create fd for netns: %w", err)
	}
	err = nsFd.Close()
	if err != nil {
		return 0, fmt.Errorf("ParseNamespace: failed to close fd for netns: %w", err)
	}
	defer os.RemoveAll(nsPath)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		err = unix.Unshare(unix.CLONE_NEWNET)
		if err != nil {
			err = fmt.Errorf("ParseNamespace: failed to unshare netns: %w", err)
			return
		}
		threadNsPath := fmt.Sprintf("/proc/%d/task/%d/ns/net", os.Getpid(), unix.Gettid())
		err = unix.Mount(threadNsPath, nsPath, "none", unix.MS_BIND|unix.MS_SHARED|unix.MS_REC, "")
		if err != nil {
			err = fmt.Errorf("ParseNamespace: failed to bind mount netns at %s: %w", nsPath, err)
			return
		}
	}()
	wg.Wait()
	if err != nil {
		return 0, fmt.Errorf("ParseNamespace: failed to create namespace: %w", err)
	}

	return netns.GetFromName(name)
}
