package main

import (
    "fmt"
    "os"
    "github.com/BurntSushi/toml"
)

type Config struct {
	themeTemplate  string
  searchTemplate string
  itemTemplate   string
  tag_contents   string
  tag_class      string
  tag_file_href  string
  tag_item_type  string
  tag_root_step  string
  tag_domain     string
  tag_root_dir   string
  tag_sidenav    string
  tag_breadcrumb string
}

// Reads info from config file
func ReadConfig(path string) Config {
	_, err := os.Stat(path)
	if err != nil {
		fmt.Println("Config file is missing: ", path)
    fmt.Println(err.Error())
    os.Exit(1)
	}

	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		fmt.Println(err.Error())
    os.Exit(1)
	}
	//log.Print(config.Index)
	return config
}
