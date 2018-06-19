package main

import (
  "io/ioutil"
  "bytes"
)

// Some information'
/*
 A GDX table is *always* going to be named `.gdx`.
 It contains one thing: A hash (xxhash) of all the names of the contents of the directory, concatenated.
 If the in-file hash does not match the computed result, then it signals that the contents have changed.
*/

// Finds the gdx located in path and returns it's contents.
func GetGdx(path string) (string, error) {
  bytes, err := LoadFile(path + "/.gdx")
  if err != nil {
    return "An error occured getting the GDX table. Check the console for more details.", err
  }
  return string(bytes), nil
}

// Compares a computed GDX hash with a saved one. Returns false if they match, and true if they don't match or one is not found.
func HasChanged(path string) bool {
  // First, check to see if it's even there.
  // If it is, load it.
  savedGdx, err := GetGdx(path)
  if err != nil { return false } // if there's an error getting the GDX table then just regenerate the whole directory.

  files, err := ioutil.ReadDir(path)
  if err != nil { console.Error("Error reading contents of ", path, " when checking GDX hash. Error: ", err.Error) }

  var fileSum bytes.Buffer
  // Now concat all the files together.
  for _, file := range files {
    fileSum.WriteString(file.Name())
  }

  // Return true if they are different. Return false if they are the same.
  return !( savedGdx == HashBytes(fileSum.Bytes()) )
}

func UpdateGdx(path string) {
  files, err := ioutil.ReadDir(path)
  if err != nil { console.Error("Error reading contents of ", path, " when checking GDX hash. Error: ", err.Error) }

  var fileSum bytes.Buffer
  // Now concat all the files together.
  for _, file := range files {
    fileSum.WriteString(file.Name())
  }

  // Write the hash to the gdx file.
  err = WriteFile(path + "/.gdx", []byte(HashBytes( fileSum.Bytes()) ), 0644)
  if err != nil {
    console.Error(err)
  }
}