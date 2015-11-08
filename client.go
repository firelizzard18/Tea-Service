package main

import (
   "errors"
   goopt "github.com/firelizzard18/goopt-fluent"
   "log"
   "os"
   "strings"
   "io"
)

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



func mainClient() {
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