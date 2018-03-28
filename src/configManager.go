package main

import (
    "fmt"
    "os"
    "github.com/BurntSushi/toml"
    "flag"
)

type Config struct {
	ThemeTemplate  string
  SearchTemplate string
  ItemTemplate   string
  Include_path   string

  Tag_contents   string
  Tag_class      string
  Tag_file_href  string
  Tag_item_type  string
  Tag_root_step  string
  Tag_domain     string
  Tag_root_dir   string
  Tag_sidenav    string
  Tag_breadcrumb string
  Tag_filename   string
  Tag_last_modified string
  Tag_filesize   string

  Follow_symlinks bool
  Excludes      []string
  Domain        string
}

type Arguments struct {
  Verbose   *bool
  Version   *bool
  Quiet     *bool
  SuperQuiet *bool
  Force     *bool
  Webroot   *string
  Unjail    *bool
  Filename  *string
  Sort      *bool
  OutFile   *string
  ConfigFile *string
  WorkPath  string
  Tail      []string
}

// Reads info from config file.
func ReadConfig(path string) Config {
	_, err := os.Stat(path)
	if err != nil {
		fmt.Println("Config file is missing: ", path)
    fmt.Println(err.Error())
    os.Exit(1)
	}

	var conf Config
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		fmt.Println(err.Error())
    os.Exit(1)
	}
	return conf
}

// Args has to be set before config, as the latter relies on the former.
func ReadArgs() Arguments {
  args := Arguments{}

  args.Verbose    = flag.Bool("v", false, "Verbose: Make the program give more detailed output.")
  args.Version    = flag.Bool("V", false, "Version: Get program version and some extra info.")
  args.Quiet      = flag.Bool("q",false, "Quiet: Decrease Logging Levels to Warning+")
  args.SuperQuiet = flag.Bool("qq",false, "Super Quiet: Makes the program not output to stdout, only displaying fatal errors.")
  args.Force      = flag.Bool("F", false, "Force: Force-regenerate all directories, even if no changes have been made.")
  args.Webroot    = flag.String("w", "", "Webroot: Specify a webroot to jail symlinks to.")
  args.Unjail     = flag.Bool("u", false, "Unjail: Use to remove the restriction jailing symlink destinations to the webroot.")
  args.Filename   = flag.String("f", "index.html", "File: Manually set the name of the HTML file containing the directory listing.")
  args.Sort       = flag.Bool("s", false, "Sort: Sort directory entries alphabetically.")
  args.OutFile    = flag.String("o", "", "lOgfile: Path to a text file to write program output to (file will be overwritten!). Use along with -qq to output to file and not stdout.")
  args.ConfigFile = flag.String("c", (os.Getenv("HOME") + "/.config/godir/config.toml"), "Specify a file to use as the godir config.")

  flag.Parse()

  tail := flag.Args()
  if (len(tail) == 0) {
    fmt.Println("No workpath supplied. Use the `-h` flag for help.")
  } else {
    args.WorkPath = tail[0]
    index, err := indexOf(tail, args.WorkPath)
    if (err != nil) {
      fmt.Println("indexOf failed with error: " + err.Error())
      os.Exit(1)
    }
    tail = remove(tail, index)
  }
  args.Tail = tail

  return args
}