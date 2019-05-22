// Package idset handles comma separated list of int64.
package idset

import (
  "sort"
  "strconv"
  "strings"
)

// IdSet is a comma separated set of record Ids.
type IdSet string

// Contains returns true if this set contains id
func (s IdSet) Contains(id int64) bool {
  m, err := s.Map()
  if err != nil {
    return false
  }
  return m[id]
}

// Map converts this set to a map.
func (s IdSet) Map() (map[int64]bool, error) {
  if s == "" {
    return map[int64]bool{}, nil
  }
  strs := strings.Split(string(s), ",")
  ids := make([]int64, len(strs))
  for i := range ids {
    var err error
    ids[i], err = strconv.ParseInt(strs[i], 10, 64)
    if err != nil {
      return map[int64]bool{}, err
    }
  }
  return toMap(ids), nil
}

// New creates a new IdSet from given ids.
func New(ids map[int64]bool) IdSet {
  return newIdSet(ids)
}

func newIdSet(m map[int64]bool) IdSet {
  ids := make(int64Slice, 0, len(m))
  for id, ok := range m {
    if ok {
      ids = append(ids, id)
    }
  }
  sort.Sort(ids)
  strs := make([]string, len(ids))
  for i := range strs {
    strs[i] = strconv.FormatInt(ids[i], 10)
  }
  return IdSet(strings.Join(strs, ","))
}

func toMap(ids []int64) map[int64]bool {
  result := make(map[int64]bool, len(ids))
  for _, id := range ids {
    result[id] = true
  }
  return result
}

type int64Slice []int64

func (s int64Slice) Len() int {
  return len(s)
}

func (s int64Slice) Swap(i, j int) {
  s[i], s[j] = s[j], s[i]
}

func (s int64Slice) Less(i, j int) bool {
  return s[i] < s[j]
}
