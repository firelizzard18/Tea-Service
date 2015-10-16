package teasvc

import (
//   "github.com/firelizzard18/goopt"
   goopt "github.com/firelizzard18/goopt-fluent"
   "os"
   "log"
   "errors"
)

var devNull *os.File

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

var description = serops.
         DefineOption("description", "service description as presented to clients").
            ShortNames('d').
            Names("desc", "description").
         AsString("")

var inSrc, outDest, errDest *os.File

func initServerModeOptions() {
   var err error
   devNull, err = os.Open(os.DevNull)
   if err != nil {
      log.Fatal(err)
   }
   
   serops.DefineOption("input", "standard input source").
            ShortNames('i').
            Names("input").
            Process(parseFileOpt(os.Stdin, &inSrc)).
         DefineAlternate("standard input file").
            ShortNames('I').
            Names("input-file").
            MutuallyExclusiveWithAlternates(true).
            ProcessAsFile(&inSrc, false, fileErrorHandler)
   
   serops.DefineOption("output", "standard output source").
            ShortNames('o').
            Names("output").
            Process(parseFileOpt(os.Stdout, &outDest)).
         DefineAlternate("standard output file").
            ShortNames('O').
            Names("output-file").
            MutuallyExclusiveWithAlternates(true).
            ProcessAsFile(&outDest, true, fileErrorHandler)
   
   serops.DefineOption("error", "standard error source").
            ShortNames('e').
            Names("error").
            Process(parseFileOpt(os.Stdout, &errDest)).
         DefineAlternate("standard error file").
            ShortNames('E').
            Names("error-file").
            MutuallyExclusiveWithAlternates(true).
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
         *ptr = std
      case "null":
         *ptr = devNull
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

var timeout = cliops.
         DefineOption("list-timeout", "timeout for listing available servers").
            ShortNames('t').
            Names("timeout").
         AsInt(100)


/* ----- Main ----- */

func init() {
   initServerModeOptions()
}

func usageAndExit(err error) {
   os.Exit(1)
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

func validateHandle(ptr **os.File, std *os.File) {
   if *ptr == devNull {
      *ptr = nil
   } else if !*asDaemon && *ptr == nil {
      *ptr = std
   }
}

func mainServer() {
   opts := goopt.MergeSets(genops, serops)
   args, err := opts.Parse(os.Args)
   if err != nil || len(args) < 2 || args[0] != "--" {
      usageAndExit(err)
   }
   
   if *asDaemon {
      log.Fatal("Golang does not support daemonization")
   }
   
   validateHandle(&inSrc, os.Stdin)
   validateHandle(&outDest, os.Stdout)
   validateHandle(&errDest, os.Stderr)
   
   proc, err := StartProcess(args[1], *description, *bus, args[2:]...)
   if err != nil {
      log.Fatal(err)
   }
   
   if inSrc != nil {
      err := proc.ConnectInput(inSrc)
      if err != nil {
         panic(err)
      }
   }
   
   if outDest != nil {
      proc.ConnectOutput(outDest)
   }
   
   if errDest != nil {
      proc.ConnectError(errDest)
   }
   
   err = proc.cmd.Run()
   if err != nil {
      log.Fatal(err)
   }
}

func mainClient() {
   opts := goopt.MergeSets(genops, cliops)
   args, err := opts.Parse(os.Args)
   if err != nil || len(args) > 0 {
      usageAndExit(err)
   }
   
   client, err := ConnectToDBus(*bus)
   if err != nil {
      log.Fatal(err)
   }
   
   if *list {
      client.ListServers(*timeout)
      return
   }
}
