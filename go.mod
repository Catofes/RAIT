module gitlab.com/NickCao/RAIT

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/osteele/liquid v1.2.4
	github.com/osteele/tuesday v1.0.3 // indirect
	github.com/stretchr/testify v1.5.1 // indirect
	github.com/urfave/cli/v2 v2.2.0
	github.com/vishvananda/netlink v1.1.0
	github.com/vishvananda/netns v0.0.0-20191106174202-0a2b9b5464df
	golang.org/x/crypto v0.0.0-20200422194213-44a606286825 // indirect
	golang.org/x/net v0.0.0-20200421231249-e086a090c8fd // indirect
	golang.org/x/sys v0.0.0-20200420163511-1957bb5e6d1f
	golang.zx2c4.com/wireguard v0.0.20200320 // indirect
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20200324154536-ceff61240acf
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

replace github.com/vishvananda/netlink => github.com/NickCao/netlink v1.1.1
