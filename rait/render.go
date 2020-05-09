package rait

import (
	"fmt"
	"github.com/osteele/liquid"
	"github.com/vishvananda/netlink"
	"gitlab.com/NickCao/RAIT/rait/internal/utils"
	"io"
	"io/ioutil"
	"strings"
)

// RenderTemplate gathers information about interfaces and renders the liquid template
func (client *Client) RenderTemplate(path string) ([]byte, error) {
	var reader io.ReadCloser
	var err error
	reader, err = utils.ReaderFromPath(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	var source []byte
	source, err = ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}


	var LinkList []string
	err = utils.WithNetns(client.InterfaceNamespace, func(handle *netlink.Handle) (err error) {
		var rawLinkList []netlink.Link
		rawLinkList, err = handle.LinkList()
		if err != nil {
			err = fmt.Errorf("failed to list interface: %w", err)
			return
		}
		for _, link := range rawLinkList {
			if link.Type() == "wireguard" && strings.HasPrefix(link.Attrs().Name, client.InterfacePrefix) {
				LinkList = append(LinkList, link.Attrs().Name)
			}
		}
		return
	})
	if err != nil {
		return nil, err
	}

	var output []byte
	output, err = liquid.NewEngine().ParseAndRender(source, map[string]interface{}{"LinkList": LinkList, "Client": client})
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}
	return output, err
}
