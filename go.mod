module github.com/Catofes/RAIT/v4

go 1.14

replace golang.zx2c4.com/wireguard/wgctrl => github.com/NickCao/wgctrl-go v0.0.0-20200721052646-81817b9b0823

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/Catofes/netlink v1.2.2
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/hashicorp/hcl/v2 v2.6.0
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.3.0 // indirect
	github.com/urfave/cli/v2 v2.2.0
	github.com/vishvananda/netlink v1.1.1-0.20200606011528-cf6600189038 // indirect
	github.com/vishvananda/netns v0.0.0-20200728191858-db3c7e526aae
	go.uber.org/zap v1.16.0
	golang.org/x/sys v0.0.0-20200930185726-fdedc70b468f
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-00010101000000-000000000000
)
