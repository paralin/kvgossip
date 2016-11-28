package grant

import "path"

// ValidateKeyAuthorization checks if a chain of grants permit access to a key.
func ValidateKeyAuthorization(grantChain []*Grant, key string) bool {
	for _, grant := range grantChain {
		matched, err := path.Match(grant.KeyRegex, key)
		if err != nil || !matched {
			return false
		}
	}
	return true
}
