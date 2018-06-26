package main

import (
  "encoding/json"
  "strings"
  "io/ioutil"
  "sync"
  "os"
  "net/url"
)

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
  
  // If we need to re-generate the directory...
  if NeedsRegen(path) || *opts.Args.Force {
    regen(path, wg, semaphore)
  } else { // if we don't need to regenerate it then simply delve into subfolders.
    files, err := ioutil.ReadDir(path)
    if err != nil { console.Error("Error reading contents of ", path, " when in main generation routine. Error: ", err.Error) }
    for _, file := range files {
      if ( !InExcludes(file.Name()) ) { // If the current item isn't in excludes...
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
        } // END if file.IsDir()
      } // END if !InExcludes(file.Name())
    } // END for _, file := range files
  
  }
  
} // END func GenerateAsync

// Re-generates the index file at the given path.
func regen(path string, wg *sync.WaitGroup, semaphore chan struct{}) {
  // We need to update the GDX table for the next run.
  UpdateGdx(path)

  // This holds a path (e.g. "../../") that leads to the root of the file directory.
  rootStep := GenRootStep(path)

  // Generate the breadcrumb.
  breadCrumb := GenBreadCrumb(path)

  page := opts.ThemeHeader
  // Substitute some global tags out of the main page, to get that work out of the way already.
  page = SubTag(page, opts.Conf.Tag_sidenav, sideNav)
  page = SubTag(page, opts.Conf.Tag_root_step, rootStep)
  page = SubTag(page, opts.Conf.Tag_breadcrumb, breadCrumb)
  if path == "." || path == "./" {
    page = SubTag(page, opts.Conf.Tag_root_dir, "")
  } else {
    page = SubTag(page, opts.Conf.Tag_root_dir, strings.TrimLeft(strings.TrimLeft(path, `./\`), `./\`))
  }
  page = SubTag(page, opts.Conf.Tag_domain, opts.Conf.Domain)
  page = SubTag(page, opts.Conf.Tag_title, opts.Conf.Title)
  page = SubTag(page, opts.Conf.Tag_root_step, rootStep)
  err := WriteFile(path + "/" + *opts.Args.Filename, []byte(page), 0644)
  if err != nil {
    console.Error("Unable to write page header to file ", *opts.Args.Filename, " : ", err)
    return
  }

  // Add in the "../" item before we generate any real items.
  tmp := opts.ItemTemplate
  tmp = SubTag(tmp, opts.Conf.Tag_file_href, "../" + *opts.Args.Filename)
  tmp = SubTag(tmp, opts.Conf.Tag_item_type, "dir")
  tmp = SubTag(tmp, opts.Conf.Tag_filename, "Parent Directory")
  tmp = SubTag(tmp, opts.Conf.Tag_last_modified, "-")
  tmp = SubTag(tmp, opts.Conf.Tag_filesize, "-")

  // Append the ../ item to the output file.
  err = AppendFile(path + "/" + *opts.Args.Filename, []byte(tmp))
  if err != nil {
    console.Error("Unable to append page item to file ", *opts.Args.Filename, " : ", err)
    return
  }

  // now iterate over each dir and spawn its goroutine
  // we do this in its own loop so they start right away
  files, err := ioutil.ReadDir(path)
  if err != nil { console.Error("Error reading contents of ", path, " when in main generation routine. Error: ", err.Error) }
  for _, file := range files {
    if ( !InExcludes(file.Name()) ) { // If the current item isn't in excludes...
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

        // Generate the entry for this single item (directory).
        tmp = opts.ItemTemplate
        tmp = SubTag(tmp, opts.Conf.Tag_item_type, "dir")
        tmp = SubTag(tmp, opts.Conf.Tag_filesize, FileSizeCount(DirSize(path + "/" + file.Name())))
        tmp = SubTag(tmp, opts.Conf.Tag_filename, file.Name())
        tmp = SubTag(tmp, opts.Conf.Tag_last_modified, file.ModTime().Format("2006-01-02 15:04:05"))
        tmp = SubTag(tmp, opts.Conf.Tag_file_href, url.QueryEscape("./" + file.Name() + "/" + *opts.Args.Filename))

        // Append the composed item to file.
        err = AppendFile(path + "/" + *opts.Args.Filename, []byte(tmp))
        if err != nil {
          console.Error("Unable to append page item to file ", *opts.Args.Filename, " : ", err)
          return
        }
      } // END if file.IsDir()
    } // END if !InExcludes(file.Name())
  } // END for _, file := range files

  // Now iterate over files only.
  for _, file := range files {
    if ( !InExcludes(file.Name()) && !file.IsDir() ) { // If the current item isn't in excludes, and it's not a directory
      tmp = opts.ItemTemplate

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

      tmp = SubTag(tmp, opts.Conf.Tag_item_type, "file")
      tmp = SubTag(tmp, opts.Conf.Tag_filesize, FileSizeCount(file.Size()))
      tmp = SubTag(tmp, opts.Conf.Tag_filename, file.Name())
      tmp = SubTag(tmp, opts.Conf.Tag_last_modified, file.ModTime().Format("2006-01-02 15:04:05"))
      tmp = SubTag(tmp, opts.Conf.Tag_file_href, url.QueryEscape("./" + file.Name()))

      // Append the composed item to file.
      err = AppendFile(path + "/" + *opts.Args.Filename, []byte(tmp))
      if err != nil {
        console.Error("Unable to append page item to file ", *opts.Args.Filename, " : ", err)
        return
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
    } // END exclude check && make sure it's a file.

    console.Ilog(MemUsage() + " Loc=PostGenFile:" + path + "/" + file.Name())
  } // END for _, file := range files


  page = opts.ThemeFooter
  // Substitute some global tags out of the main page, to get that work out of the way already.
  page = SubTag(page, opts.Conf.Tag_sidenav, sideNav)
  page = SubTag(page, opts.Conf.Tag_root_step, rootStep)
  page = SubTag(page, opts.Conf.Tag_breadcrumb, breadCrumb)
  if path == "." || path == "./" {
    page = SubTag(page, opts.Conf.Tag_root_dir, "")
  } else {
    
    page = SubTag(page, opts.Conf.Tag_root_dir, strings.TrimLeft(strings.TrimLeft(path, `./\`), `./\`))
  }
  page = SubTag(page, opts.Conf.Tag_domain, opts.Conf.Domain)
  page = SubTag(page, opts.Conf.Tag_title, opts.Conf.Title)
  page = SubTag(page, opts.Conf.Tag_root_step, rootStep)

  // Now write the page footer to the actual file.
  err = AppendFile(path + "/" + *opts.Args.Filename, []byte(page))
  if err != nil {
    console.Error("Unable to write page file ", *opts.Args.Filename, " : ", err)
    return
  }

  console.Ilog(MemUsage() + " Loc=PostGenDir:" + path)
}