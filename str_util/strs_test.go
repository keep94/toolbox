package str_util

import (
  "reflect"
  "testing"
)

func TestNormalize(t *testing.T) {
  if output := Normalize(" bE   You "); output != "be you" {
    t.Errorf("Expected 'be you', got %v", output)
  }
}

func TestAutoComplete(t *testing.T) {
  ac := AutoComplete{}
  ac.Add("")  // Should be ignored
  ac.Add("Hello")
  ac.Add("there")
  ac.Add("hEllo") // Should be ignored, already "Hello"
  ac.Add("you")
  expected := []string{"Hello", "there", "you"}
  if !reflect.DeepEqual(expected, ac.Items) {
    t.Errorf("Expected %v, got %v", expected, ac.Items)
  }
}
