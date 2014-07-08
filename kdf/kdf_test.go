package kdf_test

import (
  "crypto/hmac"
  "github.com/keep94/appcommon/kdf"
  "testing"
)

func TestHMAC(t *testing.T) {
  firstMac := kdf.NewHMAC([]byte("aardvark"), kdf.DefaultReps)
  secondMac := kdf.NewHMAC([]byte("aardvark"), kdf.DefaultReps)
  if hmac.Equal(firstMac, secondMac) {
    t.Error("Macs should not be equal")
  }
  if !kdf.VerifyHMAC([]byte("aardvark"), firstMac, kdf.DefaultReps) {
    t.Error("Mac should have verified")
  }
  if !kdf.VerifyHMAC([]byte("aardvark"), secondMac, kdf.DefaultReps) {
    t.Error("Second Mac should have verified")
  }
  if kdf.VerifyHMAC([]byte("be"), firstMac, kdf.DefaultReps) {
    t.Error("Mac should not have verified")
  }
  if kdf.VerifyHMAC([]byte("be"), secondMac, kdf.DefaultReps) {
    t.Error("Second Mac should not have verified")
  }
}

func TestKDF(t *testing.T) {
  kdf1 := kdf.KDF([]byte("aardvark"), kdf.DefaultSalt, kdf.DefaultReps)
  kdf2 := kdf.KDF([]byte("aardvark"), kdf.DefaultSalt, kdf.DefaultReps)
  if !hmac.Equal(kdf1, kdf2) {
    t.Error("Expected kdf's to be equal")
  }
  if hmac.Equal(kdf1, kdf.KDF([]byte("sailboat"), kdf.DefaultSalt, kdf.DefaultReps)) {
    t.Error("Expected kdf's not to be equal")
  }
  if len(kdf1) != 32 {
    t.Error("Expected key to be 32 bytes")
  }
}
  
