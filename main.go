package teasvc

import (
//   "github.com/firelizzard18/goopt"
   goopt "github.com/firelizzard18/goopt-fluent"
   "os"
   "log"
   "errors"
)

var asDaemon = goopt.
         DefineOption("daemon", "run as a daemon in the background").
            ShortNames('D').
            Names("daemon", "background").
         DefineAlternate("run in the foreground").
            ShortNames('F').
            Names("foreground").
         AsFlag(false)

var inSrc, outDest, errDest *os.File

func init() {
   goopt.DefineOption("input", "standard input source").
            ShortNames('i').
            Names("input").
            Process(parseFileOpt(os.Stdin, &inSrc)).
         DefineAlternate("standard input file").
            ShortNames('I').
            Names("input-file").
            Process(parseOptOpenFile(&inSrc))
   
   goopt.DefineOption("output", "standard output source").
            ShortNames('o').
            Names("output").
            Process(parseFileOpt(os.Stdout, &outDest)).
         DefineAlternate("standard output file").
            ShortNames('O').
            Names("output-file").
            Process(parseOptCreateFile(&outDest))
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
         return parseOptOpenFile(ptr)(o, os.DevNull)
      default:
         return errors.New("'" + input + "' is not a supported input source")
      }
      return nil
   }
}

func parseOptOpenFile(ptr **os.File) goopt.Processor {
   return func (o *goopt.Option, input string) error {
      if input == "" {
         return o.MissingArg()
      }
      
      var err error
      *ptr, err = os.Open(input)
      if (err != nil) {
         log.Fatalf("Failed to open '%s': %s", input, err.Error())
      }
      return nil
   }
}

func parseOptCreateFile(ptr **os.File) goopt.Processor {
   return func (o *goopt.Option, input string) error {
      if input == "" {
         return o.MissingArg()
      }
      
      var err error
      *ptr, err = os.Create(input)
      if (err != nil) {
         log.Fatalf("Failed to create '%s': %s", input, err.Error())
      }
      return nil
   }
}

func main() {
}

