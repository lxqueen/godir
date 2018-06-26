package main

import (
	"bytes"
	"errors"
	"github.com/cespare/xxhash"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"regexp"
	"fmt"
	"net/url"
)

type FileAsyncOutput struct {
	Data []byte
	Err  error
}

// https://stackoverflow.com/questions/19101419/go-golang-formatfloat-convert-float-number-to-string
func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 2, 64)
}

// Rounding function https://stackoverflow.com/questions/39544571/golang-round-to-nearest-0-05
func Round(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}

// Convert byte count to size string
func FileSizeCount(fileSize int64) string {
	if fileSize >= 1000000000000 {
		return FloatToString(Round(float64(fileSize/1000000000000), 2)) + " TB" // convert to terabytes
	} else if fileSize >= 1000000000 {
		return FloatToString(Round(float64(fileSize/1000000000), 2)) + " GB" // convert to gigabytes
	} else if fileSize >= 1000000 {
		return FloatToString(Round(float64(fileSize/1000000), 2)) + " MB" // convert to megabytes
	} else if fileSize >= 1000 {
		return FloatToString(Round(float64(fileSize/1000), 2)) + " KB" // convert to kb
	} else {
		return strconv.FormatInt(fileSize, 10) + ".00 B"
	}
}

// Get the size of a given file or folder. https://stackoverflow.com/questions/32482673/golang-how-to-get-directory-total-size
func DirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size
}

func FileSize(path string) (int64, error) {
	f, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return f.Size(), nil
}

func FileSizeAsync(path string, out chan int64) {

	f, err := os.Stat(path)
	if err != nil {
		out <- -1
		return
	}

	out <- f.Size()
}

func DirSizeAsync(path string, out chan int64) {

	f, err := os.Stat(path)
	if err != nil {
		out <- -1
		return
	}

	out <- f.Size()
}

func DirTreeCount(path string) int {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return -1
	}

	counter := 0

	// Now actually delve into subdirs recursively
	for _, f := range files {
		if f.IsDir() {
			counter += DirTreeCount(path + "/" + f.Name())
		}
		counter++
	}

	return counter
}

func DirTreeCountAsync(path string, out chan int) {
	files, _ := ioutil.ReadDir(path)
	count := 0
	subdirs := 0
	// First count the amount of subfiles.
	for _, f := range files {
		if !f.IsDir() && !InExcludes(f.Name()) {
			subdirs++
		} // only count those that aren't directories and aren't in excludes
	}

	// Make a child channel that holds an amount of ints equal to the amount of subdirectories.
	childChan := make(chan int, subdirs)

	// Now actually delve into subdirs recursively
	for _, f := range files {
		if f.IsDir() && !InExcludes(f.Name()) {
			go DirTreeCountAsync(path+"/"+f.Name(), childChan)
		}
		count++
	}

	// Now sum all elements in the channel to get the total subdirs count.
	for i := 0; i < subdirs; i++ {
		count += <-childChan
	}
	out <- count
}

func HashBytes(data []byte) string {
	return fmt.Sprintf("%X", xxhash.Sum64(data))
}

// Function to safely and abstractly load template files.
func LoadFile(path string) ([]byte, error) {

	// check if file exists
	_, err := os.Stat(path)
	if err != nil {
		return []byte{}, err
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

// Function to safely and abstractly load template files.
func LoadFileAsync(path string, out chan FileAsyncOutput) {

	// check if file exists
	_, err := os.Stat(path)
	if err != nil {
		out <- FileAsyncOutput{[]byte{}, errors.New("File is missing: " + path)}
		return
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		out <- FileAsyncOutput{[]byte{}, err}
		return
	}

	out <- FileAsyncOutput{data, nil}
}

func AppendFile(filename string, data []byte) error {
	if noWrite {
		return nil
	} else {
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		defer f.Close()

		_, err = f.WriteString(string(data))
		if err != nil {
			return err
		}
		return nil
	}
}

// Will clobber current contents.
func WriteFile(path string, data []byte, perm os.FileMode) error {
	if noWrite {
		return nil
	} else {
		return ioutil.WriteFile(path, data, perm)
	}
}

func GenRootStep(path string) string {
	split := strings.Split(path, "/")
	if len(split) <= 1 {
		return "."
	} else {
		var step bytes.Buffer
		step.WriteString(".")
		for i := 0; i < (len(split) - 1); i++ {
			step.WriteString("/..")
		}
		return step.String()
	}
}

func GenBreadCrumb(path string) string {
	pathS := strings.Split(path, "/")
	breadCrumb := ""
	crumbSep := `<a class="smaller breadcrumb" href="#"> / </a>`
	crumbItem := `<a class="breadcrumb" href="$addr$">$name$</a>`
	for index, crumb := range pathS {
		crumbTmp := ""
		crumbAddr := ""
		// First do crumbname
		if crumb == "." {
			crumbTmp = strings.Replace(crumbItem, "$name$", "", -1)
		} else {
			crumbTmp = strings.Replace(crumbItem, "$name$", strings.Trim(crumb, "./"), -1)
		}

		// Then crumb's link address
		if path == "." {
			crumbAddr = `#`
		} else {
			for i := index; i > 0; i-- {
				crumbAddr += (`../`)
			}
			crumbAddr += `./`
		}

		breadCrumb += strings.Replace(crumbTmp, "$addr$", crumbAddr, -1)

		// If is not last item in list, append >
		if crumb != pathS[len(pathS)-1] {
			breadCrumb += crumbSep
		}
	}
	return breadCrumb
}

// Modifies the sideNav global to be efficient
func GenSidenav(path string, streak int) { // Streak needs to start at 0
	console.Ilog("Running dirTree(", path, ", ", streak, ")")
	// Get a list of files and directories in PATH
	files, err := ioutil.ReadDir(path)
	if err != nil {
		console.Error("Error reading contents of ", path, " : ", err)
		return
	}

	// Loop over every file...
	for _, p := range files {
		// Only consider files that are not in the excludes list.
		if ( !InExcludes(p.Name())/* If the filename is not in the excludes list...*/ ) {
			// First, compare config information with any symlinks we may encounter
			// Check for symlinks
			fi, err := os.Lstat(path + "/" + p.Name())
			if err != nil { console.Error(err); continue; }
			if fi.Mode() & os.ModeSymlink == os.ModeSymlink {
				// if is a symlink
				realPath, err := os.Readlink(path + "/" + p.Name())
				if err != nil { console.Error(err); continue; }
				 // if the realpath is not contained within the webroot...
				 // AND if we're jailing
				 // This shouldn't run IF unjail is set to true
				if ( !(strings.HasPrefix(realPath, *opts.Args.Webroot) && !(*opts.Args.Unjail) ) ) {
					// Abort this file.
					continue
				}
			}

			// If we've gotten this far, then it must either be a regular item or a VALID symlink according to the conf
			// Now, let's check to see if it's a directory or not, since that will change what we want to do.
			if p.IsDir() {
				// If it's a directory, we first need to see if it's empty or not, to put the chevron on those that are filled.
				isEmpty := true
				files, err := ioutil.ReadDir(path + "/" + p.Name())
				if err != nil {
					console.Error("Error reading contents of ", path, " : ", err)
					continue
				}
				// If we got something other than zero files, Empty = false
				// Loop through all files in the directory. If at least one of them is a directory, break and set isEmpty to false
				for _, item := range files {
					if item.IsDir() {
						isEmpty = false
						break
					}
				}

				uid := ""
				choices := strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", "")
				for i:= 0; i<8; i++ { uid += choices[rand.Intn(len(choices))] }


				// Set up some values real quick
				fullpath := url.QueryEscape(path + "/" + p.Name())
				// if is not empty
				if !isEmpty {
					sideNav += `<li class="pure-menu-item has-children"><div class="side-checkbox"><input type="checkbox" onclick="dropdown(this)" id="collapse_` + uid + `"/><label class="list-collapsed-icon" for="collapse_` + uid + `" id="chevron_` + uid + `"></label></div><div class="side-content" id="a1"><a href="$root-step$/` + fullpath + `" class="pure-menu-link">` + p.Name() + `</a></div>` + `<ul class="pure-menu-list"></ul>` + `<ul class="pure-menu-list default-hidden" id="` + uid + `">`
				} else {
					sideNav += `<li class="pure-menu-item"><div class="side-checkbox"><input type="checkbox" id="collapse_` + uid + `"/><label class="list-collapsed-icon" for="collapse_` + uid + `" id="chevron_` + uid + `"></label></div><div class="side-content" id="a1"><a href="$root-step$/` + fullpath + `" class="pure-menu-link">` + p.Name() + `</a></div>` + `<ul class="pure-menu-list" id="` + uid + `">`
				}
        GenSidenav(path + "/" + p.Name(), streak+1) // Will not write anything if it's empty
        sideNav += "</ul>"
			} // END if p.IsDir()
		} // END if !(StringInSlice)
	} // END for _, p := range files
} // END GenSideNav()

// IsInExcludes returns true if the given text string matches an exclude.
func InExcludes(text string) bool {
	if opts.Conf.Use_regex { // Use Regex
		for _, rule := range opts.Conf.Excludes {
			match, _ := regexp.MatchString(rule, text)
			if match { return match }
		}
		return false
	} else { // Don't use regex.
		return StringInSlice(text, opts.Conf.Excludes)
	}
}
