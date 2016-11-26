K/V Gossip: Distributed Configuration
===================================

Distributed key/value store on top of Serf, using X.509 certificates for authentication.

Key-Value Gossip is a mechanism for maintaining a key/value tree of arbitrary data. Each key + value pair is signed by an **entity** authorized to modify it. Entities can make changes to streams by minting a new value and broadcasting the K/V pair over the network.

K/V entries are timestamped, and newer timestamps always take precedence over older timestamps.

A K/V entry can never be truly deleted, as a tombstone must remain to prevent old values from being re-gossiped and re-introduced to the system. A tombstone is a K/V pair with no value.

KVGossip keeps multiple k/v trees, which is useful if you want to avoid collisions. The tree ID is a single byte (255 tree limit). 0 is reserved for authentication storage.

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

The system refuses to synchronize anything until it is certain tree 0 is synced. This is important - authorization info should be synced first, so we don't accidentally reject or accept incoming k/v changes.

If we know a key or multiple keys are out of date, we can emit a state query. We then can select nodes that have a different tree hash than us, and in the order of network coordinate distance, open up a sync session with them. We do this over and over until nobody disagrees with state.

A sync session is a connection in which two nodes synchronize out-of-date keys, agreeing on the latest valid keys on either side.

Nodes remember the hash tree states of their peers, in a sort of weak cache. This cache will quickly be abandoned though, in the case of disagreement, so it is used mainly as an optimization.

Sync Sessions
=============

A sync session starts when a node that believes it has old data connects to a node that may have new data. The following is a typical conversation:

A connects to B:

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
 - A: I disagree, that looks invalid.

In this case the nodes will continue on to other keys that may be out of sync, but will never re-start the process in the same sync session.

Local Database
==============

We need to store the following things:

 - Global bucket, with list of KV trees, list of peer KV hash trees.
 - The actual KV trees. This involves a key + key/value/signature/timestamp pair.
 - KV hash trees. Same thing as KV tree, but with a hash + timestamp as value (less expensive to read out of DB).
 - KV hash trees of peers.

Using boltdb for this.
