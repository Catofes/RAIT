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
	_ = CreateNamedNamespace(r.Namespace) // Won't hurt
	err = r.SetupWireguard(ps)
	if err != nil {
		return err
	}
	err = r.SetupVethPair()
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
	err = r.DestroyWireguard()
	if err != nil {
		return err
	}
	err = r.DestroyVethPair()
	if err != nil {
		return err
	}
	return nil
}

func EntryRender(raitFile string, tmplFile string) error {
	r, err := RAITFromFile(raitFile)
	if err != nil {
		return err
	}
	err = r.RenderTemplate(tmplFile)
	if err != nil {
		return err
	}
	return nil
}
