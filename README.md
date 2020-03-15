## RAIT - Redundant Array of Inexpensive Tunnels

##### Disclaimer: RAIT is highly opinionated and only suitable for a strictly specific configuration.

###### Configuration File Format

```toml
# Fields listed below are there default values
# rait.conf
PrivateKey = ""
SendPort = 0
Interface = "raitif" # The local peer of the veth pair
Addresses = [] # The addresses to configure on the local peerm, CIDR requied
Namespace = "raitns" # the netns to move the wireguard interfaces into
IFPrefix = "rait" # the common prefix of the wireguard interfaces
MTU = 1400 # the MTU of the wireguard interfaces
FwMark = 54
Name = "os.Hostname()" # the node name will be encoded as part of a dumb address used as ICMP src

# peer.conf (it has to have the suffix ".conf")
PublicKey = ""
SendPort = 0
Endpoint = "127.0.0.1" # Optional
```

Note: outside NS denotes the network namespace in which rait is called, while inside NS denotes the network namespace specified in rait.conf. RAIT creates the wireguard interfaces in the outside NS, then moves them into the inside NS. Additionally, a veth pair is created across the two NSes, with addresses listed in rait.conf configured on the outside peer. To form Layer 3 connectivity, babeld is chosen as the routing daemon (actually, we need one instance of it inside each network namespace). A sane configuration file is generated for the one in inside NS. While for the one in the outside NS, its intentionally left unconfigured.

