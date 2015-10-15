package teasvc

import (
//   "github.com/firelizzard18/goopt"
   goopt "github.com/firelizzard18/goopt-fluent"
   "os"
   "log"
   "errors"
)

/* ----- General Options ----- */

var genops = goopt.NewOptionSet()

var bus = genops.
         DefineOption("dbus", "set the DBus bus").
            ShortNames('b').
            Names("bus", "dbus-bus").
         AsString("session")


/* ----- Server Mode Options ----- */

var serops = goopt.NewOptionSet()

var asDaemon = serops.
         DefineOption("daemon", "run as a daemon in the background").
            ShortNames('D').
            Names("daemon", "background").
         DefineAlternate("run in the foreground").
            ShortNames('F').
            Names("foreground").
         AsFlag(false)

var inSrc, outDest, errDest *os.File

func initServerModeOptions() {
   serops.DefineOption("input", "standard input source").
            ShortNames('i').
            Names("input").
            Process(parseFileOpt(os.Stdin, &inSrc)).
         DefineAlternate("standard input file").
            ShortNames('I').
            Names("input-file").
            Required().
            ProcessAsFile(&inSrc, false, fileErrorHandler)
   
   serops.DefineOption("output", "standard output source").
            ShortNames('o').
            Names("output").
            Process(parseFileOpt(os.Stdout, &outDest)).
         DefineAlternate("standard output file").
            ShortNames('O').
            Names("output-file").
            Required().
            ProcessAsFile(&outDest, true, fileErrorHandler)
   
   serops.DefineOption("error", "standard error source").
            ShortNames('e').
            Names("error").
            Process(parseFileOpt(os.Stdout, &errDest)).
         DefineAlternate("standard error file").
            ShortNames('E').
            Names("error-file").
            Required().
            ProcessAsFile(&errDest, true, fileErrorHandler)
}

func fileErrorHandler(input string, err error) error {
   log.Fatalf("Failed to create '%s': %s", input, err.Error())
   return nil
}

func parseFileOpt(std *os.File, ptr **os.File) goopt.Processor {
   return func (o *goopt.Option, input string) error {
      if input == "" {
         return o.MissingArg()
      }
      
      switch input {
      case "std":
         *ptr = os.Stdin
      case "null":
//         return parseOptOpenFile(ptr)(o, os.DevNull)
          return func (_ *goopt.Option, _ string) error {
             *ptr = nil
          }()
      default:
         return errors.New("'" + input + "' is not a supported input source")
      }
      return nil
   }
}

/* ----- Client Mode Options ----- */

var cliops = goopt.NewOptionSet()

var list = cliops.
         DefineOption("list-servers", "list the available servers").
            ShortNames('l').
            Names("ls", "list", "list-servers").
         AsFlag(false)


/* ----- Main ----- */

func init() {
   initServerModeOptions()
}

func usageAndExit(err error) {
   os.Exit(1)
}

func main() {
   for _, arg := os.Args {
      if arg == "--" {
         mainServer()
         return
      }
   }
   mainClient()
}

func mainServer() {
   opts := goopt.MergeSets(genops, serops)
   args, err := opts.Parse(os.Args)
   if err != nil || len(args) < 1 {
      usageAndExit(err)
   }
   
   proc, err2 := StartProcess(args[0], "What the fuck is the description?", args[1:])
   if err2 != nil {
      usageAndExit(err2)
   }
}

func mainClient() {
   opts := goopt.MergeSets(genops, cliops)
   args, err := opts.Parse(os.Args)
   if err != nil || len(args) > 0 {
      usageAndExit(err)
   }
}
