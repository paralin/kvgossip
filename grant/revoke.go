package grant

import "github.com/fuserobotics/kvgossip/data"

type RevocationChecker interface {
	GetRevocation(grantData []byte) *data.SignedData
}

type GrantStore interface {
	GetAllGrants() []*data.SignedData
}
