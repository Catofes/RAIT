module gitlab.com/NickCao/RAIT

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/mitchellh/mapstructure v1.3.0 // indirect
	github.com/osteele/liquid v1.2.4
	github.com/osteele/tuesday v1.0.3 // indirect
	github.com/pelletier/go-toml v1.7.0 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.5.1 // indirect
	github.com/vishvananda/netlink v1.1.0
	github.com/vishvananda/netns v0.0.0-20191106174202-0a2b9b5464df
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37 // indirect
	golang.org/x/net v0.0.0-20200506145744-7e3656a0809f // indirect
	golang.org/x/sys v0.0.0-20200511232937-7e40ca221e25
	golang.zx2c4.com/wireguard v0.0.20200320 // indirect
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20200511024508-91d9787b944f
	gopkg.in/ini.v1 v1.56.0 // indirect
)

replace github.com/vishvananda/netlink => github.com/NickCao/netlink v1.1.1
