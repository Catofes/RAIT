package rait

import (
	"fmt"
	"github.com/osteele/liquid"
	"github.com/vishvananda/netlink"
	"gitlab.com/NickCao/RAIT/rait/internal/utils"
	"io"
	"io/ioutil"
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

	var rawLinkList []netlink.Link
	rawLinkList, err = client.ListInterfaces()
	if err != nil {
		return nil, err
	}

	var linkList []string
	for _, link := range rawLinkList {
		linkList = append(linkList, link.Attrs().Name)
	}

	var output []byte
	output, err = liquid.NewEngine().ParseAndRender(source, map[string]interface{}{"LinkList": linkList, "Client": client})
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}
	return output, err
}
