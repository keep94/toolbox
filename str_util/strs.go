// Package str_util contains basic string utilities.
package str_util

import (
  "regexp"
  "strings"
)

var (
  re = mustre(regexp.Compile("\\s+"))
)

// Normalize normalizes a string for compare. It does by converting to
// lowercase and removing extra whitespace between words as well as
// trimming any leading or trailing whitespace.
func Normalize(s string) string {
  return re.ReplaceAllString(
      strings.TrimSpace(
          strings.ToLower(s)), " ")
}

// AutoComplete keeps track of auto-complete candidates.
type AutoComplete struct {
  // Items are the candidates so far with most recently added items at the end.
  // Clients should not modify directly.
  Items []string
  itemMap map[string]bool
}

// Add adds another auto-complete candidate. If a candidate that equals s,
// ignoring case, has already been added then s is not added again.
// Also if s is the empty string, it is not added.
func (a *AutoComplete) Add(s string) {
  if a.itemMap == nil {
    a.itemMap = map[string]bool {"": true}
  }
  lower := strings.ToLower(s)
  if !a.itemMap[lower] {
    a.itemMap[lower] = true
    a.Items = append(a.Items, s)
  }
}

func mustre(re *regexp.Regexp, err error) *regexp.Regexp {
  if err != nil {
    panic(err.Error())
  }
  return re
}

