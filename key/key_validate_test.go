package key

import (
	"testing"
)

func TestValidKeys(t *testing.T) {
	ExpectKeyValid("/", false, t)
	ExpectKeyValid("", false, t)
	ExpectKeyValid(" /test/test.json", false, t)
	ExpectKeyValid("/test/te st.json", false, t)
	ExpectKeyValid("/test/", false, t)
	ExpectKeyValid("/test", true, t)
	ExpectKeyValid("/test.com", true, t)
	ExpectKeyValid("/test.com/testing/test2.json", true, t)
	ExpectKeyValid("/test.com/testing/", false, t)
}

func ExpectKeyValid(key string, valid bool, t *testing.T) {
	verb := " not"
	if valid {
		verb = ""
	}
	actual := IsValidKey(key)
	if actual != valid {
		t.Errorf("Expected '%s' to%s be valid.", key, verb)
		t.Fail()
	}
}
