package babeld

import (
	"bufio"
	"fmt"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net"
	"strings"
)

type Babeld struct {
	network string
	address string
}

func NewBabeld(network, address string) *Babeld {
	return &Babeld{
		network: network,
		address: address,
	}
}

func (b *Babeld) LinkList() ([]string, error) {
	conn, err := net.Dial(b.network, b.address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to babeld control socket: %w", err)
	}
	defer conn.Close()

	_, err = fmt.Fprint(conn, "dump\nquit\n")
	if err != nil {
		return nil, fmt.Errorf("failed to write to babeld control socket: %w", err)
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
			zap.S().Debugf("found babeld interface: %s", scanner.Text())
			interfaces = append(interfaces, tokens[2])
		}
	}
	return interfaces, nil
}

func (b *Babeld) LinkAdd(i, conf string) error {
	conn, err := net.Dial(b.network, b.address)
	if err != nil {
		return fmt.Errorf("failed to connect to babeld control socket: %w", err)
	}
	defer conn.Close()

	_, err = fmt.Fprintf(conn, "interface %s %s\nquit\n", i, conf)
	if err != nil {
		return fmt.Errorf("failed to write to babeld control socket: %w", err)
	}
	zap.S().Debugf("added babeld interface: %s", i)

	_, err = io.Copy(ioutil.Discard, conn)
	if err != nil {
		return fmt.Errorf("failed to drain babeld control socket: %w", err)
	}
	return nil
}

func (b *Babeld) LinkDel(i string) error {
	conn, err := net.Dial(b.network, b.address)
	if err != nil {
		return fmt.Errorf("failed to connect to babeld control socket: %w", err)
	}
	defer conn.Close()

	_, err = fmt.Fprintf(conn, "flush interface %s\nquit\n", i)
	if err != nil {
		return fmt.Errorf("failed to write to babeld control socket: %w", err)
	}
	zap.S().Debugf("removed babeld interface: %s", i)

	_, err = io.Copy(ioutil.Discard, conn)
	if err != nil {
		return fmt.Errorf("failed to drain babeld control socket: %w", err)
	}
	return nil
}

func (b *Babeld) LinkSync(target []string, conf string) error {
	current, err := b.LinkList()
	if err != nil {
		return err
	}

	for _, link := range current {
		if !stringIn(target, link) {
			err = b.LinkDel(link)
			if err != nil {
				return err
			}
		}
	}

	for _, link := range target {
		if !stringIn(current, link) {
			err = b.LinkAdd(link, conf)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func stringIn(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}
