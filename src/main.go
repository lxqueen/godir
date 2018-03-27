package main

import (
    "encoding/json"
    "fmt"
    "os"
    "time"
    "gopkg.in/cheggaaa/pb.v1" // Used later
)

func main() {

  // Time program execution
  start := time.Now()


  Ver := "0.1.2"  // Version
  Rev := "a"      // Revision (how many times has this version been committed to fix bugs.)

  // Just exit with a message if we're running with no arguments.
  if(len(os.Args) < 2) {
    fmt.Println("No arguments supplied. Use the `-h` flag for help.")
    os.Exit(0)
  }

  /*

  ARGUMENT PARSING AND DEBUG OUTPUT

  */

  // Load the args.
  args := ReadArgs()
  config := ReadConfig(*args.ConfigFile)

  // --version output
  if ( *args.Version ) {
    fmt.Println("godir V." + Ver + Rev)
    fmt.Println("\ngodir is Licensed under the GNU GPL v3.\nCode copyright (c) 2018 Nicolas \"Montessquio\" Suarez.")
    os.Exit(0)
  }

  // func Logger(level int, sendILogs bool, quiet bool, oFile string) *_logger
  console := Logger(2, *args.Verbose, *args.Quiet, *args.OutFile)

  // *** Debug Mode Sanity Output *** //

  if ( *args.Verbose ) { // AKA test mode.
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
    data, err := json.Marshal(args)
    if err != nil { console.Fatal(err.Error()) }
    fmt.Printf("%s\n", data)

    console.Ilog("\nThe following configuration options were loaded from " + *args.ConfigFile + ":")
    data, err = json.Marshal(config)
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

  go LoadFileAsync(config.ThemeTemplate, chanTheme)
  go LoadFileAsync(config.SearchTemplate, chanSearch)
  go LoadFileAsync(config.ItemTemplate, chanItem)

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

  console.Ilog("Theme file sum: " + Hash([]byte(themeRaw)))
  console.Ilog("Search file sum: " + Hash([]byte(searchRaw)))
  console.Ilog("Item file sum: " + Hash([]byte(itemRaw)))


  /*

  CREATE PROGRESS BAR

  */

  // Change directory into workpath.
  // Loop through workpath and count how much we have to process.

  // The async function is insanely fast.
  console.Log("Counting objects...")
  timer = time.Now()
  outChan := make(chan int)
  go DirTreeCountAsync(args.WorkPath, outChan)
  memberCount := <- outChan
  console.Log("Found ", memberCount, " objects in ", time.Since(timer))

  // Set up the bar itself, now that we know how much we need to do.
  //bar := pb.New(memberCount)

  // Use this to start the bar: //bar.Start()
  // Use barIncrement() to increase the bar

  // Program end.
  console.Log("Done. Took ", time.Since(start))
}
