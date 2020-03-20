package rait

func EntryUp(raitFile string, peerDir string, babeld string, veth bool) error {
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
	if veth {
		err = r.SetupVethPair()
		if err != nil {
			return err
		}
	}
	if babeld != "" {
		err = GenerateBabeldConfig(r.IFPrefix, len(ps), babeld)
		if err != nil {
			return err
		}
	}
	return nil
}

func EntryDown(raitFile string) error {
	r, err := RAITFromFile(raitFile)
	if err != nil {
		return err
	}
	_ = r.DestroyVethPair()
	_ = r.DestroyWireguard()
	return nil
}
