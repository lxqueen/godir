package main

import (
    "encoding/json"
    "fmt"
    "os"
    "time"
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
    if err != nil { console.Error(err.Error()) }
    fmt.Printf("%s\n", data)

    console.Ilog("\nThe following configuration options were loaded from " + *args.ConfigFile + ":")
    data, err = json.Marshal(config)
    if err != nil { console.Error(err.Error()) }
    fmt.Printf("%s\n", data)
  }


  // Program end.
  fmt.Printf("Done. Took %s\n", time.Since(start))
}
