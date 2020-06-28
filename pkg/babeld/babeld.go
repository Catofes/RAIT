package babeld

import (
	"bufio"
	"fmt"
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
	"io"
	"io/ioutil"
	"net"
	"strings"
)

type Babeld struct {
	Network string
	Address string
}

func (b *Babeld) ListList() ([]string, error) {
	conn, err := net.Dial(b.Network, b.Address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_, err = fmt.Fprint(conn, "dump\nquit\n")
	if err != nil {
		return nil, err
	}

	var interfaces []string
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), " ")
		if tokens[0] != "add" {
			continue
		}
		switch tokens[1] {
		case "interface":
			interfaces = append(interfaces, tokens[2])
		}
	}
	return interfaces, nil
}

func (b *Babeld) LinkAdd(i string) error {
	conn, err := net.Dial(b.Network, b.Address)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = fmt.Fprintf(conn, "interface %s\nquit\n", i)
	if err != nil {
		return err
	}

	_, err = io.Copy(ioutil.Discard, conn)
	if err != nil {
		return err
	}
	return nil
}

func (b *Babeld) LinkDel(i string) error {
	conn, err := net.Dial(b.Network, b.Address)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = fmt.Fprintf(conn, "flush interface %s\nquit\n", i)
	if err != nil {
		return err
	}

	_, err = io.Copy(ioutil.Discard, conn)
	if err != nil {
		return err
	}
	return nil
}

func (b *Babeld) LinkSync(target []string) error {
	current, err := b.ListList()
	if err != nil {
		return err
	}

	for _, link := range current {
		if !misc.In(target, link) {
			err = b.LinkDel(link)
			if err != nil {
				return err
			}
		}
	}

	for _, link := range target {
		if !misc.In(current, link) {
			err = b.LinkAdd(link)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
