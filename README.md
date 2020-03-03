## RAIT - Redundant Array of Inexpensive Tunnels

###### Configuration File Format

```toml
# rait.conf
IFPrefix = "grv" # naming prefix to prepend to the wireguard interfaces
PrivateKey = "4CReT4TKD4AO7mYz1V6SusU0XN5HCV52/x6rhqh6uGM=" # your wireguard private key
SendPort = 50153 # as the name suggests
PeerDir = "peers" # directory of the peer.conf(s) (relative to this file)
# the fields bellow are optional
TagPolicy = "different City; same Country" # to be implemented
[Tags]
Name = "operator_hostname"
Country = "CN"
City = "SH"

# peer.conf
PublicKey = "j8Fq5NN3snH3Xv4mjyIpaNpRkqNXu9q8oar9HcFRjxA=" # public key of the peer
SendPort = 50144 # as the name suggests
# the fields bellow are optional
Endpoint = "1.1.1.1" # the ip address or fqdn of the peer
[Tags]
Name = "operator_hostname"
Country = "US"
City = "NYC"
```

