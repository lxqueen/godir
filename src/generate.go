package main

import (
  "encoding/json"
  "io/ioutil"
  "bytes"
  "strings"
  "sync"
)

type GenOpts struct {
  Conf Config
  Args Arguments

  ThemeTemplate string
  ItemTemplate  string
}

type ObjData struct {
  Hash string
  Html string
}


// Recursive generate async.
func GenerateAsync(path string, console LogObject, wg *sync.WaitGroup, opts GenOpts) {
  defer wg.Done() // Terminate the goroutine in the waitgroup when we've finished.

  console.Log("Generating for ", path)

  // Get a list of files and directories in PATH
  files, err := ioutil.ReadDir(path)
  if (err != nil) { console.Fatal("Error reading contents of ", path, " : ", err) }

  // Load dir.gdx, and deserialize it into a filename:hash slice.
  // This way, if the name changes, it re generates, and if the contents change, it also regens.
  idx, err := LoadGdx(path)
  if (err != nil) {
    console.Error("JSON Unmarshal Error: ", err)
  }

  // This string holds this directory's copy of the page template.
  page := opts.ThemeTemplate

  // This holds a path (e.g. "../../") that leads to the root of the file directory.
  rootStep := GenRootStep(path)

  // Generate the breadcrumb.
  breadCrumb := GenBreadCrumb(path)


  var itemBuffer bytes.Buffer // Will hold all the items (to be inserted into page at the end.)

  // Add in the "../" item before we generate any real items.
  tmp := opts.ItemTemplate
  tmp = SubTag(tmp, opts.Conf.Tag_class, "icon dir")
  tmp = SubTag(tmp, opts.Conf.Tag_file_href, "../")
  tmp = SubTag(tmp, opts.Conf.Tag_item_type, "icon dir-icon")
  tmp = SubTag(tmp, opts.Conf.Tag_filename, "Parent Directory")
  tmp = SubTag(tmp, opts.Conf.Tag_last_modified, "-")
  tmp = SubTag(tmp, opts.Conf.Tag_filesize, "-")
  itemBuffer.WriteString(tmp)

  // iterate over every file & dir in the directory.
  for _, file := range files {
    if ( !(StringInSlice(file.Name(), opts.Conf.Excludes) ) ) { // If the current item isn't in excludes...
      tmp := opts.ItemTemplate
      if ( file.IsDir() ) { // if it's a directory...

        // Add one to the waitgroup, and start the goroutine for that subdir.
        wg.Add(1)
        console.Ilog("Spawning new goroutine for subdir ", path + "/" + file.Name())
        go GenerateAsync(path + "/" + file.Name(), console, wg, opts)

        // Sub in tags
        tmp = SubTag(tmp, opts.Conf.Tag_class, "icon dir")
        tmp = SubTag(tmp, opts.Conf.Tag_item_type, "icon dir-icon")
        tmp = SubTag(tmp, opts.Conf.Tag_filesize, FileSizeCount(DirSize(path + "/" + file.Name())))
        tmp = SubTag(tmp, opts.Conf.Tag_filename, file.Name())
        tmp = SubTag(tmp, opts.Conf.Tag_last_modified, file.ModTime().Format("2006-01-02 15:04:05"))
        tmp = SubTag(tmp, opts.Conf.Tag_file_href, "./" + file.Name())

        itemBuffer.WriteString(tmp) // write the composed item into the buffer.
        
      } else { // not a dir, must be file
        // First check to see if the file has changed
        fDat, err := LoadFile(path + "/" + file.Name())
        if ( err != nil ) { console.Error("Unable to open file ", path + "/" + file.Name()) }
        changed := RecordChanged(idx, file.Name(), fDat, console)

        if (changed || *opts.Args.Force) {
          tmp = SubTag(tmp, opts.Conf.Tag_class, "icon file")
          tmp = SubTag(tmp, opts.Conf.Tag_item_type, "icon file-icon")
          tmp = SubTag(tmp, opts.Conf.Tag_filesize, FileSizeCount(file.Size()))
          tmp = SubTag(tmp, opts.Conf.Tag_filename, file.Name())
          tmp = SubTag(tmp, opts.Conf.Tag_last_modified, file.ModTime().Format("2006-01-02 15:04:05"))
          tmp = SubTag(tmp, opts.Conf.Tag_file_href, "./" + file.Name())

          fileData, err := LoadFile( path + "/" + file.Name()) // open the file for hashing
          if (err != nil) { console.Error("Error opening file ", path, "/", file.Name(), " for hashing : ", err) }
          idx[file.Name()] = ObjData{ Hash: Hash(fileData), Html: tmp } // Re-set the appropriate fields, since we've changed something.
          itemBuffer.WriteString(tmp)
        } else {
          itemBuffer.WriteString(idx[file.Name()].Html) // If it hasn't changed, and we're not forcing, just use the existing html.
        }
      } // END if/else IsDir()
    } // END if ( !(StringInSlice(f.Name(), opts.Conf.Excludes) ) )
  } // END for _, file := range files


  // Substitute some global tags out of the main page, to get that work out of the way already.
  page = SubTag(page, opts.Conf.Tag_root_step, rootStep)
  page = SubTag(page, opts.Conf.Tag_breadcrumb, breadCrumb)
  page = SubTag(page, opts.Conf.Tag_root_dir, path)
  page = SubTag(page, opts.Conf.Tag_domain, opts.Conf.Domain)
  page = SubTag(page, opts.Conf.Tag_root_step, rootStep)


  // Now sub the contents into the page, as generated.
  page = SubTag(page, opts.Conf.Tag_contents, itemBuffer.String())
  // Now write the page to the actual file.
  err = ioutil.WriteFile(path + "/" + *opts.Args.Filename, []byte(page), 0644)
  if err != nil { console.Fatal("Unable to write page file ", *opts.Args.Filename, " : ", err) }

  // Also, write in the dir.gdx file, for skipDirs
  data, err := json.Marshal(idx)
  if (err != nil) { console.Fatal("Unable to write to ", path, "/dir.idx : ", err) }
  err = ioutil.WriteFile(path + "/dir.gdx", data, 0644)
} // END func GenerateAsync


// Generates root step from path.
func GenRootStep(path string) string {
  split := strings.Split(path, "/")
  if (len(split) <= 1) {
    return "./"
  } else {
    var step bytes.Buffer
    for i := 0; i < (len(split)-1); i++ {
  		step.WriteString("../")
  	}
    return step.String()
  }
}

func GenBreadCrumb(path string) string {
  return "BREADCRUMB WIP"
}

func RecordChanged(idx map[string]ObjData, fName string, fDat []byte, console LogObject) bool {
  record, exists := idx[fName]
  if ( exists ) {
    console.Ilog("File ", fName, " not in dir.gdx.")
    return true
  }
  if ( (Hash(fDat) == record.Hash) ) { // if it's unchanged...
    console.Ilog("File ", fName, " unchanged.")
    return false
  } else { // if it doesn't match, it's been changed.
    console.Ilog("File ", fName, " changed.")
    return true
  }
}

func LoadGdx(path string) (map[string]ObjData, error) {
  idx := map[string]ObjData{}
  idxRaw, err := LoadFile(path + "/dir.gdx")
  if (err != nil) {
    // This can fail silently since all it means is we don't have a dir.gdx yet.
    return idx, nil
  }
  err = json.Unmarshal(idxRaw, &idx) // unmarshal file contents into idx
  if (err != nil) {
    return idx, err
  }
  return idx, nil
}
