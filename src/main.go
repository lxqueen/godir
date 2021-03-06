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

var sideNav string  // Contains the sideNav string.
                    // This is generated beforehand used in the generation goroutines.

var Ver string = "0.2.2"  // Version
var Rev string = ""      // Revision (how many times has this version been committed to fix bugs.)

func main() {
  // Time program execution
  start := time.Now()

  // Set a max number of processes equal to the amount of cpus.
  // https://reddit.com/r/golang/comments/290znn/goroutine_crazy_memory_usage
  runtime.GOMAXPROCS( runtime.NumCPU() )


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
  
  if *opts.Args.SideBarOnly {
    *opts.Args.Verbose = false
    *opts.Args.Quiet = true
  } else {
    opts.Conf = ReadConfig(*opts.Args.ConfigFile)
  }

  

  // func Logger(level int, sendILogs bool, quiet bool, oFile string) *_logger
  console = Logger(2, *opts.Args.Verbose, *opts.Args.Quiet, "")

  // *** Debug Mode Sanity Output *** //

  if ( *opts.Args.Verbose ) { // AKA test mode.
    console.Ilog("The following Args were Parsed:")
  
    data, err := json.Marshal(opts.Args)
    if err != nil { console.Fatal(err.Error()) }
    fmt.Printf("%s\n", data)

    console.Ilog("\nThe following configuration options were loaded from " + *opts.Args.ConfigFile + ":")
    data, err = json.Marshal(opts.Conf)
    if err != nil { console.Fatal(err.Error()) }
    fmt.Printf("%s\n", data)

    console.Log("Loaded args and configs in " + time.Since(start).String())
  }

  if *opts.Args.SideBarOnly {
    err := os.Chdir(opts.Args.WorkPath)
    if (err != nil) { console.Fatal(err.Error()) }
    GenSidenav(".", 0)
    fmt.Print(sideNav)
    os.Exit(0)
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
  console.Ilog("Loading file " + opts.Conf.ThemeTemplate)
  go LoadFileAsync(opts.Conf.SearchTemplate, chanSearch)
  console.Ilog("Loading file " + opts.Conf.SearchTemplate)
  go LoadFileAsync(opts.Conf.ItemTemplate, chanItem)
  console.Ilog("Loading file " + opts.Conf.ItemTemplate)

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


  // Change directory to the working path.
  err = os.Chdir(opts.Args.WorkPath) // We are now in the workpath, and can use "." to refer to the current location.
  if (err != nil) { console.Fatal(err.Error()) }
  

  // Generate the sidenav...
  console.Log("Beginning sidenav generation")
  GenSidenav(".", 0)

  /*

    MAIN PROGRAM START

  */
  console.Ilog("Performing static substitutions...")
  // Sub and write the searchText header
  themeTmp := strings.Split(SubTag(string(themeRaw), opts.Conf.Tag_domain, opts.Conf.Domain), opts.Conf.Tag_contents) // Split at the contents tag to make a distinct header and footer.
  opts.ThemeHeader = themeTmp[0]
  opts.ThemeFooter = themeTmp[len(themeTmp)-1]
  searchText := SubTag(string(searchRaw), opts.Conf.Tag_domain, opts.Conf.Domain)
  searchText = SubTag(searchText, opts.Conf.Tag_root_step, "./")
  searchText = SubTag(searchText, opts.Conf.Tag_breadcrumb, " / Search")
  searchText = SubTag(searchText, opts.Conf.Tag_root_dir, "./")
  searchText = SubTag(searchText, opts.Conf.Tag_domain, opts.Conf.Domain)
  searchText = SubTag(searchText, opts.Conf.Tag_root_step, "./")
  searchText = SubTag(searchText, opts.Conf.Tag_sidenav, sideNav)
  opts.ItemTemplate = SubTag(string(itemRaw), opts.Conf.Tag_domain, opts.Conf.Domain)
  console.Ilog("Theme text sum: " + HashBytes([]byte(opts.ThemeHeader + opts.ThemeFooter)))
  console.Ilog("Search text sum: " + HashBytes([]byte(searchText)))
  console.Ilog("Item text sum: " + HashBytes([]byte(opts.ItemTemplate)))

  console.Log("Copying search.html from ", opts.Conf.SearchTemplate, " to ", opts.Args.WorkPath,  "/search.html")
  err = WriteFile("./search.html", []byte(searchText), 0644)
  if (err != nil) {
    console.Fatal(err)
  }

  console.Log("Generating objects...")
  timer = time.Now()

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