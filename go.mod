module gitlab.com/NickCao/RAIT/v2

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/go-playground/validator/v10 v10.3.0
	github.com/osteele/liquid v1.2.4
	github.com/osteele/tuesday v1.0.3 // indirect
	github.com/stretchr/testify v1.5.1 // indirect
	github.com/urfave/cli/v2 v2.2.0
	github.com/vishvananda/netlink v1.1.1-0.20200606011528-cf6600189038
	github.com/vishvananda/netns v0.0.0-20200520041808-52d707b772fe
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9 // indirect
	golang.org/x/sys v0.0.0-20200622214017-ed371f2e16b4
	golang.zx2c4.com/wireguard v0.0.20200320 // indirect
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20200609130330-bd2cb7843e1b
	gopkg.in/yaml.v2 v2.3.0 // indirect
)

replace golang.zx2c4.com/wireguard/wgctrl => github.com/NickCao/wgctrl-go v0.0.0-20200623070442-89366cff0bcc
