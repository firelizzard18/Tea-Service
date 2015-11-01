package main

import (
	"errors"
	"fmt"
	goopt "github.com/firelizzard18/goopt-fluent"
	"log"
	"os"
	"strings"
   "io"
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

var connect = cliops.
   DefineOption("connect", "connect to a server's output").
      Names("connect").
   AsString("")

var command = cliops.
   DefineOption("command", "connect to a server's input and output").
      Names("command").
   AsString("")

var send = cliops.
   DefineOption("send", "send a command to a server and connect to it's output").
      Names("send").
   AsString("")

var outType = cliops.
   DefineOption("out-type", "the output type when connecting to a server").
      Names("out-type").
   AsChoice(3, "none", "out", "err", "both")

/* ----- Main ----- */

func init() {
	initServerModeOptions()
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
	opts := goopt.NewMergedSet(genops, serops)
	args, err := opts.Parse(os.Args)
   if err != nil {
      usageAndExit(err)
   } else if len(args) > 0 && args[0] != "--" {
      usageAndExit(errors.New("Unparsed arguments: " + strings.Join(args, " ")))
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

func mainClient() {
   _ = "breakpoint"
	opts := goopt.NewMergedSet(genops, cliops)
	args, err := opts.Parse(os.Args)
	if err != nil {
		usageAndExit(err)
	} else if len(args) > 0 {
		usageAndExit(errors.New("Unparsed arguments: " + strings.Join(args, " ")))
	}

	client, err := ConnectToDBus(*bus)
	if err != nil {
		log.Fatal(err)
	}

	if *list {
		for server := range client.ListServers(*timeout) {
			log.Print(server)
		}
		return
	}

   if *connect != "" {
      out, err := client.RequestOutput(*connect, OutputType(*outType))
      if err != nil {
         log.Fatal(err)
      }

      dumpOutput(out)
      return
   }

   if *command != "" {
      if *send != "" {
         out, err := client.SendCommand(*command, OutputType(*outType), *send)
         if err != nil {
            log.Fatal(err)
         }

         dumpOutput(out)
      } else {
         in, out, err := client.RequestCommand(*command, OutputType(*outType))
         if err != nil {
            log.Fatal(err)
         }

         go sendInput(in)
         dumpOutput(out)
      }
      return
   }

	usageAndExit(nil)
}

func dumpOutput(source *os.File) {
   if source == nil {
      return
   }

   defer source.Close()

   b := make([]byte, 256)
   for {
      n, err := source.Read(b)
      if err == io.EOF {
         log.Print("The output file descriptor has been closed")
         return
      }
      
      for o := 0; o < n; {
         m, err := os.Stdout.Write(b[o:n])
         if err != nil {
            log.Panic(err)
            return
         }
         o += m
      }
   }
}

func sendInput(sink *os.File) {
   defer sink.Close()

   b := make([]byte, 256)
   for {
      n, err := os.Stdin.Read(b)
      if err == io.EOF {
         log.Panic("Stdin has been closed")
         return
      }
      
      for o := 0; o < n; {
         m, err := sink.Write(b[o:n])
         if err != nil {
            log.Print(err)
            return
         }
         o += m
      }
   }
}