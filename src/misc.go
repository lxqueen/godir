package main

import (
  "os"
  "math"
  "strconv"
  "io/ioutil"
  "path/filepath"
  "github.com/kalafut/imohash"
  "errors"
  "fmt"
  "encoding/hex"
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
    if (fileSize >= 1000000000000) {
      return FloatToString(Round(float64(fileSize / 1000000000000), 2)) + " TB" // convert to terabytes
    } else if (fileSize >= 1000000000) {
      return FloatToString(Round(float64(fileSize / 1000000000), 2)) + " GB" // convert to gigabytes
    } else if (fileSize >= 1000000) {
      return FloatToString(Round(float64(fileSize / 1000000), 2)) + " MB" // convert to megabytes
    } else if (fileSize >= 1000) {
       return FloatToString(Round(float64(fileSize / 1000), 2)) + " KB" // convert to kb
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
        return  err
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
  if (err != nil) { return -1 }

  counter := 0

  // Now actually delve into subdirs recursively
  for _, f := range files {
      if (f.IsDir()) {
        counter += DirTreeCount(path + "/" + f.Name())
      }
      counter++
  }

  return counter
}

// Add excludes to this eventually
func DirTreeCountAsync(path string, excludes []string, out chan int) {
  files,_ := ioutil.ReadDir(path)
  count := 0
  subdirs := 0
  // First count the amount of subdirectories.
  for _, f := range files {
    if (!f.IsDir() && !StringInSlice(f.Name(), excludes)) { subdirs++ } // only count those that aren't directories and aren't in excludes
  }

  // Make a child channel that holds an amount of ints equal to the amount of subdirectories.
  childChan := make(chan int, subdirs)

  // Now actually delve into subdirs recursively
  for _, f := range files {
      if (!f.IsDir() && !StringInSlice(f.Name(), excludes)) {
        go DirTreeCountAsync(path + "/" + f.Name(), excludes, childChan)
      }
      count++
  }

  // Now sum all elements in the channel to get the total subdirs count.
  for i := 0; i < subdirs; i++ {
		count += <- childChan
	}
  out <- count
}

func HashBytes(data []byte) string {
  sum := imohash.Sum(data)
  return hex.EncodeToString(sum[:16])
}

func HashFile(path string) string {
  sum, err := imohash.SumFile(path)
  if (err != nil) { fmt.Println("[ERR] ", err) }
  return hex.EncodeToString(sum[:16])
}


// Function to safely and abstractly load template files.
func LoadFile(path string) ([]byte, error) {

  // check if file exists
  _, err := os.Stat(path)
  if err != nil { return []byte{}, err }

  data, err := ioutil.ReadFile(path)
  if err != nil { return []byte{}, err }

  return data, nil
}


// Function to safely and abstractly load template files.
func LoadFileAsync(path string, out chan FileAsyncOutput) {

  // check if file exists
  _, err := os.Stat(path)
  if err != nil {
    out <- FileAsyncOutput{ []byte{}, errors.New("File is missing: " + path)}
    return
  }

  data, err := ioutil.ReadFile(path)
  if err != nil {
    out <- FileAsyncOutput{ []byte{}, err}
    return
  }

  out <- FileAsyncOutput{data, nil}
}

func AppendFile(filename string, data []byte) error {
  if (noWrite) {
    return nil
  } else {
    f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
    if (err != nil) {
      return err
    }

    defer f.Close()

    _, err = f.WriteString(string(data))
    if (err != nil) {
      return err
    }
    return nil
  }
}

func WriteFile(path string, data []byte, perm os.FileMode) error {
  if (noWrite) {
    return nil
  } else {
    return ioutil.WriteFile(path, data, perm)
  }
}

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
    pathS := strings.Split(path, "/")
    breadCrumb := ""
    crumbAddr := ""
    crumbSep := `<a class="smaller" href="#"> > </a>`
    crumbItem := `<a class="smaller" href="$addr$">$name$</a>`
    for index, crumb := range pathS {
        // First do crumbname
        if crumb == "." {
          breadCrumb += strings.Replace(crumbItem, "$name$", "", -1)
        } else {
            breadCrumb += strings.Replace(crumbItem, "$name$", strings.Trim(crumb, "./"), -1)
        }

        // Then crumb's link address
        if path == "." {
          crumbAddr = `#`
        } else {
          for i:=0; i < (len(pathS) - index+1); i++ { crumbAddr += (`../`) }
            crumbAddr += `./`
        }

        breadCrumb += strings.Replace(breadCrumb, "$addr$", crumbAddr, -1)

        // If is not last item in list, append >
        if crumb != pathS[len(pathS)-1] {
          breadCrumb += crumbSep
        }
    }
  return breadCrumb
}
