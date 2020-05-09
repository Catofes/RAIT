package types

type AF string

const (
	AF_UNSPEC AF = ""
	AF_INET      = "ip4"
	AF_INET6     = "ip6"
)

func (af *AF) UnmarshalText(text []byte) (err error) {
	*af = AF(text)
	return
}
