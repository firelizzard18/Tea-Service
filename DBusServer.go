package main

import (
   "errors"
   "os"
   "fmt"
   "log"
   "github.com/godbus/dbus"
)

type DBusServer struct {
   proc *Process
   path dbus.ObjectPath
   bus *dbus.Conn
}

func ExportToDBus(proc *Process, bus string) (*DBusServer, error) {
   s := new(DBusServer)
   s.proc = proc
   
   var err error
   switch bus {
   case "session":
      s.bus, err = dbus.SessionBus()
      
   case "system":
      s.bus, err = dbus.SystemBus()
      
   default:
      s.bus, err = dbus.Dial(bus)
   }
   
   if err != nil {
      return nil, err
   }
   if !s.bus.SupportsUnixFDs() {
      return nil, errors.New("DBus connection does not support file descriptors")
   }
   
   path := fmt.Sprintf("/com/firelizzard/teasvc/%d/Server", os.Getpid())
   s.path = dbus.ObjectPath(path)
   
   go s.handleSignals()
   
   s.bus.Export(s, s.path, "com.firelizzard.teasvc.Server")
   return s, nil
}

func (s *DBusServer) handleSignals() {
   ch := make(chan *dbus.Signal, 50)
   s.bus.Signal(ch)
   
   for sig := range ch {
      if (sig.Name == "com.firelizzard.teasvc.Ping") {
         err := s.bus.Emit(s.path, "com.firelizzard.teasvc.Pong", s.proc.Description)
         if err != nil {
            log.Print(err)
         }
      }
   }
}

func (s *DBusServer) RequestOutput(sender dbus.Sender, otype OutputType) (dbus.UnixFDIndex, error) {
   output, err := s.proc.RequestOutput(otype)
   if err != nil {
      return 0, err
   }

   return dbus.UnixFDIndex(output.Fd()), nil
}

func (s *DBusServer) RequestCommand(sender dbus.Sender, otype OutputType) (dbus.UnixFDIndex, dbus.UnixFDIndex, error) {
   input, output, err := s.proc.RequestCommand(otype)
   if err != nil {
      return 0, 0, err
   }

   return dbus.UnixFDIndex(input.Fd()), dbus.UnixFDIndex(output.Fd()), nil
}

func (s *DBusServer) SendCommand(sender dbus.Sender, otype OutputType, command string) (dbus.UnixFDIndex, error) {
   output, err := s.proc.SendCommand(otype, command)
   if err != nil {
      return 0, err
   }

   return dbus.UnixFDIndex(output.Fd()), nil
}