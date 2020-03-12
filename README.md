## RAIT - Redundant Array of Inexpensive Tunnels

##### Disclaimer: RAIT is highly opinionated and only suitable for a strictly specific configuration.

###### Configuration File Format

```toml
# rait.conf
PrivateKey = "yJfVm2jfFtxW1pIAR52fJfrmbxCNlkdEUsUN7vdgnn4="
SendPort = 50123
Interface = "rait"
Addresses = ["1.1.1.1/24"]
Namespace = "raitns"

# peer.conf (it has to have the suffix ".conf")
PublicKey = "m4UZot4m0KXtfZRLI5MoyZrVPNlMG2PvPFVrM9I+3zc="
SendPort = 50456
Endpoint = "1.1.1.1" # Optional
```

Note: outside NS denotes the network namespace in which rait is called, while inside NS denotes the network namespace specified in rait.conf.

RAIT creates the wireguard interfaces in the outside NS, then moves them into the inside NS. Additionally, a veth pair is created across the two NSes, with addresses listed in rait.conf configured on the outside peer.

To form Layer 3 connectivity, babeld is chosen as the routing daemon (actually, we need one instance of it inside each network namespace). A sane configuration file is generated for the one in inside NS, and you can start the corresponding babeld with the following command.

```
sudo ip netns exec <ns> babeld -c /var/run/babeld.rait.conf
```

While for the one in the outside NS, its intentionally left unconfigured.