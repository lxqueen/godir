package main

import (
  "errors"
)

func remove(slice []string, s int) []string {
    return append(slice[:s], slice[s+1:]...)
}

func removeFast(s []string, i int) []string {
    s[len(s)-1], s[i] = s[i], s[len(s)-1]
    return s[:len(s)-1]
}

func indexOf(s []string, i string) (int, error) {
  for p, v := range s {
        if (v == i) {
            return p, nil
        }
  }
  return -1, errors.New("Item not found in slice.")
}
