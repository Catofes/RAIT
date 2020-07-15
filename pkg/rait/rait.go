package rait

type RAIT struct {
	Instances []*Instance
}

func RAITFromPaths(paths []string) (*RAIT, error) {
	var ra RAIT
	for _, p := range paths {
		i, err := InstanceFromPath(p)
		if err != nil {
			return nil, err
		}
		ra.Instances = append(ra.Instances, i)
	}
	return &ra, nil
}

func (ra *RAIT) ListInterfaceName() ([]string, error) {
	var list []string
	for _, i := range ra.Instances {
		l, err := i.ListInterfaceName()
		if err != nil {
			return nil, err
		}
		list = append(list, l...)
	}
	return list, nil
}

func (ra *RAIT) SyncInterfaces(up bool) error {
	for _, i := range ra.Instances {
		err := i.SyncInterfaces(up)
		if err != nil {
			return err
		}
	}
	return nil
}
