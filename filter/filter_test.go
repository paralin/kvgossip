package filter

import (
	"testing"
)

func TestFilter(t *testing.T) {
	Expect("/test.fb.io/", []string{"**"}, true, t)
	Expect("/test.fb.io/", []string{"!**"}, false, t)
	Expect("/test.fb.io/", []string{"**", "!**"}, false, t)
	Expect("/test.fb.io/", []string{"**", "!**", "/*fb*/"}, true, t)
}

func Expect(key string, pats []string, expected bool, t *testing.T) {
	act := MatchesFilters(key, pats)
	if act != expected {
		verb := "not "
		if expected {
			verb = ""
		}
		t.Errorf("Expected %s to %smatch %v!", key, verb, pats)
		t.Fail()
	}
}
