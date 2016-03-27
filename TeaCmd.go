package main

type TeaCmd int

const (
   Command_None = iota
   Command_Help
   Command_Interactive

   Command_List

   Command_Connect
   Command_Command
   
   Command_Launch

   Command_Invalid
)

func getCommand(arg string) TeaCmd {
   switch arg {
   case "":
      return Command_None

   case "help":
      return Command_Help

   case "interactive":
      return Command_Interactive

   case "list":
      return Command_List

   case "connect":
      return Command_Connect

   case "command":
      return Command_Command

   case "launch":
      return Command_Launch

   default:
      return Command_Invalid
   }
}