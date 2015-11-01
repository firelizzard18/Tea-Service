package main

import (
   "time"
   "errors"
   "github.com/godbus/dbus"
   "os"
   "log"
   "fmt"
)

type empty struct {}

type ServerInfo struct {
   Path dbus.ObjectPath
   Description string
}

type DBusClient struct {
   bus *dbus.Conn
   path dbus.ObjectPath
   sigchans map[string](chan *dbus.Signal)
}

func ConnectToDBus(bus string) (*DBusClient, error) {
   c := new(DBusClient)
   
   var err error
   switch bus {
   case "session":
      c.bus, err = dbus.SessionBus()
      
   case "system":
      c.bus, err = dbus.SystemBus()
      
   default:
      c.bus, err = dbus.Dial(bus)
   }
   
   if err != nil {
      return nil, err
   }
   if !c.bus.SupportsUnixFDs() {
      return nil, errors.New("DBus connection does not support file descriptors")
   }
   
   path := fmt.Sprintf("/com/firelizzard/teasvc/%d/Client", os.Getpid())
   c.path = dbus.ObjectPath(path)
   
   c.sigchans = make(map[string](chan *dbus.Signal))
   chsig := make(chan *dbus.Signal, 10)
   
   go func() {
      for {
         sig := <-chsig
         ch, ok := c.sigchans[sig.Name]
         if !ok {
            log.Print("Unhandled signal: " + sig.Name)
         }
         
         select {
         case ch <- sig:
            // sent singal, done
            
         default:
            log.Print("Unhandled signal (full channel): " + sig.Name)
         }
      }
   }()
   c.bus.Signal(chsig)
   c.bus.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, "type='signal',interface='com.firelizzard.teasvc',member='Pong'")
   
   return c, nil
}

func (c *DBusClient) ListServers(timeout int) chan *ServerInfo {
   if _, ok := c.sigchans["com.firelizzard.teasvc.Pong"]; ok {
      panic("This client is already pinging")
   }
   
   list := make(chan *ServerInfo, 10)
   found := make(map[dbus.ObjectPath]empty)
   
   chsig := make(chan *dbus.Signal, 50)
   chtime := make(chan empty)
   
   go func() {
      for {
         select {
            case sig := <- chsig:
               // if multiple clients simultaneously ping
               // we may receive multiple pongs 
               if _, ok := found[sig.Path]; ok {
                  continue
               }
               
               server := new(ServerInfo)
               server.Path = sig.Path
               
               var ok bool
               if len(sig.Body) > 0 {
                  if server.Description, ok = sig.Body[0].(string); !ok {
                     server.Description = "No description"
                  }
               }
               
               list <- server
               found[sig.Path] = empty{}
               
            case <- chtime:
               close(list)
               close(chsig)
               close(chtime)
               delete(c.sigchans, "com.firelizzard.teasvc.Pong")
               return
         }
      }
   }()
   c.sigchans["com.firelizzard.teasvc.Pong"] = chsig
   c.bus.Emit(c.path, "com.firelizzard.teasvc.Ping")
   
   go func() {
      time.Sleep(time.Duration(timeout) * time.Millisecond)
      chtime <- empty{}
   }()
   
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
