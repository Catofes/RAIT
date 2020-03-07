## RAIT - Redundant Array of Inexpensive Tunnels

Disclaimer: RAIT's purpose is to provide linklocal full-mesh connectivity between nodes, while the specific method of node discovery or routing is out-of-sope.

###### Configuration File Format

```toml
# rait.conf
PrivateKey = <your wireguard private key>
SendPort = <as the name suggests>
# the fields bellow are optional
TagPolicy = <peering policy>
[Tags]
<key> = <value>

# peer.conf (it has to have suffix .conf)
PublicKey = <public key of the peer>
SendPort = <as the name suggests>
# the fields bellow are optional
Endpoint = <the ip address or fqdn of the peer>
[Tags]
<key> = <value>
```

Additionally, the "rait load" subcommand accepts a json stream from stdin to load the equivalent config files in a programmatic way. The scheme of the json stream is documented below.

```json
{
    "rait": {
        ...
    },
    "peers": [
        {
            ...
        },
        {
            ...
        }
    ]
}
```

