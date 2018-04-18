package main

import (
  "encoding/json"
  "io/ioutil"
  "strings"
  "sync"
  "strconv"
  "os"
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
  console.Ilog(MemUsage() + " " + "Loc=DirPreload:" + path)

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

  console.Ilog(MemUsage() + "Bucket:" + strconv.Itoa(len(gdx.Bucket)) + " " + "Loc=IDXLoaded:" + path)

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
  err = WriteFile(path + "/" + *opts.Args.Filename, []byte(page), 0644)
  if err != nil {
    console.Error("Unable to write page header to file ", *opts.Args.Filename, " : ", err)
    return
  }

  // Add in the "../" item before we generate any real items.
  tmp := opts.ItemTemplate
  tmp = SubTag(tmp, opts.Conf.Tag_class, "icon dir")
  tmp = SubTag(tmp, opts.Conf.Tag_file_href, "../" + *opts.Args.Filename)
  tmp = SubTag(tmp, opts.Conf.Tag_item_type, "icon dir-icon")
  tmp = SubTag(tmp, opts.Conf.Tag_filename, "Parent Directory")
  tmp = SubTag(tmp, opts.Conf.Tag_last_modified, "-")
  tmp = SubTag(tmp, opts.Conf.Tag_filesize, "-")

  err = AppendFile(path + "/" + *opts.Args.Filename, []byte(tmp))
  if err != nil {
    console.Error("Unable to append page item to file ", *opts.Args.Filename, " : ", err)
    return
  }

  // now iterate over each dir and spawn it's goroutine
  // we do this in it's own loop so they start right away
  for _, file := range files {
    if ( !(StringInSlice(file.Name(), opts.Conf.Excludes) ) ) { // If the current item isn't in excludes...
      if ( file.IsDir() ) { // if it's a directory...

        // Check for symlinks
        fi, err := os.Lstat(path + "/" + file.Name())
        if err != nil { console.Error(err); }
      	if fi.Mode() & os.ModeSymlink == os.ModeSymlink {
          // if is a symlink
          realPath, err := os.Readlink(path + "/" + file.Name())
          if err != nil { console.Error(err) }
           // if the realpath is not contained within the webroot...
           // AND if we're jailing
           // This shouldn't run IF unjail is set to true
          if ( !(strings.HasPrefix(realPath, *opts.Args.Webroot) && !(*opts.Args.Unjail) ) ) {
            // Abort this file.
            continue
          }
      	}
        // If we've made it this far, it must be a legal folder or symlink.

        wg.Add(1)
        console.Ilog("Spawning new goroutine for subdir ", path + "/" + file.Name())
        go GenerateAsync(path + "/" + file.Name(), wg, semaphore)

        // Sub in tags
        tmp = opts.ItemTemplate
        tmp = SubTag(tmp, opts.Conf.Tag_class, "icon dir")
        tmp = SubTag(tmp, opts.Conf.Tag_item_type, "icon dir-icon")
        tmp = SubTag(tmp, opts.Conf.Tag_filesize, FileSizeCount(DirSize(path + "/" + file.Name())))
        tmp = SubTag(tmp, opts.Conf.Tag_filename, file.Name())
        tmp = SubTag(tmp, opts.Conf.Tag_last_modified, file.ModTime().Format("2006-01-02 15:04:05"))
        tmp = SubTag(tmp, opts.Conf.Tag_file_href, "./" + file.Name() + "/" + *opts.Args.Filename)

        // Append the composed item to file.
        err = AppendFile(path + "/" + *opts.Args.Filename, []byte(tmp))
        if err != nil {
          console.Error("Unable to append page item to file ", *opts.Args.Filename, " : ", err)
          return
        }
      }
    }
  }

  // iterate over every file & dir in the directory.
  for _, file := range files {
    if ( !(StringInSlice(file.Name(), opts.Conf.Excludes) ) ) { // If the current item isn't in excludes...
      tmp = opts.ItemTemplate
      if (!file.IsDir()) { // not a dir, must be file
        // Check for symlinks
        fi, err := os.Lstat(path + "/" + file.Name())
        if err != nil { console.Error(err); return; }
        if fi.Mode() & os.ModeSymlink == os.ModeSymlink {
          // if is a symlink
          realPath, err := os.Readlink(path + "/" + file.Name())
          if err != nil { console.Error(err) }
           // if the realpath is not contained within the webroot...
           // AND if we're jailing
           // This shouldn't run IF unjail is set to true
          if ( !(strings.HasPrefix(realPath, *opts.Args.Webroot) && !(*opts.Args.Unjail) ) ) {
            // Abort this file.
            continue
          }
        }

        regen := true
        if !(*opts.Args.Force) {
          fHash := HashFile(path + "/" + file.Name())

          // If the name is already in the DB
          if (gdx.ExistsName(file.Name())){
            entry, err := gdx.GetAllName(file.Name())
            if err != nil {
              console.Error("An error occured while querying the GDX table: ", err)
              // Continue with regen = true
            }
            // If the retrieved entry's hash does not match the current hash...
            if entry[0].Hash == fHash { regen = false }
          }
        }

        if (regen) {
          console.Log("Regenerating " + path + "/" + file.Name())
          fHash := HashFile(path + "/" + file.Name())

          tmp = SubTag(tmp, opts.Conf.Tag_class, "icon file")
          tmp = SubTag(tmp, opts.Conf.Tag_item_type, "icon file-icon")
          tmp = SubTag(tmp, opts.Conf.Tag_filesize, FileSizeCount(file.Size()))
          tmp = SubTag(tmp, opts.Conf.Tag_filename, file.Name())
          tmp = SubTag(tmp, opts.Conf.Tag_last_modified, file.ModTime().Format("2006-01-02 15:04:05"))
          tmp = SubTag(tmp, opts.Conf.Tag_file_href, "./" + file.Name())

          gdx.Insert( ObjData{ Name: file.Name(), Hash: fHash, Html: tmp  } ) // Re-set the appropriate fields, since we've changed something.
          // Append the composed item to file.
          err = AppendFile(path + "/" + *opts.Args.Filename, []byte(tmp))
          if err != nil {
            console.Error("Unable to append page item to file ", *opts.Args.Filename, " : ", err)
            return
          }
        } else {// If it hasn't changed, and we're not forcing, just use the existing html.
          // Append the composed item to file.
          objl, err := gdx.GetAllName(file.Name())
          if err != nil {
            console.Error("Unable to get page item from GDX : ", err)
            return
          }
          err = AppendFile(path + "/" + *opts.Args.Filename, []byte(objl[0].Html))
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

    console.Ilog(MemUsage() + "Bucket:" + strconv.Itoa(len(gdx.Bucket)) + " " + "Loc=PostGenFile:" + path + "/" + file.Name())
  } // END for _, file := range files


  page = opts.ThemeFooter
  // Substitute some global tags out of the main page, to get that work out of the way already.
  page = SubTag(page, opts.Conf.Tag_root_step, rootStep)
  page = SubTag(page, opts.Conf.Tag_breadcrumb, breadCrumb)
  page = SubTag(page, opts.Conf.Tag_root_dir, path)
  page = SubTag(page, opts.Conf.Tag_domain, opts.Conf.Domain)
  page = SubTag(page, opts.Conf.Tag_root_step, rootStep)

  // Now write the page footer to the actual file.
  err = AppendFile(path + "/" + *opts.Args.Filename, []byte(page))
  if err != nil {
    console.Error("Unable to write page file ", *opts.Args.Filename, " : ", err)
    return
  }

  console.Ilog(MemUsage() + "Bucket:" + strconv.Itoa(len(gdx.Bucket)) + " " + "Loc=PostGenDir:" + path)
} // END func GenerateAsync
