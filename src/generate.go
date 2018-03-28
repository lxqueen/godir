package main

import (
  "encoding/json"
  "io/ioutil"
  "bytes"
  "strings"
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

// Some constants for easier expression of statuses
type Status int
const (
  OK Status = iota
  ERR
  QUIT
)

// Recursive generate async.
func GenerateAsync(path string, console LogObject, status chan Status, opts GenOpts) {
  console.Log("Generating for ", path)
  files,_ := ioutil.ReadDir(path)
  subdirs := 0

  // Load dir.idx, and deserialize it into a filename:hash slice.
  // This way, if the name changes, it re generates, and if the contents change, it also regens.
  idx := map[string]ObjData{}
  idxRaw, err := LoadFile(path + "/dir.idx")
  if (err != nil) {
    // This can fail silently since all it means is we don't have a dir.idx yet.
    idxRaw = []byte{}
  } else {
    err := json.Unmarshal(idxRaw, &idx)
    if (err != nil) {
      console.Error(err)
    }
  }

  // This is the buffer that holds the current page.
  pageBuffer := (opts.ThemeTemplate)

  // This buffer holds all itemtemplates of this dir
  var itemBuffer bytes.Buffer

  rootStep := GenRootStep(path)

  // Gen Breadcrumbs from rootstep:
  breadcrumb := "BREADCRUMB WIP"

  // Add in the "../ item"
  tmp := opts.ItemTemplate
  tmp = SubTag(tmp, opts.Conf.Tag_class, "icon dir")
  tmp = SubTag(tmp, opts.Conf.Tag_file_href, "../")
  tmp = SubTag(tmp, opts.Conf.Tag_item_type, "icon dir-icon")
  tmp = SubTag(tmp, opts.Conf.Tag_filename, "Parent Directory")
  tmp = SubTag(tmp, opts.Conf.Tag_last_modified, "-")
  tmp = SubTag(tmp, opts.Conf.Tag_filesize, "-")
  itemBuffer.WriteString(tmp)

  // REMEMBER func SubTag(raw string, tag string, new string) string
  // Global tag subs
  pageBuffer = SubTag(pageBuffer, opts.Conf.Tag_root_step, rootStep)
  pageBuffer = SubTag(pageBuffer, opts.Conf.Tag_breadcrumb, breadcrumb)
  pageBuffer = SubTag(pageBuffer, opts.Conf.Tag_root_dir, path)
  pageBuffer = SubTag(pageBuffer, opts.Conf.Tag_domain, opts.Conf.Domain)
  pageBuffer = SubTag(pageBuffer, opts.Conf.Tag_root_step, rootStep)

  // Make a child channel that holds an amount of ints equal to the amount of subdirectories.
  childChan := make(chan Status, subdirs)

  // Now actually delve into subdirs recursively
  for _, f := range files {
    file, exists := idx[f.Name()]
    fd, err := LoadFile(path + "/" + f.Name())
    if (err != nil) {
      if (!f.IsDir()){
        console.Error(err)
      }
    }
    if ( (Hash(fd) == file.Hash) ) { // if it's unchanged...
      itemBuffer.WriteString(file.Html)
    } else { // if it doesn't match, it's been changed.
      // We change exists to false because we know it exists,
      // we just want to make sure the next if statement fires.
      exists = false
    }

    if (!exists || *opts.Args.Force) { // if it doesn't exist or we're forcing...
      if ( !(StringInSlice(f.Name(), opts.Conf.Excludes)) ) { // if it's not in the excludes...
        tmp = opts.ItemTemplate
        if (f.IsDir()) {
          go GenerateAsync(path + "/" + f.Name(), console, childChan, opts)
          subdirs++ // Keep track of how many subdir processes we are spawning

          tmp = SubTag(tmp, opts.Conf.Tag_class, "icon dir")
          tmp = SubTag(tmp, opts.Conf.Tag_item_type, "icon file-dir")
          tmp = SubTag(tmp, opts.Conf.Tag_filesize, FileSizeCount(DirSize(path + "/" + f.Name())))
        } else { // else is file
          tmp = SubTag(tmp, opts.Conf.Tag_class, "icon file")
          tmp = SubTag(tmp, opts.Conf.Tag_item_type, "icon file-icon")
          tmp = SubTag(tmp, opts.Conf.Tag_filesize, FileSizeCount(f.Size()))
        }

        // Now are the classes that are set regardless of dir/file
        tmp = SubTag(tmp, opts.Conf.Tag_filename, f.Name())
        tmp = SubTag(tmp, opts.Conf.Tag_last_modified, f.ModTime().Format("2006-01-02 15:04:05"))
        tmp = SubTag(tmp, opts.Conf.Tag_file_href, path + "/" + f.Name())

        fileData, err := LoadFile(path + "/" + f.Name())
        if (err != nil) { console.Error(err) }
        idx[f.Name()] = ObjData{ Hash: Hash(fileData), Html: tmp } // Re-set the appropriate fields, since we've changed something.
        itemBuffer.WriteString(tmp)
      }
    }

  }
  pageBuffer = SubTag(pageBuffer, opts.Conf.Tag_contents, itemBuffer.String()) // Actually sub into page.

  // Now write the page to the actual file.
  err = ioutil.WriteFile(path + "/" + *opts.Args.Filename, []byte(pageBuffer), 0644)
  if err != nil { panic(err) }

  // Also, write in the dir.idx file, for skipDirs
  data, err := json.Marshal(idx)
  if (err != nil) { console.Error(err) }
  err = ioutil.WriteFile(path + "/dir.idx", data, 0644)

  // Wait for all the sub-dirs to finish executing before returning yourself.
  for i := 0; i < subdirs; i++ {
		<- childChan
	}
  status <- QUIT
}


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
