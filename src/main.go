package main

import (
    "fmt"
    "os"
    "flag"
)

func main() {

  Ver := "0.1.2"  // Version
  Rev := "a"      // Revision (how many times has this version been committed to fix bugs.)

  // Just exit with a message if we're running with no arguments.
  if(len(os.Args) < 2) {
    fmt.Println("No arguments supplied. Use the `-h` flag for help.")
    os.Exit(0)
  }
  // *** Argparse Stuff *** //

  verbose := flag.Bool("v", false, "Verbose: Make the program give more detailed output.")
  version := flag.Bool("V", false, "Version: Get program version and some extra info.")
  quiet   := flag.Bool("q",false, "Quiet: Decrease Logging Levels to Warning+")
  superQuiet := flag.Bool("qq",false, "Super Quiet: Makes the program not output to stdout, only displaying fatal errors.")
  force   := flag.Bool("F", false, "Force: Force-regenerate all directories, even if no changes have been made.")
  webroot := flag.String("w", "", "Webroot: Specify a webroot to jail symlinks to.")
  unjail   := flag.Bool("u", false, "Unjail: Use to remove the restriction jailing symlink destinations to the webroot.")
  filename   := flag.String("f", "", "File: Manually set the name of the HTML file containing the directory listing.")
  sort   := flag.Bool("s", false, "Sort: Sort directory entries alphabetically.")
  outFile   := flag.String("o", "", "lOgfile: Path to a text file to write program output to (file will be overwritten!). Use along with -qq to output to file and not stdout.")

  flag.Parse()

  // --version output
  if ( *version ) {
    fmt.Println("godir V." + Ver + Rev)
    fmt.Println("\ngodir is Licensed under the GNU GPL v3.\nCode copyright (c) 2018 Nicolas \"Montessquio\" Suarez.")
    os.Exit(0)
  }

  // *** END Argparse Stuff *** //

  // *** Logger Config *** //

  // func Logger(level int, sendILogs bool, quiet bool, oFile string) *_logger
  console := Logger(2, *verbose, *quiet, *outFile)
  // *** END Logger Config *** //

  // *** Set up tail *** //

  tail := flag.Args()
  workPath := tail[0]
  index, err := indexOf(tail, workPath)
  if (err != nil) {
    console.Error("indexOf failed with error: " + err.Error())
  }
  remove(tail, index)

  // ** End Set up tail *** //

  // *** Debug Mode Sanity Output *** //

  if ( *verbose ) { // AKA test mode.
    console.Ilog("The following Args were Parsed:")
    fmt.Printf("Verbose %t\n", *verbose)
    fmt.Printf("Version %t\n", *version)
    fmt.Printf("Quiet %t\n", *quiet)
    fmt.Printf("SQuiet %t\n", *superQuiet)
    fmt.Printf("Force %t\n", *force)
    fmt.Printf("Webroot %q\n", *webroot)
    fmt.Printf("Unjail %t\n", *unjail)
    fmt.Printf("Filename %q\n", *filename)
    fmt.Printf("Sort %t\n", *sort)
    fmt.Printf("oFile %q\n", *outFile)
    fmt.Printf("workPath %q\n", workPath)
    fmt.Printf("tail: %q\n", flag.Args())
  }

  // *** END Debug Mode Sanity Output *** //


}
