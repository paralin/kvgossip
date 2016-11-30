package util

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashToString(hash []byte) string {
	return hex.EncodeToString(hash)
}

// Get string hash.
func HexSha256(data []byte) string {
	hash := sha256.Sum256(data)
	return HashToString(hash[:])
}
