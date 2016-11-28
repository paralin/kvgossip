package key

import "path"

// Check if a pattern is valid.
func ValidatePattern(pattern string) error {
	// the only error that can be returned from this is invalidpatternerr.
	_, err := path.Match(pattern, "/")
	return err
}
