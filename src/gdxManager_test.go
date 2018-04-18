package main

import (
  "testing"
  "encoding/json"
  "fmt"
)

func TestNewGdxTable(t *testing.T) {
  fmt.Println("Testing table creation")
  db, err := NewGdxTable("./dir.gdx")
  if err != nil { t.Error( "GDX Creation failed. Reason: " + err.Error() ) }
  db.Close()
}

func TestGetNonExistentValue(t *testing.T) {
  fmt.Println("Testing GetNonExistentValue")
  db, err := NewGdxTable("./dir.gdx")
  if err != nil { t.Error( "GDX Creation failed. Reason: " + err.Error() ) }

  objs, err := db.GetAll("g90734yogiwrygo3y4o")
  if len(objs) != 0 { t.Error( "Impossible value did not fetch zero rows: " + err.Error() )}

  db.Close()
}

func TestGdxFunctionality(t *testing.T) {
  fmt.Println("Testing Insert/UPDATE")
  fmt.Println("-> Creating table")
  db, err := NewGdxTable("./test")
  if err != nil { t.Error( "GDX Creation failed. Reason: " + err.Error() ) }

  fmt.Println("-> Attempting to get nonexistent value")
  objs, err := db.GetAll("201912187")
  if (db.ExistsName("201912187")) { t.Error( "Nonexistent value exists: " + err.Error() )}

  dat := ObjData{Name:"201912187", Hash:"qwiuy4gfi4y3y", Html:`I am definitely HT&amp;ML\\`}

  fmt.Println("-> Inserting value")
  err = db.Insert(dat)
  if err != nil { t.Error("There was a problem when inserting objdata: " + err.Error() ) }

  fmt.Println("-> Checking for newly inserted value")
  objs, err = db.GetAllName(dat.Hash)
  if !(db.ExistsName(dat.Name)) { t.Error( "Existing value does not exist: " + err.Error() )}

  datJ, err := json.Marshal(dat)
  objJ, err := json.Marshal(objs[0])
  if(err != nil) { t.Error("Error marshalling JSON " + err.Error() ) }

  if(string(datJ) != string(objJ)) { t.Error("Expected " + string(datJ) + " and got " + string(objJ)) }

  db.Close()
}
