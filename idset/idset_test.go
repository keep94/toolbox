package idset_test

import (
	"github.com/keep94/appcommon/idset"
	"testing"
)

func TestNew(t *testing.T) {
	result := idset.New(map[int64]bool{
		3: true, 13: true, 9: true, 2: true, 26: false})
	if result != "2,3,9,13" {
		t.Errorf("Expected 2,3,9,13 got %s", result)
	}
	result = idset.New(nil)
	if result != "" {
		t.Errorf("Expected empty string, got %s", result)
	}
}

func TestMap(t *testing.T) {
	var set idset.IdSet = "2,3,9"
	if !set.Contains(2) {
		t.Error("Expected set to contain 2")
	}
	if set.Contains(5) {
		t.Error("Expected set not to contain 5")
	}
	m, err := set.Map()
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 3 {
		t.Error("Expected map length to be 3")
	}
	if !m[9] {
		t.Error("Expected map to contain 9")
	}
	if m[4] {
		t.Error("Expected map not to contain 4")
	}

	set = "73"
	m, err = set.Map()
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 1 {
		t.Error("Expected map length to be 1")
	}
	if !m[73] {
		t.Error("Expected map to contain 73")
	}

	set = "hello there"
	if set.Contains(21) {
		t.Error("Expected set not to contain 21")
	}
	_, err = set.Map()
	if err == nil {
		t.Error("Expected error to be thrown")
	}

	set = ""
	m, err = set.Map()
	if err != nil {
		t.Error("Expected no error")
	}
	if len(m) > 0 {
		t.Error("Expected map length to be 0")
	}
}
