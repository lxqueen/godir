package main

import (
  "os"
  "math"
  "strconv"
  "io/ioutil"
  "path/filepath"
  "github.com/OneOfOne/xxhash"
)

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
func DirSize(path string) (int64, error) {
    var size int64
    err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
        if !info.IsDir() {
            size += info.Size()
        }
        return  err
    })
    return size, err
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

func DirTreeCountAsync(path string, out chan int) {
  files,_ := ioutil.ReadDir(path)
  count := 0
  subdirs := 0
  // First count the amount of subdirectories.
  for _, f := range files {
    if (f.IsDir()) { subdirs++ }
  }

  // Make a child channel that holds an amount of ints equal to the amount of subdirectories.
  childChan := make(chan int, subdirs)

  // Now actually delve into subdirs recursively
  for _, f := range files {
      if (f.IsDir()) {
        go DirTreeCountAsync(path + "/" + f.Name(), childChan)
      }
      count++
  }

  // Now sum all elements in the channel to get the total subdirs count.
  for i := 0; i < subdirs; i++ {
		count += <- childChan
	}
  out <- count
}

func Hash(data []byte) string {
  h := xxhash.New64()
	h.Write(data)
  return strconv.FormatUint(h.Sum64(), 16)
}

/* https://stackoverflow.com/questions/3173320/text-progress-bar-in-the-console
func printProgressBar (iteration, total, prefix = '', suffix = '', decimals = 1, length = 100, fill = '█'):
    """
    Call in a loop to create terminal progress bar
    @params:
        iteration   - Required  : current iteration (Int)
        total       - Required  : total iterations (Int)
        prefix      - Optional  : prefix string (Str)
        suffix      - Optional  : suffix string (Str)
        decimals    - Optional  : positive number of decimals in percent complete (Int)
        length      - Optional  : character length of bar (Int)
        fill        - Optional  : bar fill character (Str)
    """
    if iteration >= total:
        pass
    else:
        percent = ("{0:." + str(decimals) + "f}").format(100 * (iteration / float(total)))
        filledLength = int(length * iteration // total)
        bar = fill * filledLength + '-' * (length - filledLength)
        print('\r%s [%s] %s%% %s' % (prefix, bar, percent, suffix), end = '\r')
*/