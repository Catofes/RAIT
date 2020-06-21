## R.A.I.T. - Redundant Array of Inexpensive Tunnels
#### About

rait, acronym for redundant array of inexpensive tunnels, is the missing the missing piece of the puzzle, for using wireguard to create distributed overlay networks. It serves the purpose by creating point to point tunnels between all participants, forming fully-meshed link-local connectivity. Meanwhile, the site scope routing and underlying signaling mechanism employed to exchange node metadata, is out of scope for this project.

#### Operation

Due to technical limitation of wireguard, namely crypto routing, it struggles to be integrated into routing daemons, thus we takes a different approach, creating a separate interface for each peer, *abusing* wireguard as a point to point transport, opposing to it's original design. While this approach do ensures fully-meshed connectivity instead of a hub and spoke architecture, it also voids the possibility to reuse a single port for multiple peers, though the consumption of port range is negligible (after all, we have 65535 ports to waste ¯\\_(ツ)_/¯), the coordination of port usage is a challenging task. rait solves the problem with the concept of "SendPort", a unique port assigned to each node, as the destination port of all packets originated by it. To separate overlay from underlay and avoid routing loops, rait extends the fwmark and netns used by wireguard with two other means, ifgroup and vrf, both eases the management of large volume of interfaces.

#### Configuration Files

rait uses two set of configuration files, rait.conf and peers.conf, and they are all in toml format

```toml
# /etc/rait/rait.conf
PrivateKey = "+FtC0RrIEV6iIXiNyPZYDhQGWYdqb8Z30G6JB7foGWM=" # required, the private key of current node
AddressFamily = "ip4" # required, [ip4]/ip6, the address family of current node
SendPort = 53276 # required, the sending (destination) port of wireguard sockets
BindAddress = 1.1.1.1 # the local address for wireguard sockets to bind to, has no effect for now

InterfacePrefix = "mesh" # [rait], the common prefix to name the wireguard interfaces
InterfaceGroup  = 54 # [54], the ifgroup for the wireguard interfaces
MTU = 1400 # [1400], the MTU of the wireguard interfaces
FwMark = 54 # [0x36], the fwmark on packets sent by wireguard sockets

Isolation = vrf # [netns]/vrf, the isolation method to separate overlay from underlay
InterfaceNamespace = "mesh-vrf" # the netns or vrf to move wireguard interface into
TransitNamespace = "" # the netns or vrf to create wireguard sockets in

Peers string = "https://example.com/peers" # [/etc/rait/peers.conf], the url of the peer list
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
rait render <in> <out> # render template based on the desired state of the tunnels
```

#### Render

The render subcommand gathers information about the wireguard links and the interface itself, and the renders the given liquid template to be used as configuration file of routing daemons. an example for using rait together with babeld is given as bellow.

```
random-id true
export-table 254
kernel-priority 256

default type tunnel link-quality true split-horizon false
default rxcost 32 hello-interval 20 max-rtt-penalty 1024 rtt-max 1024

{% for i in LinkList %}interface {{ i }}
{% endfor %}
redistribute ip dead:beaf:f00::/44 ge 64 le 64 allow
redistribute local deny
install ip ::/0 pref-src dead:beaf:f00:ba2::1
```