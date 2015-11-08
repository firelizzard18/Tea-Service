package main

import (
	"fmt"
	goopt "github.com/firelizzard18/goopt-fluent"
	"os"
)

var genops = goopt.NewOptionSet()

var bus = genops.
	DefineOption("dbus", "set the DBus bus").
   	ShortNames('b').
   	Names("bus", "dbus-bus").
	AsString("session")

func init() {
	initServerModeOptions()
}

func main() {
	for _, arg := range os.Args {
		if arg == "--" {
			mainServer()
			return
		}
	}
	mainClient()
}

func usageAndExit(err error) {
   if err == nil {
      fmt.Print("Usage!")
   } else {
      fmt.Print("Usage! " + err.Error())
   }
   fmt.Print("\n")
   os.Exit(1)
}