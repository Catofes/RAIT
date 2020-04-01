## R.A.I.T. - Redundant Array of Inexpensive Tunnels
#### Disclaimer

RAIT is still in its early days of active development and **breaking changes** are expected to occur. **Linux** is the only officially supported platform.

#### About

RAIT, despite its name, is tightly tied with wireguard as the underlying transport, supporting the automated creation of full-mesh tunnels within a cluster of nodes. The **single** goal of RAIT is to form link local connectivity.

#### Architecture

As wireguard natively supports Linux network namespace for the isolation of overlay and underlay traffic, RAIT moves all the wireguard interfaces into a denoted namespace, leaving the supporting UDP sockets in the **calling** namespace, and optionally creates a veth pair connecting the two namespaces.

#### Configuration Files

RAIT uses two sets of configuration files. "rait.conf" is the private part, residing on individual nodes holding the wireguard private key, as well as node specific configurations. "peer.conf" is the public part, with the information just enough for other nodes to form a connection to the publishing node, one common practice is for the participants to exchange "peer.conf"(s) via a git repo.

###### rait.conf
```toml
PrivateKey = "MKOOS4vi0gb6U46ZwSenHK7p4XyHW/UAkUjBBF9Cz1M="
SendPort = 54632 # A port that is unique among all the nodes in the cluster
IFPrefix = "rait" # the common prefix of the wireguard interfaces
MTU = 1400 # the MTU of the wireguard interfaces, as well as the veth pair if enabled

# Fields below are optional, set to "off" to disable related feature
Namespace = "rait" # the netns to move the wireguard interfaces into
Veth = "rait" # The local peer of the veth pair, the other peer will be named "host"
FwMark = 54 # the fwmark assigned to all wireguard generated packets
```
###### peer.conf
```toml
PublicKey = "dDhKUs11CVqDrHlYWHuJZ4Jg/39TvkkdFthCNWqPMHQ="
SendPort = 54632 # the port has the be consistent with the prior one
Endpoint = "127.0.0.1" # Optional, IP only
```

#### Render

Though RAIT provides nothing beyond link local connectivity, it is intended to be used in conjunction with routing daemons to form site local connectivity, the "rait render" subcommand thus exists to generate configuration files for routing daemons, the template engine used by RAIT is [liquid](https://shopify.github.io/liquid/), and the object passed to the template engine would be:

```
IFList []string # A list of the wireguard interfaces created in the specified netns
RAIT   RAIT # The exact representation of "rait.conf"
```

An example babeld.conf.template can be found in misc/