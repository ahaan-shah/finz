package main

import (
	"crypto/rand"
	"encoding/hex"
)

// newID returns a random 32-character hex string, the Go equivalent of the
// Python side's `uuid4().hex` - same 32-hex-char shape, same purpose
// (an opaque, effectively-unique id), just generated without pulling in a
// UUID library for it.
func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
