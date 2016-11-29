package util

import (
	"crypto/sha256"
	"encoding/base64"
)

// Get string hash.
func Base64Sha256(data []byte) string {
	hash := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}
