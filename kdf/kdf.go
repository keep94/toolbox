// Package kdf contains useful key derivation functions.
package kdf

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/pbkdf2"
	"io"
)

const (
	// Default repititions to use to derive a key.
	DefaultReps = 1972
)

var (
	// Default salt
	DefaultSalt = []byte{0x31, 0x41, 0x59, 0x26, 0x53, 0x59, 0xAB, 0x24}
)

// NewHMAC creates a one way hash of plain performing reps repitition.
// Resulting hash is 40 bytes and contains 8 bytes of random salt. The larger
// reps is the longer it takes to build the hash.
func NewHMAC(plain []byte, reps int) []byte {
	salt := Random(8)
	kdf := KDF(plain, salt, reps)
	result := make([]byte, len(salt)+len(kdf))
	idx := copy(result, salt)
	copy(result[idx:], kdf)
	return result
}

// VerifyHMAC returns true if mac is a valid one way hash of plain. reps
// must be the same as what was passed to NewHMAC to create the one way hash.
func VerifyHMAC(plain []byte, mac []byte, reps int) bool {
	return hmac.Equal(mac[8:], KDF(plain, mac[:8], reps))
}

// KDF derives a 32 byte encryption key from plain by using salt and reps
// repititions. The larger reps is, the longer it takes to drive the key.
// For a given plain text, salt, and reps, KDF will consistently produce
// the same encryption key.
func KDF(plain []byte, salt []byte, reps int) []byte {
	return pbkdf2.Key(plain, salt, reps, 32, sha256.New)
}

// Random produces a random sequence of count bytes
func Random(count int) []byte {
	result := make([]byte, count)
	if _, err := io.ReadFull(rand.Reader, result); err != nil {
		panic(err)
	}
	return result
}
