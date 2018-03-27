package main

import (
    "strings"
    "errors"
    "io/ioutil"
)

// Literally just abstractions of the strings.replace method for readability.

// Modifies a copy.
func SubTag(raw string, tag string, new string) string {
  o := strings.Replace(raw, tag, new, -1)
  return o
}

// Function to safely and abstractly load template files.
func LoadFile(path string) ([]byte, error) {

  // check if file exists
  _, err := os.Stat(path)
  if err != nil { return "", errors.New("File is missing: ", path) }

  data, err := ioutil.ReadFile(path)
  if err != nil { return "", err }

  return data, nil
}


type FileAsyncOutput struct {
  data []byte
  err  error
}

// Function to safely and abstractly load template files.
func LoadFileAsync(path string, out chan FileAsyncOutput {

  // check if file exists
  _, err := os.Stat(path)
  if err != nil {
    out <- FileAsyncOutput{"", errors.New("File is missing: ", path)}
    return
  }

  data, err := ioutil.ReadFile(path)
  if err != nil {
    out <- FileAsyncOutput{"", err}
    return
  }

  out <- FileAsyncOutput{data, nil}
}
