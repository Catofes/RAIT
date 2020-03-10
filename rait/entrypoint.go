package rait

import (
	"io"
	"io/ioutil"
)

func RAITUp(config string, peers string) error {
	var r *RAIT
	var p []*Peer
	var err error
	r, err = NewRAITFromToml(config)
	if err != nil {
		return err
	}
	p, err = LoadPeersFromTomls(peers, ".conf")
	if err != nil {
		return err
	}
	err = r.SetupWireguardLinks(p)
	if err != nil {
		return err
	}
	if r.DummyName != "" {
		err = r.SetupDummyInterface()
		if err != nil {
			return err
		}
	}
	return nil
}

func RAITLoad(stream io.Reader) error {
	var r *RAIT
	var p []*Peer
	var data []byte
	var err error
	data, err = ioutil.ReadAll(stream)
	if err != nil {
		return nil
	}
	r, p, err = LoadFromJSON(data)
	if err != nil {
		return nil
	}
	err = r.SetupWireguardLinks(p)
	if err != nil {
		return err
	}
	if r.DummyName != "" {
		err = r.SetupDummyInterface()
		if err != nil {
			return err
		}
	}
	return nil
}

func RAITDown(config string) error {
	var r *RAIT
	var err error
	r, err = NewRAITFromToml(config)
	if err != nil {
		return err
	}
	err = r.DestroyWireguardLinks()
	if err != nil {
		return err
	}
	err = r.DestroyDummyInterface()
	if err != nil {
		return err
	}
	return nil
}
