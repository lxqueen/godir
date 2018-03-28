package main

import (
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


// Some constants for easier expression of statuses
type Status int
const (
  OK Status = iota
  ERR
  QUIT
)

// Recursive generate async.
func GenerateAsync(path string, status chan Status, opts GenOpts) {
  files,_ := ioutil.ReadDir(path)
  subdirs := 0

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
  // Global
  pageBuffer = SubTag(pageBuffer, opts.Conf.Tag_root_step, rootStep)
  pageBuffer = SubTag(pageBuffer, opts.Conf.Tag_breadcrumb, breadcrumb)
  pageBuffer = SubTag(pageBuffer, opts.Conf.Tag_root_dir, path)
  pageBuffer = SubTag(pageBuffer, opts.Conf.Tag_domain, opts.Conf.Domain)
  pageBuffer = SubTag(pageBuffer, opts.Conf.Tag_root_step, rootStep)

  // Make a child channel that holds an amount of ints equal to the amount of subdirectories.
  childChan := make(chan Status, subdirs)

  // Now actually delve into subdirs recursively
  for _, f := range files {
    if (StringInSlice(f.Name(), opts.Config.Excludes)) {
      tmp = opts.ItemTemplate
      if (f.IsDir()) {
        go GenerateAsync(path + "/" + f.Name(), childChan, opts)
        subdirs++ // Keep track of how many subdir processes we are spawning

        tmp = SubTag(tmp, opts.Conf.Tag_class, "icon dir")
        tmp = SubTag(tmp, opts.Conf.Tag_item_type, "icon file-dir")
        tmp = SubTag(tmp, opts.Conf.Tag_filesize, FileSizeCount(DirSize(path + "/" + f.Name()))) // This does not get the right dir size.

      } else { // else is file
        tmp = SubTag(tmp, opts.Conf.Tag_class, "icon file")
        tmp = SubTag(tmp, opts.Conf.Tag_item_type, "icon file-icon")
        tmp = SubTag(tmp, opts.Conf.Tag_filesize, FileSizeCount(f.Size()))
      }

      // Now are the classes that are set regardless of dir/file
      tmp = SubTag(tmp, opts.Conf.Tag_filename, f.Name())
      tmp = SubTag(tmp, opts.Conf.Tag_last_modified, f.ModTime().Format("2006-01-02 15:04:05"))
      tmp = SubTag(tmp, opts.Conf.Tag_file_href, path + "/" + f.Name())

      itemBuffer.WriteString(tmp)
    }
  }
  pageBuffer = SubTag(pageBuffer, opts.Conf.Tag_contents, itemBuffer.String()) // Actually sub into page.

  err := ioutil.WriteFile(path + "/" + *opts.Args.Filename, []byte(pageBuffer), 0644)
  if err != nil { panic(err) }


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
