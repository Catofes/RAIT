package rait

func EntryUp(raitFile string, peerPath string) error {
	client, err := ClientFromURL(raitFile)
	if err != nil {
		return err
	}
	var peers []*Peer
	peers, err = PeersFromURL(peerPath)
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
	client, err := ClientFromURL(raitFile)
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
	client, err := ClientFromURL(raitFile)
	if err != nil {
		return err
	}
	err = client.RenderTemplate(tmplFile)
	if err != nil {
		return err
	}
	return nil
}
