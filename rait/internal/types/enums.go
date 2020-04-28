package types

type AF int

const (
	AF_UNSPEC AF = iota
	AF_INET
	AF_INET6
)

func (af *AF) UnmarshalText(text []byte) error {
	switch string(text) {
	case "ip4":
		*af = AF_INET
	case "ip6":
		*af = AF_INET6
	default:
		*af = AF_UNSPEC
	}
	return nil
}

func (af *AF) String() string {
	return []string{"unspec", "ip4", "ip6"}[*af]
}
