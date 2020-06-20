package rait

import (
	"fmt"
	"github.com/osteele/liquid"
	"gitlab.com/NickCao/RAIT/v2/pkg/isolation"
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
	"io/ioutil"
)

// RenderTemplate gathers information about interfaces and renders the liquid template
func (instance *Instance) RenderTemplate(in string, out string) error {
	reader, err := misc.ReadCloserFromPath(in)
	if err != nil {
		return err
	}
	defer reader.Close()
	writer, err := misc.WriteCloserFromPath(out)
	if err != nil {
		return err
	}
	defer writer.Close()

	tmpl, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("RenderTemplate: failed to read template: %w", err)
	}


	gi, err := isolation.NewGenericIsolation(instance.Isolation, instance.TransitNamespace, instance.InterfaceNamespace)
	if err != nil {
		return err
	}

	linkList, err := gi.LinkList(instance.InterfacePrefix,instance.InterfaceGroup)
	if err != nil {
		return err
	}

	output, err := liquid.NewEngine().ParseAndRender(tmpl, map[string]interface{}{"LinkList": linkList, "Instance": instance})
	if err != nil {
		return fmt.Errorf("RenderTemplate: failed to render template: %w", err)
	}

	_, err = writer.Write(output)
	if err != nil {
		return fmt.Errorf("RenderTemplate: failed to write output: %w", err)
	}

	return nil
}
