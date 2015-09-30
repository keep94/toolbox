// Package passwords provides password utilities.
package passwords

import (
  "crypto/hmac"
  "crypto/rand"
  "crypto/sha1"
  "encoding/base64"
  "golang.org/x/crypto/pbkdf2"
  "io"
)

// Password is a one-way encryption of a password.
type Password string

// New creates a new Password from a plain text password.
func New(password string) Password {
  salt := make([]byte, 8, 28)
  if _, err := io.ReadFull(rand.Reader, salt); err != nil {
    panic(err.Error())
  }
  gen := pbkdf2.Key([]byte(password), salt, 4096, 20, sha1.New)
  return Password(base64.StdEncoding.EncodeToString(append(salt, gen...)))
}

// Verify returns true if the provided plain text password matches this instance.
func (p Password) Verify(password string) bool {
  bytes, err := base64.StdEncoding.DecodeString(string(p))
  if err != nil {
    return false
  }
  if len(bytes) < 8 {
    return false
  }
  gen := pbkdf2.Key([]byte(password), bytes[:8], 4096, 20, sha1.New)
  return hmac.Equal(gen, bytes[8:])
}
