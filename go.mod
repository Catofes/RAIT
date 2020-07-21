module gitlab.com/NickCao/RAIT/v2

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/google/go-cmp v0.5.0 // indirect
	github.com/hashicorp/hcl/v2 v2.6.0
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/urfave/cli/v2 v2.2.0
	github.com/vishvananda/netlink v1.1.1-0.20200606011528-cf6600189038
	github.com/vishvananda/netns v0.0.0-20200520041808-52d707b772fe
	github.com/zclconf/go-cty v1.5.1 // indirect
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899 // indirect
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	golang.org/x/sys v0.0.0-20200720211630-cb9d2d5c5666
	golang.org/x/text v0.3.3 // indirect
	golang.zx2c4.com/wireguard v0.0.20200320 // indirect
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20200609130330-bd2cb7843e1b
)

replace golang.zx2c4.com/wireguard/wgctrl => github.com/NickCao/wgctrl-go v0.0.0-20200721052646-81817b9b0823
