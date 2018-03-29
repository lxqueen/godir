# [GODIR](https://github.com/montessquio/godir)

## Generate static HTML pages for directory listings.

This program will scan a given directory and all it contains, recursively, and within each of those directories (including the root) create a directory listing HTML page that not only looks nice but is completely static - removing the need for often bandwidth-intensive client-side handlers. Not only that, but it also has support for custom styles, so no more of those ugly default Apache directory listings.

It first will compile HTML for the folders, then takes file names and calculates file sizes for their listings.

# Compatibility

This program is a golang rewrite of [pydir](https://github.com/montessquio/pydir)

# Requirements
 - GO compiler and tools.

# Usage

`godir /path/to/root/working/directory`(Note no ending forward slash!) You may use `.` or any relative naming to describe the directory. Please note that the program will write files to the chosen directory. Also note that the `include` folder will be copied in its entirety to the root directory of the target.

Here is the help program help message, when run with `-h` or `--help`. Please note that this *may* not be up to date at all times.

Also note that windows users will always need to manually specify the `config.toml` file, because of the way home directories work on non-unix systems.

```
Usage of ./godir:
  godir WorkPath

Additional Arguments:
  -F	Force: Force-regenerate all directories, even if no changes have been made.
  -V	Version: Get program version and some extra info.
  -c string
    	Specify a file to use as the godir config. (default "$HOME/.config/godir/config.toml")
  -f string
    	File: Manually set the name of the HTML file containing the directory listing. (default "index.html")
  -q	Quiet: Decrease Logging Levels to Warning+
  -qq
    	Super Quiet: Makes the program not output to stdout, only displaying fatal errors.
  -s	Sort: Sort directory entries alphabetically.
  -u	Unjail: Use to remove the restriction jailing symlink destinations to the webroot.

```

# Configuration

Godir is configured in two places: in `config.toml` (below) and using command line options (above).


The default godir config file is a (TOML)[https://github.com/toml-lang/toml] file.

By default godir looks in `$HOME/.config/godir/config.toml`. You may specify a different TOML file to use with the `-c` flag. You can find a sample config file in `src/config.toml.example`.

# Performance
Godir has been found to be 650 times faster than Pydir (when tested on very large file trees.)

Benchmarking has shown godir to be able to completely generate for a ~250gb (200,000 files) tree in just under 15 seconds.
Please note these numbers may change depending on the nature, distribution, and size of your filesystem.

# Search
Search is reasonably fast, and since it uses [lunr.js](https://lunrjs.com) it supports wildcards (`*`) and boosts (`^x`), which allow you to prioritize certain parts of the search by `x` times.

# Examples
See an example page listing (may not always be up to date) at [static.dnd.guide](https://static.dnd.guide)

# Theme-Making
GoDir-gen uses tags to search and insert certain values. Knowing these tags and inserting them into your theme and itemtemplate HTML files is necessary for the tool to work.

First is the `$content$` tag. This goes in your main `theme.html` file and will be replaced with a single instance of `item-template.html` for every file or folder in the current directory.

Inside the `item-template.html` file, there are more tags. The `$class$` tag will be replaced with `icon dir` or `icon file` (referring to css classes) for directories and files, respectively.

The `$file-href$` tag will be replaced with the relative path to the file that that specific entry points to.

`$item-type$ ` will be replaced with `icon dir-icon` or `icon file-icon`, which sets the icons to the left of the entries.

`$root-step$` will be replaced with an amount of `../`'s required to get to the root site directory you specified as a comand line argument (NOT \_WEBROOT).

`$domain$` will be replaced with the domain specified in the config. `$root-dir$` refers to the name of the folder that the final html file will be placed in.

`$sidenav$` will be replaced with an unordered, dropdown list representing a directory tree of the website.

`$breadcrumb$` will be replaced with a hyperlinked breadcrumb leading back to the root working directory.

# FAQ
- Q: Help! Symlinks aren't showing up!
  - A: First, make sure the `FOLLOWSYMLINKS` option is set to true in the config file. If that doesn't work, then make sure you have set the proper webroot in `cfg.py` and there is no trailing slash `/` at the end of it. If you want the symlinks to go out of your set webroot, make sure the appropriate setting is also changed in the config.

- Q: Can I use this for my commercial product/website without crediting you?
  - A: Absolutely! There are a few requirements, though. First, this software must retain the MIT license. Second, I am not liable for what this program does or what you do with it. Use it at your own risk.

- Q: Why is this FAQ so short?
  - A: People don't ask many questions.

# License and Attribution
This work is licensed under the GNU GPL V.3.0 license. Some parts of this work may be licensed differently.
Additionally, there is an attribution in the footer of the default template. Feel free to edit or remove it, but it would be appreciated if it were left in.
