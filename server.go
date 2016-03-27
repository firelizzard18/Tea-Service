package main

import (
   "errors"
   goopt "github.com/firelizzard18/goopt-fluent"
   "log"
   "os"
   "strings"
)

var devNull *os.File

var serops = goopt.NewOptionSet()

var asDaemon = serops.
   DefineOption("daemon", "run in the foreground").
      ShortNames('F').
      Names("foreground").
   DefineAlternate("run as a daemon in the background").
      ShortNames('D').
      Names("daemon", "background").
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

   serops.
      DefineOption("input", "standard input source").
         ShortNames('i').
         Names("input").
         Process(parseFileOpt(os.Stdin, &inSrc)).
      DefineAlternate("standard input file").
         ShortNames('I').
         Names("input-file").
         MutuallyExclusiveWithAlternates(true).
         ProcessAsFile(&inSrc, false, fileErrorHandler)

   serops.
      DefineOption("output", "standard output source").
         ShortNames('o').
         Names("output").
         Process(parseFileOpt(os.Stdout, &outDest)).
      DefineAlternate("standard output file").
         ShortNames('O').
         Names("output-file").
         MutuallyExclusiveWithAlternates(true).
         ProcessAsFile(&outDest, true, fileErrorHandler)

   serops.
      DefineOption("error", "standard error source").
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
   return func(o *goopt.Option, input string) error {
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

func mainServer() {
   opts := goopt.NewMergedSet(genops, serops)
   args, err := opts.Parse(os.Args)
   if err != nil {
      mainUsage(err)
   } else if len(args) > 0 && args[0] != "--" {
      mainUsage(errors.New("Unparsed arguments: " + strings.Join(args, " ")))
   }

   if err := devNull.Close(); err != nil {
      log.Fatal(err)
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
      if err := proc.ConnectInput(inSrc); err != nil {
         log.Panic(err)
      }
   }

   if outDest != nil {
      proc.ConnectOutput(outDest)
   }

   if errDest != nil {
      proc.ConnectError(errDest)
   }

   if err = proc.Start(); err != nil {
      log.Fatal(err)
   }

   if err = proc.Wait(); err != nil {
      log.Fatal(err)
   }
}

func validateHandle(ptr **os.File, std *os.File) {
   if *ptr == devNull {
      *ptr = nil
   } else if !*asDaemon && *ptr == nil {
      *ptr = std
   }
}