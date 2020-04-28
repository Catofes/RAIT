package types

import (
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Key is a wrapper around wgtypes.Key, and implements encoding.TextUnmarshaler
type Key struct {
	wgtypes.Key
}

func (k *Key) UnmarshalText(text []byte) error {
	var err error
	k.Key, err = wgtypes.ParseKey(string(text))
	return err
}
