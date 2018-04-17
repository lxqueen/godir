package main

import (
    "encoding/json"
    "os"
    "time"
    "fmt"
    "github.com/otiai10/copy"
    "sync"
    "io/ioutil"
    "strings"
    "runtime"
)

type GenOpts struct {
  Conf Config
  Args Arguments

  ThemeHeader string
  ThemeFooter string
  ItemTemplate  string
}

// Global var - read-only for config.
var opts GenOpts

// Global console object for debugging and log output.
var console *LogObject

var noWrite bool = false // if true, will prevent *actually* writing files

var TimeNow string

func main() {
  // Time program execution
  start := time.Now()

  // Set a max number of processes equal to the amount of cpus.
  // https://reddit.com/r/golang/comments/290znn/goroutine_crazy_memory_usage
  runtime.GOMAXPROCS( runtime.NumCPU() )


  Ver := "0.1.7"  // Version
  Rev := "b"      // Revision (how many times has this version been committed to fix bugs.)

  // Just exit with a message if we're running with no arguments.
  if(len(os.Args) < 2) {
    fmt.Println("No valid arguments supplied. Use the `-h` flag for help.")
    os.Exit(0)
  }

  /*

  ARGUMENT PARSING AND DEBUG OUTPUT

  */

  // Load the args.
  opts = GenOpts{}
  opts.Args = ReadArgs()
  opts.Conf = ReadConfig(*opts.Args.ConfigFile)

  // --version output
  if ( *opts.Args.Version ) {
    fmt.Println("godir V." + Ver + Rev)
    fmt.Println("\ngodir is Licensed under the GNU GPL v3.\nCode copyright (c) 2018 Nicolas \"Montessquio\" Suarez.")
    os.Exit(0)
  }

  // func Logger(level int, sendILogs bool, quiet bool, oFile string) *_logger
  console = Logger(2, *opts.Args.Verbose, *opts.Args.Quiet, "")

  // *** Debug Mode Sanity Output *** //

  if ( *opts.Args.Verbose ) { // AKA test mode.
    console.Ilog("The following Args were Parsed:")
    /*
    fmt.Printf("Verbose %t\n", *args.Verbose)
    fmt.Printf("Version %t\n", *args.Version)
    fmt.Printf("Quiet %t\n", *args.Quiet)
    fmt.Printf("SQuiet %t\n", *args.SuperQuiet)
    fmt.Printf("Force %t\n", *args.Force)
    fmt.Printf("Webroot %q\n", *args.Webroot)
    fmt.Printf("Unjail %t\n", *args.Unjail)
    fmt.Printf("Filename %q\n", *args.Filename)
    fmt.Printf("Sort %t\n", *args.Sort)
    fmt.Printf("oFile %q\n", *args.OutFile)
    fmt.Printf("cfgFile %q\n", *args.ConfigFile)
    fmt.Printf("workPath %q\n", args.WorkPath)
    fmt.Printf("tail: %q\n", args.Tail)
    */
    data, err := json.Marshal(opts.Args)
    if err != nil { console.Fatal(err.Error()) }
    fmt.Printf("%s\n", data)

    console.Ilog("\nThe following configuration options were loaded from " + *opts.Args.ConfigFile + ":")
    data, err = json.Marshal(opts.Conf)
    if err != nil { console.Fatal(err.Error()) }
    fmt.Printf("%s\n", data)

    console.Log("Loaded args and configs in " + time.Since(start).String())
  }

  /*

  LOAD TEMPLATE FILES

  */

  console.Ilog("Loading template files into memory.")

  // Goroutines to load these, since they may be large.
  // Channels
  chanTheme := make( chan FileAsyncOutput )
  chanSearch := make( chan FileAsyncOutput )
  chanItem := make( chan FileAsyncOutput )

  timer := time.Now() // Timer

  go LoadFileAsync(opts.Conf.ThemeTemplate, chanTheme)
  go LoadFileAsync(opts.Conf.SearchTemplate, chanSearch)
  go LoadFileAsync(opts.Conf.ItemTemplate, chanItem)

  // Receive, then verify, each template file.
  themeOut := <- chanTheme
  if (themeOut.Err != nil) {
    console.Fatal(themeOut.Err.Error())
  }
  themeRaw := themeOut.Data


  searchOut := <- chanSearch
  if (searchOut.Err != nil) {
    console.Fatal(searchOut.Err.Error())
  }
  searchRaw := searchOut.Data


  itemOut := <- chanItem
  if (itemOut.Err != nil) {
    console.Fatal(itemOut.Err.Error())
  }
  itemRaw := itemOut.Data

  console.Log("Loaded 3 template files in " + time.Since(timer).String())

  console.Ilog("Theme file sum: " + HashBytes([]byte(themeRaw)))
  console.Ilog("Search file sum: " + HashBytes([]byte(searchRaw)))
  console.Ilog("Item file sum: " + HashBytes([]byte(itemRaw)))


  /*

  CREATE PROGRESS BAR

  */

  // Change directory into workpath.
  // Loop through workpath and count how much we have to process.

  // The async function is insanely fast.
  console.Log("Counting objects...")
  timer = time.Now()
  outChan := make(chan int)
  go DirTreeCountAsync(opts.Args.WorkPath, opts.Conf.Excludes, outChan)
  memberCount := <- outChan
  console.Log("Found ", memberCount, " objects in ", time.Since(timer))

  /*

    MAIN PROGRAM START

  */
  console.Ilog("Performing static substitutions...")
  // Sub and write the page header
  themeTmp := strings.Split(SubTag(string(themeRaw), opts.Conf.Tag_domain, opts.Conf.Domain), opts.Conf.Tag_contents) // Split at the contents tag to make a distinct header and footer.
  opts.ThemeHeader = themeTmp[0]
  opts.ThemeFooter = themeTmp[len(themeTmp)-1]
  searchText := SubTag(string(searchRaw), opts.Conf.Tag_domain, opts.Conf.Domain)
  opts.ItemTemplate = SubTag(string(itemRaw), opts.Conf.Tag_domain, opts.Conf.Domain)
  console.Ilog("Theme text sum: " + HashBytes([]byte(opts.ThemeHeader + opts.ThemeFooter)))
  console.Ilog("Search text sum: " + HashBytes([]byte(searchText)))
  console.Ilog("Item text sum: " + HashBytes([]byte(opts.ItemTemplate)))


  console.Log("Copying includes from ", opts.Conf.Include_path, " to ", opts.Args.WorkPath + "/")
  var err error
  if (noWrite) {
    err = nil
  } else {
    err = copy.Copy(opts.Conf.Include_path, opts.Args.WorkPath)
  }
  if (err != nil) {
    console.Fatal(err)
  }

  console.Log("Copying search.html from ", opts.Conf.SearchTemplate, " to ", opts.Args.WorkPath + "/search.html")
  err = WriteFile(opts.Args.WorkPath + "/search.html", []byte(searchText), 0644)
  if (err != nil) {
    console.Fatal(err)
  }

  err = os.Chdir(opts.Args.WorkPath) // We are now in the workpath, and can use "." to refer to the current location.
  if (err != nil) { console.Fatal(err.Error()) }

  console.Log("Generating objects...")
  timer = time.Now()

  // Use this to start the bar: //bar.Start()
  // Use barIncrement() to increase the bar

  // Create the files.json file in includes/
  err = WriteFile("./include/files.json", []byte("var jsonText = '["), 0644)
  if (err != nil) {
    console.Fatal(err)
  }

  semaphore := make(chan struct{}, *opts.Args.MaxRoutines) // Semaphore to limit max running goroutines
  var wg sync.WaitGroup
  wg.Add(1)
  go GenerateAsync(".", &wg, semaphore)

  wg.Wait() // wait for completion.

  // Add a closing quote to file to make valid JS/JSON
  AppendFile("./include/files.json", []byte("]'"))

  // Now sanitize the JSON.
  if (!noWrite) {
    filesJSON, err := ioutil.ReadFile("./include/files.json")
    if (err != nil) { console.Fatal(err.Error()) }
    filesJSON = []byte(strings.Replace(string(filesJSON), ", ", "", 1))
    err = WriteFile("./include/files.json", []byte(filesJSON), 0644)
  }

  // Program end.
  console.Log("Done. Took ", time.Since(timer), " (From launch: ", time.Since(start), ")")
}
