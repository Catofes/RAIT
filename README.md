## R.A.I.T. - Redundant Array of Inexpensive Tunnels
#### Disclaimer

RAIT is still in its early days of active development and **breaking changes** are expected to occur. **Linux** is the only officially supported platform.

#### About

RAIT, despite its name, is tightly tied with wireguard as the underlying transport, supporting the automated creation of full-mesh tunnels within a cluster of nodes. The **single** goal of RAIT is to form link local connectivity.

#### Configuration Files

For a reference of the configuration format, see [instance.go](rait/instance.go) and [peer.go](rait/peer.go)

#### Render

Though RAIT provides nothing beyond link local connectivity, it is intended to be used in conjunction with routing daemons to form site local connectivity, the "rait render" subcommand thus exists to generate configuration files for routing daemons, the template engine used by RAIT is [liquid](https://shopify.github.io/liquid/), and the object passed to the template engine would be:

```
LinkList []string # A list of the wireguard interfaces managed by rait
Client   Client # The exact representation of "rait.conf"
```

#### URL

RAIT loads files from 'URLs', they can be of a file system path, http(s) url, or "stdin://"
