package teasvc

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

func ExportToDBus(proc *Process) (*DBusServer, error) {
   s := new(DBusServer)
   s.proc = proc
   
   var err error
   s.bus, err = dbus.SessionBus()
   if err != nil {
      return nil, err
   }
   if !s.bus.SupportsUnixFDs() {
      return nil, errors.New("DBus connection does not support file descriptors")
   }
   
   path := fmt.Sprintf("/com/firelizzard/teasvc/%d/Server", os.Getpid())
   s.path = dbus.ObjectPath(path)
   
   ch := make(chan *dbus.Signal, 50)
   go s.handleSignals(ch)
   s.bus.Signal(ch)
   
   s.bus.Export(s, s.path, "com.firelizzard.teasvc.Server")
   return s, nil
}

func (s *DBusServer) handleSignals(ch chan *dbus.Signal) {
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