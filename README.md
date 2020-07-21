## R.A.I.T. - Redundant Array of Inexpensive Tunnels
#### About

rait, acronym for redundant array of inexpensive tunnels, is the missing the missing piece of the puzzle, for using wireguard to create distributed overlay networks. It serves the purpose by creating point to point tunnels between all participants, forming fully-meshed link-local connectivity. Meanwhile, the site scope routing and underlying signaling mechanism employed to exchange node metadata, is out of scope for this project.

#### Operation

Due to technical limitation of wireguard, namely crypto routing, it struggles to be integrated into routing daemons, thus we takes a different approach, creating a separate interface for each peer, *abusing* wireguard as a point to point transport, opposing to it's original design. While this approach do ensures fully-meshed connectivity instead of a hub and spoke architecture, it also voids the possibility to reuse a single port for multiple peers, though the consumption of port range is negligible (after all, we have 65535 ports to waste ¯\\_(ツ)_/¯), the coordination of port usage is a challenging task. rait solves the problem with the concept of "SendPort", a unique port assigned to each node, as the destination port of all packets originated by it. To separate overlay from underlay and avoid routing loops, rait extends the fwmark and netns used by wireguard with two other means, ifgroup and vrf, both eases the management of large volume of interfaces.

#### Configuration Files

rait uses two set of configuration files, rait.conf and peers.conf, and they are all in toml format

```hcl
# /etc/rait/rait.conf
private_key = "KJJXmDtAXSrMGuIJVy/2eP65gXm1PTy7vCR/4O/vEEI="
transport {
  address_family = "ip4"
  send_port = 50153
  mtu = 1420
  ifprefix = "rait4x"
  fwmark = 54
}
transport {
  address_family = "ip6"
  send_port = 50154
  mtu = 1420
  ifprefix = "rait6x"
  fwmark = 54
  dynamic_port = true
}
isolation {
  type = "netns"
  ifgroup = 54
  transit = ""
  target = "raitns"
}
babeld {
  socket_type = "unix"
  socket_addr = "/run/babeld.ctl"
}
peers = "/etc/rait/peers.conf"
```

```toml
# /etc/rait/peers.conf
[[Peers]]
PublicKey = "+Y23l1qr9oEM4WlqPvp4IG2TPXdNCa11twipDvOHT3w=" # required, the public key of the peer
SendPort = 50221 # required, the sending port of the peer
Endpoint = "example.org" # the endpoint ip address or resolvable hostname
AddressFamily = "ip4" # required, [ip4]/ip6, the address family of this node
[[Peers]]
PublicKey = "+Y23l1qr9oEM4WlqPvp4IG2TPXdNCa11twipDvOHT3w="
SendPort = 50160
AddressFamily = "ip6"
[[Peers]]
PublicKey = "+Y23l1qr9oEM4WlqPvp4IG2TPXdNCa11twipDvOHT3w="
SendPort = 50047
Endpoint = "1.1.1.1"
AddressFamily = "ip4"
```

#### URL

rait accepts the use of url in configuration files or in the command line, the url scheme is defined bellow

```bash
/some/random/path # filesystem path
- # stdin or stdout, depending on the context
https://example.com/some/file # http url
```

#### CLI

```bash
rait up # create or sync the tunnels
rait down # destroy the tunnels
```
