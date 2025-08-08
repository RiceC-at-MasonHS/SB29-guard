// Package hash provides simple hashing utilities.
package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256Hex returns the hex-encoded sha256 of the input bytes
func SHA256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
