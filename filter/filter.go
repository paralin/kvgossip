package filter

import (
	"github.com/bmatcuk/doublestar"
	"strings"
)

// MatchesFilters checks if key matches filters.
func MatchesFilters(key string, filters []string) bool {
	currMatch := false
	for _, filt := range filters {
		negFilt := strings.HasPrefix(filt, "!")
		if negFilt {
			if !currMatch {
				continue
			}
			filt = filt[1:]
		} else if currMatch {
			continue
		}

		matches, _ := doublestar.Match(filt, key)
		if currMatch {
			if matches {
				currMatch = false
			}
		} else {
			currMatch = matches
		}
	}

	return currMatch
}
