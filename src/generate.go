package main

import (
  "encoding/json"
  "io/ioutil"
  "bytes"
  "strings"
  "sync"
)

type ObjData struct {
  Name string
  Hash string
  Html string
}


// Recursive generate async.
func GenerateAsync(path string, wg *sync.WaitGroup, semaphore chan struct{}) {

  // Large buffer to simulate heavy memory use
  //buffer := make([]byte, 4*4096*4096)
  // defer buffer[0] = 0x00 // Don't free buffer until the very end of the goroutine

  defer wg.Done() // Terminate the goroutine in the waitgroup when we've finished.

  semaphore <- struct{}{}  // lock
  defer func() {
    <-semaphore //unlock
  }()


  console.Log("Generating for ", path)
  console.Ilog(MemUsage() + "Loc=DirPreload:" + path)

  // Get a list of files and directories in PATH
  files, err := ioutil.ReadDir(path)
  if (err != nil) {
    console.Error("Error reading contents of ", path, " : ", err)
    return
  }

  // Load dir.gdx into a database wrapper object
  gdx, err := NewGdxTable(path)
  if (err != nil) {
    console.Error("Error getting GDX table: ", err)
  }
  defer gdx.Close()

  console.Ilog(MemUsage() + "Loc=IDXLoaded:" + path)

  // This holds a path (e.g. "../../") that leads to the root of the file directory.
  rootStep := GenRootStep(path)

  // Generate the breadcrumb.
  breadCrumb := GenBreadCrumb(path)

  page := opts.ThemeHeader
  // Substitute some global tags out of the main page, to get that work out of the way already.
  page = SubTag(page, opts.Conf.Tag_root_step, rootStep)
  page = SubTag(page, opts.Conf.Tag_breadcrumb, breadCrumb)
  page = SubTag(page, opts.Conf.Tag_root_dir, path)
  page = SubTag(page, opts.Conf.Tag_domain, opts.Conf.Domain)
  page = SubTag(page, opts.Conf.Tag_root_step, rootStep)
  err = WriteFile(path + "/" + *opts.Args.Filename, []byte(opts.ThemeHeader), 0644)
  if err != nil {
    console.Error("Unable to write page header to file ", *opts.Args.Filename, " : ", err)
    return
  }

  // Add in the "../" item before we generate any real items.
  tmp := opts.ItemTemplate
  tmp = SubTag(tmp, opts.Conf.Tag_class, "icon dir")
  tmp = SubTag(tmp, opts.Conf.Tag_file_href, "../")
  tmp = SubTag(tmp, opts.Conf.Tag_item_type, "icon dir-icon")
  tmp = SubTag(tmp, opts.Conf.Tag_filename, "Parent Directory")
  tmp = SubTag(tmp, opts.Conf.Tag_last_modified, "-")
  tmp = SubTag(tmp, opts.Conf.Tag_filesize, "-")

  err = AppendFile(path + "/" + *opts.Args.Filename, []byte(tmp))
  if err != nil {
    console.Error("Unable to append page item to file ", *opts.Args.Filename, " : ", err)
    return
  }

  // iterate over every file & dir in the directory.
  for _, file := range files {
    if ( !(StringInSlice(file.Name(), opts.Conf.Excludes) ) ) { // If the current item isn't in excludes...
      tmp = opts.ItemTemplate
      if ( file.IsDir() ) { // if it's a directory...
        // Add one to the waitgroup, and start the goroutine for that subdir.
        wg.Add(1)
        console.Ilog("Spawning new goroutine for subdir ", path + "/" + file.Name())
        go GenerateAsync(path + "/" + file.Name(), wg, semaphore)

        // Sub in tags
        tmp = SubTag(tmp, opts.Conf.Tag_class, "icon dir")
        tmp = SubTag(tmp, opts.Conf.Tag_item_type, "icon dir-icon")
        tmp = SubTag(tmp, opts.Conf.Tag_filesize, FileSizeCount(DirSize(path + "/" + file.Name())))
        tmp = SubTag(tmp, opts.Conf.Tag_filename, file.Name())
        tmp = SubTag(tmp, opts.Conf.Tag_last_modified, file.ModTime().Format("2006-01-02 15:04:05"))
        tmp = SubTag(tmp, opts.Conf.Tag_file_href, "./" + file.Name())

        // Append the composed item to file.
        err = AppendFile(path + "/" + *opts.Args.Filename, []byte(tmp))
        if err != nil {
          console.Error("Unable to append page item to file ", *opts.Args.Filename, " : ", err)
          return
        }

      } else { // not a dir, must be file
        fHash := HashFile(path + "/" + file.Name())

        regen := true;
        // If the name is already in the DB
        if (gdx.ExistsName(file.Name())){
          entry, err := gdx.GetAllName(file.Name())
          if err != nil {
            console.Error("An error occured while querying the GDX table")
            // Continue with regen = true
          }
          // If the retrieved entry's hash does not match the current hash...
          if entry[0].Hash == fHash { regen = false }
        }
        if (regen || *opts.Args.Force) {
          tmp = SubTag(tmp, opts.Conf.Tag_class, "icon file")
          tmp = SubTag(tmp, opts.Conf.Tag_item_type, "icon file-icon")
          tmp = SubTag(tmp, opts.Conf.Tag_filesize, FileSizeCount(file.Size()))
          tmp = SubTag(tmp, opts.Conf.Tag_filename, file.Name())
          tmp = SubTag(tmp, opts.Conf.Tag_last_modified, file.ModTime().Format("2006-01-02 15:04:05"))
          tmp = SubTag(tmp, opts.Conf.Tag_file_href, "./" + file.Name())

          gdx.Insert( ObjData{ Name: file.Name(), Hash: fHash, Html: tmp  }) // Re-set the appropriate fields, since we've changed something.
          // Append the composed item to file.
          err = AppendFile(path + "/" + *opts.Args.Filename, []byte(tmp))
          if err != nil {
            console.Error("Unable to append page item to file ", *opts.Args.Filename, " : ", err)
            return
          }
        } else {// If it hasn't changed, and we're not forcing, just use the existing html.
          // Append the composed item to file.
          //err = AppendFile(path + "/" + *opts.Args.Filename, []byte(idx[file.Name()].Html))
          if err != nil {
            console.Error("Unable to append page item to file ", *opts.Args.Filename, " : ", err)
            return
          }
        }

        // Add in record for file searching.
        fileRec := make(map[string]string)
        fileRec["size"] = FileSizeCount(file.Size())
        fileRec["path"] = path + "/" + file.Name()
        fileRec["lastmodified"] = file.ModTime().Format("2006-01-02 15:04:05")
        fileRec["name"] = file.Name()

        console.Ilog("Marshalling ", path + "/" + file.Name(), " to include/files.json")
        jdata, err := json.Marshal(fileRec)
        if (err != nil) {
          console.Error(err)
          continue
        }

        // We are naively appending the JSON string to the file WITHOUT opening it to save memory.
        err = AppendFile("./include/files.json", append([]byte(", "), jdata...) )
        if (err != nil) {
          console.Error(err)
          continue
        }
      } // END if/else IsDir()
    } // END if ( !(StringInSlice(f.Name(), opts.Conf.Excludes) ) )

    console.Ilog(MemUsage() + "Loc=PostGenFile:" + path + "/" + file.Name())
  } // END for _, file := range files


  page = opts.ThemeFooter
  // Substitute some global tags out of the main page, to get that work out of the way already.
  page = SubTag(page, opts.Conf.Tag_root_step, rootStep)
  page = SubTag(page, opts.Conf.Tag_breadcrumb, breadCrumb)
  page = SubTag(page, opts.Conf.Tag_root_dir, path)
  page = SubTag(page, opts.Conf.Tag_domain, opts.Conf.Domain)
  page = SubTag(page, opts.Conf.Tag_root_step, rootStep)

  // Now write the page footer to the actual file.
  err = WriteFile(path + "/" + *opts.Args.Filename, []byte(page), 0644)
  if err != nil {
    console.Error("Unable to write page file ", *opts.Args.Filename, " : ", err)
    return
  }
  // Append the footer item to file.
  err = AppendFile(path + "/" + *opts.Args.Filename, []byte(opts.ThemeFooter))
  if err != nil {
    console.Error("Unable to append page item to file ", *opts.Args.Filename, " : ", err)
    return
  }

  console.Ilog(MemUsage() + "Loc=PostGenDir:" + path)
} // END func GenerateAsync


// Generates root step from path.
func GenRootStep(path string) string {
  split := strings.Split(path, "/")
  if (len(split) <= 1) {
    return "."
  } else {
    var step bytes.Buffer
    step.WriteString(".")
    for i := 0; i < (len(split)-1); i++ {
  		step.WriteString("/..")
  	}
    return step.String()
  }
}

func GenBreadCrumb(path string) string {
  pathSlice := strings.Split(path, "/")
  var breadCrumb bytes.Buffer
  //crumbSep := "<a class='smaller' href='#'> > </a>"
  crumbItem := "<a class='breadcrumb' href='$crumbAddr$'> $name$ </a>"

  for _, crumb := range pathSlice {
    // First do crumbname
    if (crumb == ".") {
      breadCrumb.WriteString(strings.Replace(crumbItem, "$name$", "", -1))
    } else {
      breadCrumb.WriteString(strings.Replace(crumbItem, "$name$", "", -1))
    }
    // now crumb's link address
  }
  return "BREADCRUMB WIP"
}
