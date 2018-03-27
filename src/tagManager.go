package main

import (
    "strings"
    "errors"
    "io/ioutil"
    "os"
)


type FileAsyncOutput struct {
  Data []byte
  Err  error
}

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
  if err != nil { return []byte{}, errors.New("File is missing: " + path) }

  data, err := ioutil.ReadFile(path)
  if err != nil { return []byte{}, err }

  return data, nil
}


// Function to safely and abstractly load template files.
func LoadFileAsync(path string, out chan FileAsyncOutput) {

  // check if file exists
  _, err := os.Stat(path)
  if err != nil {
    out <- FileAsyncOutput{ []byte{}, errors.New("File is missing: " + path)}
    return
  }

  data, err := ioutil.ReadFile(path)
  if err != nil {
    out <- FileAsyncOutput{ []byte{}, err}
    return
  }

  out <- FileAsyncOutput{data, nil}
}
