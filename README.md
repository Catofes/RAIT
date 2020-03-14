## RAIT - Redundant Array of Inexpensive Tunnels

##### Disclaimer: RAIT is highly opinionated and only suitable for a strictly specific configuration.

###### Configuration File Format

```toml
# rait.conf
PrivateKey = "MKIbFElB6QVTjCozZny+fqTWiG5U65u/cwiPnjqQsV0="
SendPort = 12345
Interface = "rait" # The local peer of the veth pair
Addresses = ["2a0c:abcd:5678:1234::1/64","fd44:1234:5678:abcd::1/64"] # The addresses to configure on the local peer
Namespace = "raitns" # the netns to move the wireguard interfaces into
IFPrefix = "raitif" # the common prefix of the wireguard interfaces
MTU = 1400 # the MTU of the wireguard interfaces
FwMark = 54

# peer.conf (it has to have the suffix ".conf")
PublicKey = "m4UZot4m0KXtfZRLI5MoyZrVPNlMG2PvPFVrM9I+3zc="
SendPort = 50456
Endpoint = "1.1.1.1" # Optional
```

Note: outside NS denotes the network namespace in which rait is called, while inside NS denotes the network namespace specified in rait.conf. RAIT creates the wireguard interfaces in the outside NS, then moves them into the inside NS. Additionally, a veth pair is created across the two NSes, with addresses listed in rait.conf configured on the outside peer. To form Layer 3 connectivity, babeld is chosen as the routing daemon (actually, we need one instance of it inside each network namespace). A sane configuration file is generated for the one in inside NS. While for the one in the outside NS, its intentionally left unconfigured.

