package rait

import (
	"github.com/osteele/liquid"
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
	"go.uber.org/zap"
	"io/ioutil"
)

// RenderTemplate gathers information about interfaces and renders the liquid template
func RenderTemplate(in string, out string, ifnames []string) error {
	logger := zap.S().Named("rait.RenderTemplate")

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
		logger.Errorf("failed to read template %s: %s", in, err)
		return err
	}

	output, err := liquid.NewEngine().ParseAndRender(tmpl, map[string]interface{}{"LinkList": ifnames})
	if err != nil {
		logger.Errorf("failed to render template %s: %s", in, err)
		return err
	}

	_, err = writer.Write(output)
	if err != nil {
		logger.Errorf("failed to write output %s: %s", out, err)
		return err
	}

	return nil
}
