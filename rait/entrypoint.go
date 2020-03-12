package rait

func RAITUp(config string, peers string) error {
	r, err := NewRAITFromFile(config)
	if err != nil {
		return err
	}
	p, err := NewPeersFromDirectory(peers, ".conf")
	if err != nil {
		return err
	}
	err = r.Setup(p)
	if err != nil {
		return err
	}
	return nil
}

func RAITDown(config string) error {
	r, err := NewRAITFromFile(config)
	if err != nil {
		return err
	}
	return r.Destroy()
}
