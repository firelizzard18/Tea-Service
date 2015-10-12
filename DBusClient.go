package teasvc

import (
   "time"
   "errors"
   "github.com/godbus/dbus"
   "os"
)

type empty struct{}
type ServerInfo struct {
   Path dbus.ObjectPath
   Description string
}

type DBusClient struct {
   bus *dbus.Conn
   
   siglock chan empty
}

func ConnectToDBus() (*DBusClient, error) {
   c := new(DBusClient)
   
   var err error
   c.bus, err = dbus.SessionBus()
   if err != nil {
      return nil, err
   }
   if !c.bus.SupportsUnixFDs() {
      return nil, errors.New("DBus connection does not support file descriptors")
   }
   
   c.siglock = make(chan empty, 1)
   c.siglock <- empty{}
   
   return c, nil
}

func (c *DBusClient) ListServers(timeout int) chan *ServerInfo {
   <- c.siglock
   
   list := make(chan *ServerInfo, 10)
   found := make(map[dbus.ObjectPath]empty)
   
   chsig := make(chan *dbus.Signal, 50)
   chtime := make(chan empty)
   c.bus.Signal(chsig)
   
   go func() {
      for {
         select {
            case sig := <- chsig:
               var ok bool
               
               if sig.Name != "com.firelizzard.teasvc.Pong" {
                  continue
               }
               
               // if multiple clients simultaneously ping
               // we may receive multiple pongs 
               if _, ok = found[sig.Path]; ok {
                  continue
               }
               
               server := new(ServerInfo)
               server.Path = sig.Path
               
               if len(sig.Body) > 0 {
                  if server.Description, ok = sig.Body[0].(string); !ok {
                     server.Description = "No description"
                  }
               }
               
               list <- server
               found[sig.Path] = empty{}
               
            case <- chtime:
               close(list)
               c.bus.Signal(nil)
               close(chsig)
               close(chtime)
               c.siglock <- empty{}
               return
         }
      }
   }()
   
   go func() {
      time.Sleep(time.Duration(timeout) * time.Millisecond)
      chtime <- empty{}
   }()
   
   c.bus.Signal(chsig)
   return list
}

func (c *DBusClient) RequestOutput(path dbus.ObjectPath, otype OutputType) (*os.File, error) {
   var output dbus.UnixFDIndex
   
   obj := c.bus.Object("com.firelizard.teasvc.Server", path)
   err := obj.Call("com.firelizzard.teasvc.Server.RequestOutput", 0, otype).Store(&output)
   if err != nil {
      return nil, err
   }
   
   return os.NewFile(uintptr(output), "out pipe"), nil
}

func (c *DBusClient) RequestCommand(path dbus.ObjectPath, otype OutputType) (*os.File, *os.File, error) {
   var input, output dbus.UnixFDIndex
   
   obj := c.bus.Object("com.firelizard.teasvc.Server", path)
   err := obj.Call("com.firelizzard.teasvc.Server.RequestCommand", 0, otype).Store(&input, &output)
   if err != nil {
      return nil, nil, err
   }
   
   return os.NewFile(uintptr(input), "in pipe"), os.NewFile(uintptr(output), "out pipe"), nil
}

func (c *DBusClient) SendCommand(path dbus.ObjectPath, otype OutputType, command string) (*os.File, error) {
   var output dbus.UnixFDIndex
   
   obj := c.bus.Object("com.firelizzard.teasvc.Server", path)
   err := obj.Call("com.firelizzard.teasvc.Server.SendCommand", 0, otype, command).Store(&output)
   if err != nil {
      return nil, err
   }
   
   return os.NewFile(uintptr(output), "out pipe"), nil
}
