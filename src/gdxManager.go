package main

import (
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
)

// Some information'
/*
 A GDX table is *always* going to be named `gdx`. It will always have two

*/
// You must defer GdxTable.Close() in order to close the DB.
type GdxTable struct {
  Path  string
  DB    *sql.DB
}

// For Convenience, and to allow for expansions later on
func NewGdxTable(path string) (GdxTable, error) {
  db, err := sql.Open("sqlite3", path + "/gdx.db")
  if err != nil { return GdxTable{}, err }

  // Set up table if not exists
  _, err = db.Exec(`CREATE TABLE IF NOT EXISTS GDX(
    HASH TEXT PRIMARY KEY NOT NULL,
    NAME TEXT NOT NULL,
    HTML TEXT NOT NULL
    );`)
  if err != nil { return GdxTable{}, err }

  return GdxTable{Path:path, DB:db}, nil
}

// Returns a slice of ObjData elements, one for each result
// If the db has been configured right, should only return one object per.
func (r GdxTable) GetAll(hash string) ([]ObjData, error) {
  rows, err := r.DB.Query("SELECT * FROM GDX WHERE HASH=?", hash )
  if err != nil { return []ObjData{}, err }
  defer rows.Close()

  var retVal []ObjData
  for rows.Next() {
    var hash string
    var name string
    var html string

    err = rows.Scan(&hash, &name, &html)
    if err != nil { return []ObjData{}, err }

    retVal = append(retVal, ObjData{ Name:name, Hash:hash, Html:html})
  }

  return retVal, nil
}

func (r GdxTable) GetAllName(name string) ([]ObjData, error) {
  rows, err := r.DB.Query("SELECT * FROM GDX WHERE NAME=?", name )
  if err != nil { return []ObjData{}, err }
  defer rows.Close()

  var retVal []ObjData
  for rows.Next() {
    var hash string
    var name string
    var html string

    err = rows.Scan(&hash, &name, &html)
    if err != nil { return []ObjData{}, err }

    retVal = append(retVal, ObjData{ Name:name, Hash:hash, Html:html})
  }

  return retVal, nil
}

func (r GdxTable) ExistsHash(hash string) bool {
  objs, err := r.GetAll(hash)
  if err != nil { console.Error( err ); return false; }
  if len(objs) == 0 { return false }
  return true
}

func (r GdxTable) ExistsName(name string) bool {
  objs, err := r.GetAllName(name)
  if err != nil { console.Error( err ); return false; }
  if len(objs) == 0 { return false }
  return true
}

// Will either create a new entry or update an existing one.
func (r GdxTable) Insert(dat ObjData) error {
  rows, err := r.DB.Query("SELECT * FROM GDX WHERE NAME=?", dat.Name )
  if err != nil { return err }
  defer rows.Close()

  // If we got nothing, we need to make a new entry.
  if( !(rows.Next()) ) {
    _, err := r.DB.Exec("INSERT INTO GDX VALUES (?,?,?);", dat.Hash, dat.Name, dat.Html)
    if err != nil { return err }
    return nil// If we had nothing to modify, we insert it and that's that.
  }

  // If we did indeed get something, we need to replace such entries.
  _, err = r.DB.Exec("UPDATE GDX SET HASH = ?, NAME = ?, HTML = ? WHERE NAME = ?", dat.Hash, dat.Name, dat.Html, dat.Name)
  if err != nil { return err }
  return nil
}

// Closes the database. This must be called at the end of use or deferred.
func (r GdxTable) Close() {
  r.DB.Close()
}
