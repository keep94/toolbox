package lockout_test

import (
  "testing"
  "github.com/keep94/appcommon/lockout"
)

func TestNil(t *testing.T) {
  var l *lockout.Lockout
  // A nil pointer means infinite failures needed for lockout
  assertEquals(t, false, l.Failure("alice"))
  assertEquals(t, false, l.Failure("alice"))
  assertEquals(t, false, l.Failure("alice"))
  assertEquals(t, false, l.Locked("alice"))
  l.Success("alice")
  assertEquals(t, false, l.Locked("alice"))
}


func TestAPI(t *testing.T) {
  l := lockout.New(3)
  // A success clears failures if limit hasn't been reached
  assertEquals(t, false, l.Failure("alice"))
  assertEquals(t, false, l.Failure("alice"))
  assertEquals(t, false, l.Locked("alice"))
  l.Success("alice")
  assertEquals(t, false, l.Locked("alice"))

  // Reaching limit locks account
  assertEquals(t, false, l.Failure("alice"))
  assertEquals(t, false, l.Failure("alice"))
  assertEquals(t, false, l.Locked("alice"))
  assertEquals(t, true, l.Failure("alice"))
  assertEquals(t, true, l.Locked("alice"))

  // Once account locked, it stays locked
  l.Success("alice")
  assertEquals(t, true, l.Locked("alice"))

  // Failure returns true only once per account
  assertEquals(t, false, l.Failure("alice"))
  assertEquals(t, true, l.Locked("alice"))

  // Other accounts still unlocked
  assertEquals(t, false, l.Locked("charlie"))

  // But other accounts can lock too
  assertEquals(t, false, l.Failure("charlie"))
  assertEquals(t, false, l.Failure("charlie"))
  assertEquals(t, false, l.Locked("charlie"))
  assertEquals(t, true, l.Failure("charlie"))
  assertEquals(t, true, l.Locked("charlie"))
}

func assertEquals(t *testing.T, expected, actual bool) {
  if expected != actual {
    t.Errorf("Expected %v, got %v", expected, actual)
  }
}
