package grant

import (
	kp "github.com/fuserobotics/kvgossip/key"
)

// ValidateKeyAuthorization checks if a chain of grants permit access to a key.
func ValidateKeyAuthorization(grantChain []*Grant, key string) bool {
	for _, grant := range grantChain {
		if !kp.KeyPatternContains(grant.KeyRegex, key) {
			return false
		}
	}
	return true
}
