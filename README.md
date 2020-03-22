## RAIT - Redundant Array of Inexpensive Tunnels - Gravity Ver.

##### Configuration File Format

###### rait.conf
```toml
PrivateKey = "MKOOS4vi0gb6U46ZwSenHK7p4XyHW/UAkUjBBF9Cz1M="
SendPort = 54632 # A port that is unique among all the nodes
Babeld = "/run/rait/babeld" # the path to write generated babeld.conf, "off" to disable
Veth = "gravity" # The local peer of the veth pair, "off" to disable
Namespace = "gravity" # the netns to move the wireguard interfaces into
IFPrefix = "rait" # the common prefix of the wireguard interfaces
MTU = 1400 # the MTU of the wireguard interfaces
FwMark = 54 # that is 0x36
ULAName = "off" # the node name will be encoded as part of a ULA, "off" to disable
```
```toml
# peer.conf (it has to have the suffix ".conf")
PublicKey = "dDhKUs11CVqDrHlYWHuJZ4Jg/39TvkkdFthCNWqPMHQ="
SendPort = 54632 # the port has the be consistent with the prior one
Endpoint = "127.0.0.1" # Optional, IP only
```
