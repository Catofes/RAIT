package rait

func EntryUp(raitFile string, peerDir string) error {
	client, err := ClientFromFile(raitFile)
	if err != nil {
		return err
	}
	peers, err := PeersFromDirectory(peerDir)
	if err != nil {
		return err
	}
	err = client.SetupWireguardInterfaces(peers)
	if err != nil {
		return err
	}
	return nil
}

func EntryDown(raitFile string) error {
	client, err := ClientFromFile(raitFile)
	if err != nil {
		return err
	}
	err = client.DestroyWireguardInterfaces()
	if err != nil {
		return err
	}
	return nil
}

func EntryRender(raitFile string, tmplFile string) error {
	client, err := ClientFromFile(raitFile)
	if err != nil {
		return err
	}
	err = client.RenderTemplate(tmplFile)
	if err != nil {
		return err
	}
	return nil
}
