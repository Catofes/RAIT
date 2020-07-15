package rait

import (
	"fmt"
	"github.com/osteele/liquid"
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
	"go.uber.org/zap"
	"io/ioutil"
)

// RenderTemplate gathers information about interfaces and renders the liquid template
func RenderTemplate(in string, out string, ifnames []string) error {
	zap.S().Warn("rait render is deprecated in favor of rait babeld sync, and will be removed in a future release")
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
		return fmt.Errorf("failed to read template %s: %w", in, err)
	}

	output, err := liquid.NewEngine().ParseAndRender(tmpl, map[string]interface{}{"LinkList": ifnames})
	if err != nil {
		return fmt.Errorf("failed to render template %s: %w", in, err)
	}

	_, err = writer.Write(output)
	if err != nil {
		return fmt.Errorf("failed to write template output %s: %w", out, err)
	}
	return nil
}
