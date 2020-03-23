package rait

import (
	"fmt"
	"github.com/osteele/liquid"
	"io/ioutil"
	"os"
	"strings"
)

func (r *RAIT) RenderTemplate(tmplFile string) error {
	var tmpl []byte
	var err error
	if tmplFile == "" {
		tmpl, err = ioutil.ReadAll(os.Stdin)
	} else {
		tmpl, err = ioutil.ReadFile(tmplFile)
	}
	if err != nil {
		return err
	}

	helper, err := NamespaceHelperFromName(r.Namespace)
	if err != nil {
		return fmt.Errorf("RenderTemplate: failed to get netns helper: %w", err)
	}
	defer helper.Destroy()

	linkList, err := helper.DstHandle.LinkList()
	if err != nil {
		return fmt.Errorf("RenderTemplate: failed to list interface: %w", err)
	}
	var IFList []string
	for _, link := range linkList {
		if link.Type() == "wireguard" && strings.HasPrefix(link.Attrs().Name, r.IFPrefix) {
			IFList = append(IFList, link.Attrs().Name)
		}
	}
	out, err := liquid.NewEngine().ParseAndRender(tmpl, map[string]interface{}{"IFList": IFList, "RAIT": r})
	_, err = fmt.Fprint(os.Stdout, string(out))
	if err != nil {
		return err
	}
	return nil
}
