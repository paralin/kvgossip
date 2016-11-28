package key

import (
	"errors"
	"github.com/bmatcuk/doublestar"
)

// Check if a pattern is valid.
func ValidatePattern(pattern string) error {
	// the only error that can be returned from this is invalidpatternerr.
	_, err := doublestar.Match(pattern, "/")
	if err != nil {
		return err
	}
	if len(pattern) < 2 || pattern[0] != '/' {
		return errors.New("Pattern must start with /.")
	}
	return nil
}

// KeyPatternContains checks if the pattern contains the key.
func KeyPatternContains(pattern, key string) bool {
	matched, err := doublestar.Match(pattern, key)
	return err == nil && matched
}
