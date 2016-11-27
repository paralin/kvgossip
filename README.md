K/V Gossip: Distributed Configuration
===================================

Distributed key/value store on top of Serf, using a similar mechanism to X.509 certificates for authentication.

Key-Value Gossip is a mechanism for maintaining a key/value tree of arbitrary data. Each key + value pair is signed by an **entity** authorized to modify it. Entities can make changes to streams by minting a new value and broadcasting the K/V pair over the network.

K/V entries are timestamped, and newer timestamps always take precedence over older timestamps.

A K/V entry can never be truly deleted, as a tombstone must remain to prevent old values from being re-gossiped and re-introduced to the system. A tombstone is a K/V pair with no value.

Authorization
=============

An authorization "grant" is an object signed by an entity that grants permission to another entity to edit a set of keys.

All grant chains begin at the root key. There is a single root key given to KVGossip in any case.

Grants can be regular expressions: `/fusebot.io/r/np1/*`, `/fusebot.io/r/*/devices/*/autopilot/target_state`. Grants have an associated tree number.

A grant has an identifier, which is just a hash of the grant object, which contains:

 - Grant regex: `/fusebot.io/r/np1/*`
 - Flags (see below)

Grants form a tree of permissions. The root node in the tree is always the hardcoded "authority" key, like a Certificate Authority in the X.509 system.

Grants can have a set of flags:

 - Subgrant permission: is the grant allowed to issue sub-grants?

At runtime, the system loads in all of the grants it has available. It then builds a graph of currently valid grants. No grant can ever point to the root key. The system then grabs the root key, forming a tree with the root node as the authority key. Later on, when attempting to edit something, the system computes the shortest possible grant chain that will allow an edit, and uses this chain to make an edit.

Grants may only give permission to edit more restrictive subsets of themselves.

Grant revocations are synced globally, while grants themselves are lazy synced.

When a key is changed, in the sync session between two nodes, when a key is changed, the node with the newer key provides both the new value and the grant chain that authorizes it. If the other node has a revocation that prevents accepting this change, it replies with the revocation. In this way the two are synced.

Synchronization
===============

Serf provides two tools we can use:

 - Events: fire-and-forget, reliable, less than 1kb (short!!) payload, eventually consistent, with retries
 - Queries: request-response, less than 1kb, unreliable, short-lived, not consistent, no retries.

We use these tools to do the following operations:

 - Change notification: when a key changes, emit a notification of the change, so others can mark the key as potentially outdated.
 - State query: query nearby nodes for a hash of the entire tree of KV hashes.
 - Node sync event: a node emits a signed hash + timestamp of its entire tree (in two events).

When we first start up, or first join, we send a full state query. This could also happen periodically.

If we know a key or multiple keys are out of date, we can emit a state query. We then can select nodes that have a different tree hash than us, and in the order of network coordinate distance, open up a sync session with them. We do this over and over until nobody disagrees with state.

A sync session is a connection in which two nodes synchronize out-of-date keys, agreeing on the latest valid keys on either side.

Nodes remember the hash tree states of their peers, in a sort of weak cache. This cache will quickly be abandoned though, in the case of disagreement, so it is used mainly as an optimization.

Sync Sessions
=============

A sync session starts when a node that believes it has old data connects to a node that may have new data. The following is a typical conversation:

A connects to B:

Grant synch step:

 - A: my grant tree hash is ABCD
 - B: my grant tree hash is DEFG ...

Then key synchronization starts:

 - A: my overall tree hash is ABCD...
 - B: my overall tree hash is DEFG... (if they agree, they will disconnect here.)

Foreach key in A, in order of newest changed -> oldest changed:

 - A: my hash of key H is DEFG
 - B: ok

Until eventually...

 - A: my hash of key J is FFFA
 - B: mine is different, timestamp is 1000
 - A: mine is newer, here it is (timestamp + key + value bundle).

OR

 - A: my hash of key J is FFFA
 - B: mine is different, timestamp is 1000
 - A: yours is newer, please send yours.
 - B: here is mine
 - A: my overall tree hash is ABCD (process restarts...)

Or otherwise (if there's a disagreement):

 - A: my hash of key J is FFFA
 - B: mine is different, timestamp is 1000
 - A: yours is newer, please send yours.
 - B: here is mine
 - A: I disagree, here's the revocation to prove it. Also, here's my version of the key (or null if not exists).

In this case the nodes will continue on to other keys that may be out of sync, but will never re-start the process in the same sync session (unless the grant was valid).

In the case that the nodes enter an infinite loop of disagreement, there is a built in maximum loop count for a sync session (3 on default) after which it will terminate the session and move on to another node.

Local Database
==============

We need to store the following things:

 - Global bucket, with list of KV trees, list of peer KV hash trees.
 - The actual KV trees. This involves a key + key/value/signature/timestamp pair.
 - KV hash trees. Same thing as KV tree, but with a hash + timestamp as value (less expensive to read out of DB).
 - KV hash trees of peers.

Using boltdb for this.

We store for each key in the DB, the original message we received with the value, which contains:

 - Value data
 - Set verification object, which contains:
  - key
  - Signature of hash of value data
  - Public key of entity that performed the action
  - Pool of Signed Grant objects from which chains can be built to verify the action.

This is also the exact object we transmit when we sync in a sync session.

We store the data in two places. We store the value itself in one entry, and the verification / hash data in another.

Grant Revocation
================

When a grant is revoked, everyone in the network that knows about the revocation will no longer accept any new OR OLD values that were authorized by that grant. This means that the system will converge first to null, and then to the oldest valid change to the key still remaining somewhere in the network, if such an entry exists.

The safest way to revoke a grant is to take the following steps simultaneously:

 - For every key that would become null following the revocation, transmit a new SET operation overwriting the key with identical data.
 - Transmit the revocation

Some nodes may set the key to null if they get the revocation before the new set operation... But applications should handle this possibility.
