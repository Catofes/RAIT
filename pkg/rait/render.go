package rait

import (
	"fmt"
	"github.com/osteele/liquid"
	"gitlab.com/NickCao/RAIT/pkg/utils"
	"io/ioutil"
)

// RenderTemplate gathers information about interfaces and renders the liquid template
func (instance *Instance) RenderTemplate(in string, out string) error {
	reader, err := utils.ReadCloserFromPath(in)
	if err != nil {
		return err
	}
	defer reader.Close()
	writer, err := utils.WriteCloserFromPath(out)
	if err != nil {
		return err
	}
	defer writer.Close()

	tmpl, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("RenderTemplate: failed to read template: %w", err)
	}

	rawLinkList, err := instance.ListInterfaces()
	if err != nil {
		return err
	}

	var linkList []string
	for _, link := range rawLinkList {
		linkList = append(linkList, link.Attrs().Name)
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
