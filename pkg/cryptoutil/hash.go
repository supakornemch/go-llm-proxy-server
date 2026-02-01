package cryptoutil

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashKey returns a SHA-256 hash of the input string.
func HashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
