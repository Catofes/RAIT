package types

import (
	"fmt"
	"net"
	"strconv"
)

// ParseAddressFamily sanitizes the specified address family
// namely, only allowing ip4 and ip6
func ParseAddressFamily(af string) (string, error) {
	switch af {
	case "ip4":
		return "ip4", nil
	case "ip6":
		return "ip6", nil
	default:
		return "", fmt.Errorf("ParseAddressFamily: unsupported address family")
	}
}

// ParseEndpoint resolves the given ip endpoint according to the given address family
// on resolve failures, localhost will be returns instead of a hard failure
func ParseEndpoint(endpoint string, af string) (net.IP, error) {
	addr, err := net.ResolveIPAddr(af, endpoint)
	if err != nil || addr.IP == nil {
		switch af {
		case "ip4":
			return net.ParseIP("127.0.0.1"), nil
		case "ip6":
			return net.ParseIP("::1"), nil
		default:
			return nil, fmt.Errorf("ParseEndpoint: unsupported address family")
		}
	}
	return addr.IP, nil
}

// ParseUint16 ensures the given number is within the range of uint16
func ParseUint16(num string) (int, error) {
	n, err := strconv.ParseUint(num, 10, 16)
	return int(n), fmt.Errorf("ParseUint16: error parsing uint16: %w", err)
}
