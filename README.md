## RAIT - Redundant Array of Inexpensive Tunnels

Disclaimer: RAIT's purpose is to provide linklocal full-mesh connectivity between nodes, while the specific method of node discovery or routing is out-of-sope.

###### Configuration File Format

```toml
# rait.conf
PrivateKey = "yJfVm2jfFtxW1pIAR52fJfrmbxCNlkdEUsUN7vdgnn4="
SendPort = 50123
IFPrefix = "rait"
# Fields below are optional
DummyName = "rait-local"
DummyIP = ["1.1.1.1/24","8.8.8.8/24"]
NetNS = "rait"
# peer.conf (it has to have the suffix ".conf")
PublicKey = "m4UZot4m0KXtfZRLI5MoyZrVPNlMG2PvPFVrM9I+3zc="
SendPort = 50456
Endpoint = "1.1.1.1" # Optional
```

Additionally, the "rait load" subcommand accepts a json stream from stdin to load the equivalent config files in a programmatic way. The scheme of the json stream is documented below.

```json
{
    "rait": {},
    "peers": [{},{}]
}
```
