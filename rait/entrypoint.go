package rait

func EntryUp(raitFile string, peerDir string) error {
	r, err := RAITFromFile(raitFile)
	if err != nil {
		return err
	}
	ps, err := PeersFromDirectory(peerDir)
	if err != nil {
		return err
	}
	err = r.SetupWireguard(ps)
	if err != nil {
		return err
	}
	err = r.SetupVethPair()
	if err != nil {
		return err
	}
	err = r.SetupLoopback()
	if err != nil {
		return err
	}
	err = r.GenerateBabeldConfig()
	if err != nil {
		return err
	}
	return nil
}

func EntryDown(raitFile string) error {
	r, err := RAITFromFile(raitFile)
	if err != nil {
		return err
	}
	_ = r.DestroyLoopback()
	_ = r.DestroyVethPair()
	_ = r.DestroyWireguard()
	return nil
}
