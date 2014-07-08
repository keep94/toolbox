// Package kdf contains useful key derivation functions.
package kdf

import (
  "crypto/hmac"
  "crypto/rand"
  "crypto/sha256"
  "io"
)
  
const (
  // Default repititions to use to derive a key.
  DefaultReps = 1972
)

var (
  // Default salt
  DefaultSalt = []byte{0x33, 0x2E, 0x31, 0x34, 0x31, 0x35, 0x39, 0x33}
)

// NewHMAC creates a one way hash of plain performing reps repitition.
// Resulting hash is 40 bytes and contains 8 bytes of random salt. The larger
// reps is the longer it takes to build the hash.
func NewHMAC(plain []byte, reps int) []byte {
  salt := Random(8)
  kdf := KDF(plain, salt, reps)
  result := make([]byte, len(salt) + len(kdf))
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
  mac := hmac.New(sha256.New, plain)
  result := salt
  var counter [4]byte
  for i := 0; i < reps; i++ {
    mac.Write(serializeint(i, counter[:]))
    mac.Write(result)
    result = mac.Sum(nil)
    mac.Reset()
  }
  return result
}

// Random produces a random sequence of count bytes
func Random(count int) []byte {
  result := make([]byte, count)
  io.ReadFull(rand.Reader, result)
  return result
}
  
func serializeint(i int, serial []byte) []byte {
  serial[0] = byte((i >> 24) & 0xFF)
  serial[1] = byte((i >> 16) & 0xFF)
  serial[2] = byte((i >> 8) & 0xFF)
  serial[3] = byte(i & 0xFF)
  return serial
}
