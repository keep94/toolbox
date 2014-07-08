package passwords

import (
  "testing"
)

func TestVerify(t *testing.T) {
  
  if !New("boo").Verify("boo") {
    t.Error("Password did not verify")
  }
  if New("boo").Verify("foo") {
    t.Error("Password should not have verified.")
  }
}

func TestZeroValue(t *testing.T) {
  var p Password
  if p.Verify("foo") {
    t.Error("Zero value of Password should not verify against anything.")
  }
}

func TestSalt(t *testing.T) {
  if New("foo") == New("foo") {
    t.Error("Hash should be different every time.")
  }
}
