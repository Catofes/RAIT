package types

import (
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
	"go.uber.org/zap"
	"net"
)

// ParseAddressFamily sanitizes the specified address family
// namely, only allowing ip4 and ip6, default ip4
func ParseAddressFamily(family interface{}) string {
	logger := zap.S().Named("types.ParseAddressFamily")
	familyString := misc.Fallback(family, "ip4").(string)

	switch familyString {
	case "ip4", "ip6":
		return familyString
	default:
		logger.Errorf("unsupported address family: %s", familyString)
		return "ip4"
	}
}

// ParseEndpoint resolves the given ip endpoint according to the given address family
// on resolve failures, localhost will be returns instead of a hard failure
func ParseEndpoint(endpoint interface{}, family interface{}) net.IP {
	logger := zap.S().Named("types.ParseEndpoint")
	familyString := ParseAddressFamily(family)
	endpointString := misc.Fallback(endpoint, "localhost").(string)

	addr, err := net.ResolveIPAddr(familyString, endpointString)
	if err != nil || addr.IP == nil {
		logger.Errorf("failed to resolve endpoint, falling back to localhost, error %s", err)
		switch familyString {
		case "ip4":
			return net.ParseIP("127.0.0.1")
		default:
			return net.ParseIP("::1")
		}
	}
	return addr.IP
}

// ParseInt64 ensures the given interface is int64
func ParseInt64(number interface{}, fallback int) int {
	return int(misc.Fallback(number, int64(fallback)).(int64))
}
