package main

import (
    "strings"
)

// Literally just abstractions of the strings.replace method for readability.

// Modifies a copy.
func SubTag(raw string, tag string, new string) string {
  o := strings.Replace(raw, tag, new, -1)
  return o
}
