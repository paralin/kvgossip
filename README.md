K/V Gossip: Distributed Configuration
===================================

Key-Value Gossip is a mechanism for maintaining a key/value tree of arbitrary data. Each key + value pair is signed by an **entity** authorized to modify it. Entities can make changes to streams by minting a new value and broadcasting the K/V pair over the network.

An **entity** can authorize another entity to edit a key using a **grant**. The agent starts with a root public key that has permissions to read/write everything and grant anything. An entity can only grant someone else to edit a subset of their own grant pool.

**Keys** are paths, and **grants** are path globs. An example key is `/my/key/state`, an example pattern is `/my/**/sta*`.

K/V entries are timestamped, and newer timestamps always take precedence over older timestamps.

An entity can issue a **grant revocation**. This revocation prevents a node from accepting that grant anymore. Revocations will propagate across the network only when a transaction is rejected. Therefore, it is up to the user to propagate revocations reliably if necessary.

When the agent starts, it broadcasts its own local tree hash over a Serf query. Nodes that disagree will respond with their own local tree hashes. The agent then connects to each of the nodes that disagree, and the two nodes come to an agreement of the true value of the key.

When a grant is revoked, we must iterate through all the key/value metadata objects, and delete any revoked grants from their grant pools. Next, the grant authorization should be verified again. If the verification now comes back as unsatisfied, we should unset the field completely.
