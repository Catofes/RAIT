package rait

import (
	"bufio"
	"bytes"
	"fmt"
	"gitlab.com/NickCao/RAIT/v2/pkg/misc"
	"go.uber.org/zap"
	"io"
	"net"
	"strings"
)

func (b *Babeld) WriteCommand(command string) (*bytes.Buffer, error) {
	conn, err := net.Dial(b.SocketType, b.SocketAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to babeld control socket: %s", err)
	}
	defer conn.Close()

	_, err = fmt.Fprintf(conn, "%s\nquit\n", command)
	if err != nil {
		return nil, fmt.Errorf("failed to write to babeld control socket: %s", err)
	}

	var buf = bytes.NewBuffer(nil)
	_, err = io.Copy(buf, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to drain babeld control socket: %s", err)
	}
	return buf, nil
}

func (b *Babeld) LinkList() ([]string, error) {
	out, err := b.WriteCommand("dump")
	if err != nil {
		return nil, err
	}

	var interfaces []string
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), " ")
		if tokens[0] != "add" {
			continue
		}
		switch tokens[1] {
		case "interface":
			zap.S().Debugf("found babeld interface: %s", scanner.Text())
			interfaces = append(interfaces, tokens[2])
		}
	}
	return interfaces, nil
}

func (b *Babeld) LinkAdd(name string) error {
	_, err := b.WriteCommand(fmt.Sprintf("interface %s %s", name, b.Param))
	return err
}

func (b *Babeld) LinkDel(name string) error {
	_, err := b.WriteCommand(fmt.Sprintf("flush interface %s", name))
	return err
}

func (b *Babeld) LinkSync(target []string) error {
	current, err := b.LinkList()
	if err != nil {
		return err
	}

	for _, link := range current {
		if !misc.StringIn(target, link) {
			err = b.LinkDel(link)
			if err != nil {
				return err
			}
		}
	}

	for _, link := range target {
		if !misc.StringIn(current, link) {
			err = b.LinkAdd(link)
			if err != nil {
				return err
			}
		}
	}

	_, err = b.WriteCommand(b.ExtraCmd)
	return err
}

