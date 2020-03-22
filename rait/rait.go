package rait

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

type RAIT struct {
	PrivateKey Key
	SendPort   uint16
	Babeld     string
	Veth       string
	Namespace  string
	IFPrefix   string
	MTU        uint16
	FwMark     uint16
	ULAName    string
}

func RAITFromFile(filePath string) (*RAIT, error) {
	var r = RAIT{
		Babeld:    "/run/rait/babeld.conf",
		Veth:      "gravity",
		Namespace: "gravity",
		IFPrefix:  "rait",
		MTU:       1400,
		FwMark:    0x36,
		ULAName:   "off",
	}
	_, err := toml.DecodeFile(filePath, &r)
	if err != nil {
		return nil, fmt.Errorf("RAITFromFile: failed to decode rait config at %v: %w", filePath, err)
	}
	return &r, nil
}
