package grant

import (
	"testing"
)

func buildGrantChain(patterns []string) []*Grant {
	res := []*Grant{}
	for _, pat := range patterns {
		res = append(res, &Grant{
			KeyRegex: pat,
		})
	}
	return res
}

func assertValidateKeyAuthorization(patterns []string, key string, shouldValid bool, t *testing.T) {
	grantChain := buildGrantChain(patterns)
	res := ValidateKeyAuthorization(grantChain, key)
	if res != shouldValid {
		mp := "should not"
		if shouldValid {
			mp = "should"
		}
		t.Fatalf("Key %s %s be valid under chain %v.", key, mp, patterns)
	}
}

func TestValidateKeyAuthorization(t *testing.T) {
	ch1 := []string{
		"/fusebot.io/r/**/*",
		"/fusebot.io/r/np1/devices/plane_1/*",
		"/fusebot.io/r/np1/devices/plane_1/tar*",
	}

	assertValidateKeyAuthorization(ch1, "/fusebot.io/r/np1/devices/plane_2/target", false, t)
	assertValidateKeyAuthorization(ch1, "/fusebot.io/r/np1/devices/plane_1/target", true, t)
	assertValidateKeyAuthorization(ch1, "/fusebot.io/r/np1/devices/plane_1/t", false, t)
	assertValidateKeyAuthorization(ch1, "/fusebot.io/r/*/devices/plane_1/target", false, t)

	ch2 := []string{
		"/fusebot.io/r/*/devices/**/*",
	}

	assertValidateKeyAuthorization(ch2, "/fusebot.io/r/np1/devices/plane_1/target", true, t)
}
