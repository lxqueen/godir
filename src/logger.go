package main

import (
    "fmt"
    "time"
    "os"
)

type LogObject struct {
    Lv          int
    EnableILogs bool
    Silent      bool
    OutFile     string
    FileValid   bool
}

/*
*=>  HOW TO USE LOGGER
*=>  Use the logger function to create a new logger object.
*=>  `debug := logger(level int, ilogs bool, quiet bool, ofile string)`
*=>  -1 = Disabled. 0 = INFO, 1 = WARN, 2 = ERROR, 3 = Internal Logs. FATAL will trigger at all levels aside from -1.
*=>  Manually set doIL to log ILOGS without changing overall log level. Basically -v flag.
*=>  The --quiet (-q) flag will only print WARNINGs and ERRORS (as well as FATAL messages, which print always.)
*=>  The --super-quiet (-qq) flag sets it to be COMPLETELY SILENT. Only FATAL Errors will appear.
*/

// Pass `nil` to oFile to disable file logging
func Logger(level int, sendILogs bool, quiet bool, oFile string) *LogObject {
  if oFile == "" { // Check to see if the ofile is valid..
    // Return a pointer to a new object.
    return &LogObject{Lv: level, EnableILogs: sendILogs, Silent: quiet, OutFile: oFile, FileValid: true}
  } else { // If ofile does not exist return NO_FILE as ofile.
    return &LogObject{Lv: level, EnableILogs: sendILogs, Silent: quiet, OutFile: "NO_FILE", FileValid: false}
  }
}

// Standard log level
func (d *LogObject) Log(msg string) {
  if ((d.Lv >= 0) && (d.Silent == false)) { // Only trigger if logging is enabled.
    fmt.Printf("[LOG][%s] %s\n", time.Now().Format("3:04 PM"), msg)
  }
}

func (d *LogObject) Warn(msg string) {
  if ((d.Lv >= 1) && (d.Silent == false)) { // Only trigger if logging is enabled.
    fmt.Printf("[WARN][%s] %s\n", time.Now().Format("3:04 PM"), msg)
  }
}

func (d *LogObject) Error(msg string) {
  if ((d.Lv >= 2) && (d.Silent == false)) { // Only trigger if logging is enabled.
    fmt.Printf("[ERROR][%s] %s\n", time.Now().Format("3:04 PM"), msg)
  }
}

// Fatal error. Always prints and exits 1.
func (d *LogObject) Fatal(msg string) {
  fmt.Printf("[FATAL][%s] %s\n", time.Now().Format("3:04 PM"), msg)
  os.Exit(1)
}

// Does not exit 1 afterwards. This should not be used, instead, use [ERROR].
func (d *LogObject) Fatals(msg string) {
  fmt.Printf("[FATAL][%s] %s\n", time.Now().Format("3:04 PM"), msg)
}

func (d *LogObject) Ilog(msg string) {
  if ((d.Lv >= 3) || (d.EnableILogs == true)) { // Only trigger if internal (super debug) is enabled.
    fmt.Printf("[IL][%s] %s\n", time.Now().Format("3:04 PM"), msg)
  }
}
