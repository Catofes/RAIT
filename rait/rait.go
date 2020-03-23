package rait

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

type RAIT struct {
	PrivateKey Key
	SendPort   uint16
	Namespace  string
	IFPrefix   string
	MTU        uint16
	Veth       string
	FwMark     uint16
}

func RAITFromFile(filePath string) (*RAIT, error) {
	var r = RAIT{
		Namespace: "rait",
		IFPrefix:  "rait",
		MTU:       1400,
	}
	_, err := toml.DecodeFile(filePath, &r)
	if err != nil {
		return nil, fmt.Errorf("RAITFromFile: failed to decode rait config at %v: %w", filePath, err)
	}
	return &r, nil
}
