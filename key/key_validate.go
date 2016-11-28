package key

import "path"

// Check if a pattern is valid.
func ValidatePattern(pattern string) error {
	// the only error that can be returned from this is invalidpatternerr.
	_, err := path.Match(pattern, "/")
	return err
}

// KeyPatternContains checks if the pattern contains the key.
func KeyPatternContains(pattern, key string) bool {
	matched, err := path.Match(pattern, key)
	return err == nil && matched
}
