package rait

import (
	"fmt"
	"github.com/osteele/liquid"
	"github.com/vishvananda/netlink"
	"gitlab.com/NickCao/RAIT/rait/internal/utils"
	"io/ioutil"
	"os"
	"strings"
)

// RenderTemplate gathers rait related information and renders the liquid template pro
func (client *Client) RenderTemplate(tmplFile string) error {
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

	var helper *utils.NetlinkHelper
	helper, err = utils.NetlinkHelperFromName(client.InterfaceNamespace)
	if err != nil {
		return fmt.Errorf("failed to get netlink helper: %w", err)
	}
	defer helper.Destroy()

	var linkList []netlink.Link
	linkList, err = helper.NetlinkHandle.LinkList()
	if err != nil {
		return fmt.Errorf("failed to list interface: %w", err)
	}
	var IFList []string
	for _, link := range linkList {
		if link.Type() == "wireguard" && strings.HasPrefix(link.Attrs().Name, client.InterfacePrefix) {
			IFList = append(IFList, link.Attrs().Name)
		}
	}
	var out []byte
	out, err = liquid.NewEngine().ParseAndRender(tmpl, map[string]interface{}{"IFList": IFList, "Client": client})
	_, err = fmt.Fprint(os.Stdout, string(out))
	if err != nil {
		return err
	}
	return nil
}
