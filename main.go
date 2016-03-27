package main

import (
	"fmt"
	goopt "github.com/firelizzard18/goopt-fluent"
	"os"
   "errors"
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
   os.Exit(mainMain())
}

func mainMain() int {
   var cmd TeaCmd
   if len(os.Args) < 2 || os.Args[1][0] == '-' {
      cmd = Command_None
   } else {
      cmd = getCommand(os.Args[1])
   }

   switch cmd {
   case Command_None:
      opts, interactive := getInteractiveOptions()
      args, err := opts.Parse(os.Args)
      if err != nil {
         return mainUsage(err)
      }

      if *interactive {
         return mainInteractive(args)
      } else if len(args) > 0 {
         if args[0][1] == '-' {
            return mainUsage(errors.New("Unknown option: " + args[0]))
         } else {
            return mainUsage(errors.New("Unknown command: " + args[0]))
         }
      } else {
         return mainUsage(nil)
      }

   case Command_Help:
      return mainHelp(os.Args[1:])

   case Command_Interactive:
      return mainInteractive(os.Args[1:])

   case Command_List:
      return mainList(os.Args[1:])

   case Command_Connect:
      return mainConnect(os.Args[1:])

   case Command_Command:
      return mainCommand(os.Args[1:])

   case Command_Launch:
      return mainLaunch(os.Args[1:])

   case Command_Invalid:
   default:
      if os.Args[1][1] == '-' {
         return mainUsage(errors.New("Unknown option: " + os.Args[1]))
      } else {
         return mainUsage(errors.New("Unknown command: " + os.Args[1]))
      }
   }

   return -2
}

func mainUsage(err error) int {
   if err == nil {
      fmt.Print("Usage!")
   } else {
      fmt.Print("Usage! " + err.Error())
   }
   fmt.Print("\n")

   if err == nil {
      return 0
   } else {
      return -1
   }
}

func mainHelp(args []string) int {
   return 0
}