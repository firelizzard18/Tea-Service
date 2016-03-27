package main

import (
   "errors"
   goopt "github.com/firelizzard18/goopt-fluent"
   "log"
)

func getInteractiveOptions() (set *goopt.OptionSet, interactive *bool){
   set = goopt.NewOptionSet()

   interactive = set.
      DefineOption("interactive", "runs teasvc in interactive mode").
         ShortNames('i').
         Names("interactive").
      AsFlag(false)

   return
}

func mainInteractive(args []string) int {
   if len(args) > 0 {
      return mainUsage(errors.New("Interactive mode does not take any arguments"))
   }

   log.Print("Interactive mode has not been implemented")
   return -1
}