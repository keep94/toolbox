// Package etag computes Etags using the encoding.gob package.
package etag

import (
  "encoding/gob"
  "hash/fnv"
)

// Etag64 computes a 64-bit etag from a pointer to an arbitrary value.
func Etag64(ptr interface{}) (tag uint64, err error) {
  h := fnv.New64a()
  e := gob.NewEncoder(h)
  if err = e.Encode(ptr); err != nil {
    return
  }
  tag = h.Sum64()
  return
}

// Etag32 computes a 32-bit etag from a pointer to an arbitrary value.
func Etag32(ptr interface{}) (tag uint32, err error) {
  h := fnv.New32a()
  e := gob.NewEncoder(h)
  if err = e.Encode(ptr); err != nil {
    return
  }
  tag = h.Sum32()
  return
}


