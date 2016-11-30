package grant

import "github.com/fuserobotics/kvgossip/data"

func (pool *GrantAuthorizationPool) Dedupe() {
	pool.SignedGrants = data.DedupeSignedData(pool.SignedGrants)
}
