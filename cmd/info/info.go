package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/Catofes/RAIT/pkg/misc"
	"github.com/Catofes/RAIT/pkg/rait"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type app struct {
	listen string
	url    string
	path   string
}

type peerInfo struct {
	Name             string
	RouteID          string
	Wg4Address       string
	Wg6Address       string
	Vxlan4Address    string
	Vxlan6Address    string
	AnnouncedAddress []string
}

func (s *app) get(ctx echo.Context) error {
	peers, err := rait.NewPeers(s.url, nil)
	if err != nil {
		ctx.Error(err)
		return err
	}
	infos := make(map[string]peerInfo, 0)
	for _, peer := range peers {
		if _, ok := infos[s.generateRouteID(peer)]; !ok {
			infos[s.generateRouteID(peer)] = peerInfo{}
		}
		info := infos[s.generateRouteID(peer)]
		info.Name = peer.Name
		info.RouteID = s.generateRouteID(peer)
		peer.GenerateInnerAddress().String()
		switch peer.Endpoint.AddressFamily {
		case "ip4":
			info.Wg4Address = peer.Endpoint.InnerAddress
			info.Vxlan4Address = misc.NewLLAddrFromMac(peer.GenerateMac()).String()
		case "ip6":
			info.Wg6Address = peer.Endpoint.InnerAddress
			info.Vxlan6Address = misc.NewLLAddrFromMac(peer.GenerateMac()).String()
		}
		info.AnnouncedAddress = make([]string, 0)
		infos[s.generateRouteID(peer)] = info
	}
	babel := rait.Babeld{
		SocketType: "unix",
		SocketAddr: s.path,
	}
	dump, err := babel.WriteCommand("dump")
	if err != nil {
		ctx.Error(err)
		return err
	}
	scanner := bufio.NewScanner(dump)
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), " ")
		if tokens[0] == "add" && tokens[1] == "route" && tokens[8] == "yes" {
			id := strings.Trim(tokens[10], " ")
			if info, ok := infos[id]; ok {
				info.AnnouncedAddress = append(info.AnnouncedAddress, tokens[4])
				infos[id] = info
			}
		}
	}

	data := make(map[string]peerInfo)
	for _, v := range infos {
		if v.Vxlan4Address == "" {
			v.Vxlan4Address = "-"
		}
		if v.Vxlan6Address == "" {
			v.Vxlan6Address = "-"
		}
		if v.Wg4Address == "" {
			v.Wg4Address = "-"
		}
		if v.Wg6Address == "" {
			v.Wg6Address = "-"
		}
		data[v.Name] = v
	}

	t := template.Must(template.New("").Parse(`
<html>
<head>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@4.5.3/dist/css/bootstrap.min.css" integrity="sha384-TX8t27EcRE3e/ihU7zmQxVncDAy5uIKz4rEkgIXeMed4M0jlfIDPvg6uqKI2xXr2" crossorigin="anonymous">
</head>
<body>
<div class="container">
<table class="table table-hover table-sm">
</thead>
<tr>
<td>Name</td>
<td>RouteID</td>
<td>Wireguard(4/6)</td>
<td>VXLAN(4/6)</td>
<td>AnnouncedAddress</td>
</tr>
</thead>
{{range .}}<tr>
<td>{{.Name}}</td>
<td>{{.RouteID}}</td>
<td>{{.Wg4Address}}<br>{{.Wg6Address}}</td>
<td>{{.Vxlan4Address}}<br>{{.Vxlan6Address}}</td>
<td>
{{range .AnnouncedAddress}}
{{.}}
<br>
{{end}}
</td>
</tr>{{end}}
</table>
</div>
</body></html>`))
	body := bytes.Buffer{}
	if err := t.Execute(&body, data); err != nil {
		log.Fatal(err)
	}
	ctx.HTML(200, body.String())
	return nil
}

func (s *app) generateRouteID(peer rait.Peer) string {
	hash := md5.Sum([]byte(peer.PublicKey + "\n"))
	id := hex.EncodeToString(hash[:])
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%s",
		id[0:2], id[2:4], id[4:6], id[6:8], id[8:10], id[10:12], id[12:14], id[12:14])
}

func (s *app) run() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/", s.get)
	e.Logger.Fatal(e.Start(s.listen))
}

func main() {
	url := flag.String("u", "https://www.catofes.com/higgs.hcl", "peers url")
	socket := flag.String("c", "/run/higgs.ctl", "babeld socket")
	listen := flag.String("l", "0.0.0.0:80", "listen address")
	flag.Parse()
	a := app{
		url:    *url,
		path:   *socket,
		listen: *listen,
	}
	a.run()
}
