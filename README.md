## R.A.I.T. - Redundant Array of Inexpensive Tunnels
#### About

rait, acronym for redundant array of inexpensive tunnels, is the missing the missing piece of the puzzle, for using wireguard to create distributed overlay networks. It serves the purpose by creating point to point tunnels between all participants, forming fully-meshed link-local connectivity. Meanwhile, the site scope routing and underlying signaling mechanism employed to exchange node metadata, is out of scope for this project.

#### Operation

Due to technical limitation of wireguard, namely crypto routing, it struggles to be integrated into routing daemons, thus we takes a different approach, creating a separate interface for each peer, *abusing* wireguard as a point to point transport, opposing to it's original design. While this approach do ensures fully-meshed connectivity instead of a hub and spoke architecture, it also voids the possibility to reuse a single port for multiple peers, though the consumption of port range is negligible (after all, we have 65535 ports to waste ¯\\_(ツ)_/¯), the coordination of port usage is a challenging task. rait solves the problem with the concept of "SendPort", a unique port assigned to each node, as the destination port of all packets originated by it. To separate overlay from underlay and avoid routing loops, rait extends the fwmark and netns used by wireguard with two other means, ifgroup and vrf, both eases the management of large volume of interfaces.

#### Configuration Files

TODO

