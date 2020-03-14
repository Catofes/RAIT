package rait

func EntryUp(raitFile string, peerDir string, babeld string) error {
	r, err := RAITFromFile(raitFile)
	if err != nil {
		return err
	}
	ps, err := PeersFromDirectory(peerDir)
	if err != nil {
		return err
	}
	err = r.Setup(ps)
	if err != nil {
		return err
	}
	if babeld != "" {
		return GenerateBabeldConfig(r.IFPrefix, len(ps), babeld)
	}
	return nil
}

func EntryDown(raitFile string) error {
	r, err := RAITFromFile(raitFile)
	if err != nil {
		return err
	}
	return r.Destroy()
}
