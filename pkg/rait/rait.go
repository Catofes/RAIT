package rait

type RAIT struct {
	Instances []*Instance
}

func RAITFromPaths(paths []string) (*RAIT, error) {
	var ra RAIT
	for _, path := range paths {
		instance, err := InstanceFromPath(path)
		if err != nil {
			return nil, err
		}
		ra.Instances = append(ra.Instances, instance)
	}
	return &ra, nil
}

func (ra *RAIT) ListInterfaceName() ([]string, error) {
	var list []string
	for _, instance := range ra.Instances {
		names, err := instance.ListInterfaceName()
		if err != nil {
			return nil, err
		}
		list = append(list, names...)
	}
	return list, nil
}

func (ra *RAIT) SyncInterfaces(up bool) error {
	for _, instance := range ra.Instances {
		err := instance.SyncInterfaces(up)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ra *RAIT) RenderTemplate(in, out string) error {
	list, err := ra.ListInterfaceName()
	if err != nil {
		return err
	}
	return RenderTemplate(in, out, list)
}
