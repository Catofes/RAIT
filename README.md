## RAIT - Redundant Array of Inexpensive Tunnels

Disclaimer: RAIT is highly opinionated in its tech stack and only suitable for a strictly  specific configuration.

###### Configuration File Format

```toml
# rait.conf
PrivateKey = "yJfVm2jfFtxW1pIAR52fJfrmbxCNlkdEUsUN7vdgnn4="
SendPort = 50123
Interface = "rait"
Addresses = ["1.1.1.1/24","8.8.8.8/24"]
Namespace = "raitns"

# peer.conf (it has to have the suffix ".conf")
PublicKey = "m4UZot4m0KXtfZRLI5MoyZrVPNlMG2PvPFVrM9I+3zc="
SendPort = 50456
Endpoint = "1.1.1.1" # Optional
```