module gitlab.com/NickCao/RAIT

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/urfave/cli/v2 v2.1.1
	github.com/vishvananda/netlink v1.1.0
	github.com/vishvananda/netns v0.0.0-20191106174202-0a2b9b5464df
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20200205215550-e35592f146e4
)

replace github.com/vishvananda/netlink => github.com/NickCao/netlink v1.1.1
