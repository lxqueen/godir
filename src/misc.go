package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/kalafut/imohash"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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

// Add excludes to this eventually
func DirTreeCountAsync(path string, excludes []string, out chan int) {
	files, _ := ioutil.ReadDir(path)
	count := 0
	subdirs := 0
	// First count the amount of subdirectories.
	for _, f := range files {
		if !f.IsDir() && !StringInSlice(f.Name(), excludes) {
			subdirs++
		} // only count those that aren't directories and aren't in excludes
	}

	// Make a child channel that holds an amount of ints equal to the amount of subdirectories.
	childChan := make(chan int, subdirs)

	// Now actually delve into subdirs recursively
	for _, f := range files {
		if !f.IsDir() && !StringInSlice(f.Name(), excludes) {
			go DirTreeCountAsync(path+"/"+f.Name(), excludes, childChan)
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
	sum := imohash.Sum(data)
	return hex.EncodeToString(sum[:16])
}

func HashFile(path string) string {
	sum, err := imohash.SumFile(path)
	if err != nil {
		fmt.Println("[ERR] ", err)
	}
	return hex.EncodeToString(sum[:16])
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
			for i := 0; i < index; i++ {
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
func GenSidenav(path string, indent int, streak int) { // Indent and streak need to start at 0
	console.Ilog("Running dirTree(", path, ", ", indent, ", ", streak, ")")
	// Get a list of files and directories in PATH
	files, err := ioutil.ReadDir(path)
	if err != nil {
		console.Error("Error reading contents of ", path, " : ", err)
		return
	}

	for _, p := range files {
		if !(StringInSlice(p.Name(), opts.Conf.Excludes)) {
			// Remove symlinks if they leave the webroot
			fi, err := os.Lstat(path + "/" + p.Name())
			if err != nil {
				console.Error(err)
			}
			if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
				// if is a symlink
				realPath, err := os.Readlink(path + "/" + p.Name())
				if err != nil {
					console.Error(err)
				}
				// if the realpath is not contained within the webroot...
				// AND if we're jailing
				// This shouldn't run IF unjail is set to true
				if !(strings.HasPrefix(realPath, *opts.Args.Webroot) && !(*opts.Args.Unjail)) {
					// Abort this file.
					continue
				}
			}

			// If we've made it this far, it must be a legal folder or symlink.
			fullpath := path + "/" + p.Name()

			// Check to see if the folder is empty and hide the chevron if it is (replace it with the folder icon)
			isEmpty := true
			if p.IsDir() {
				files, err := ioutil.ReadDir(fullpath)
				if err != nil {
					console.Error("Error reading contents of ", fullpath, " : ", err)
					continue
				}
				for _, subd := range files {
					if subd.IsDir() {
						isEmpty = false
					}
				}

				// Chevron UID generation
				rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
				choices := `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz`
				uid := ""
				for i := 0; i < 8; i++ {
					uid += string(choices[rand.Intn(len(choices))])
				}
				console.Ilog("CSS UID for item " + p.Name() + " is " + uid)
				// Handle some css magic for the dropdowns.
				// Not the nicest thing but it works.
				// I need to dynamically write styles here to make use of the button hack
				if !isEmpty {
					sideNav += (`<li class="pure-menu-item" style="padding-left: ` + strconv.Itoa(indent*10) + `px"><div class="side-checkbox"><input type="checkbox" onclick="dropdown(this)" id="collapse_` + uid + `"/><label class="list-collapsed-icon" for="collapse_` + uid + `" id="chevron_` + uid + `"></label></div><div class="side-content" id="a1"><a href="$root-step$/` + string(fullpath) + `" class="pure-menu-link">` + string(p.Name()) + `</a></div>`)
					sideNav += (`<ul class="pure-menu-list"></ul>`) //This is here to fix an issue regarding display:none, where it would randomly indent following elements.
					//It guarantees there's one element underneath the li(s), and secures the indentation.
					sideNav += (`<ul class="pure-menu-list default-hidden" id="` + uid + `">`) // This is the "real" <ul> to hold the subcontent, if at all.
				} else {
					sideNav += (`<li class="pure-menu-item" style="padding-left: ` + strconv.Itoa(indent*10) + `px"><div class="side-checkbox"><input type="checkbox" id="collapse_` + uid + `"/><label class="list-collapsed-icon" for="collapse_` + uid + `" id="chevron_` + uid + `"></label></div><div class="side-content" id="a1"><a href="$root-step$/` + string(fullpath) + `" class="pure-menu-link">` + string(p.Name()) + `</a></div>`)
					sideNav += (`<ul class="pure-menu-list" id="` + uid + `">`)
				}
				GenSidenav(fullpath, indent+1, streak+1)
				sideNav += ("</ul>")

				if isEmpty { // If there are no subfolders, switch the chevron for the folder icon
					sideNav += (`<style>#chevron_` + uid + `{background-image:url(/include/images/fallback/folder.png);background-size:70%;background-position:right center}</style>`)
					console.Ilog("Working Dir is Empty. Removing Chevron")
				}
				sideNav += (`</li>`)
			}
		}
	}
}
