package main

import (
	"errors"
	"github.com/godbus/dbus"
	"log"
	"os"
	"time"
)

type empty struct{}

type ServerInfo struct {
	Sender      string
	Description string
}

type DBusClient struct {
	bus      *dbus.Conn
	path     dbus.ObjectPath
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

	c.path = dbus.ObjectPath("/com/firelizzard/teasvc/Client")
	err = c.bus.Export(c, c.path, "com.firelizzard.teasvc.Client")
	if err != nil {
		return nil, err
	}

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
			case sig := <-chsig:
				// if multiple clients simultaneously ping
				// we may receive multiple pongs
				if _, ok := found[sig.Path]; ok {
					continue
				}

				server := new(ServerInfo)
				server.Sender = sig.Sender

				ok := len(sig.Body) > 0
				if ok {
					server.Description, ok = sig.Body[0].(string)
				}
				if !ok {
					server.Description = "No description"
				}

				list <- server
				found[sig.Path] = empty{}

			case <-chtime:
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

func (c *DBusClient) RequestOutput(dest string, otype OutputType) (*os.File, error) {
	var output dbus.UnixFD

	if dest == "" {
		return nil, errors.New("Invalid destination")
	}

	obj := c.bus.Object(dest, "/com/firelizzard/teasvc/Server")
	err := obj.Call("com.firelizzard.teasvc.Server.RequestOutput", 0, byte(otype)).Store(&output)
	if err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(output), "out pipe"), nil
}

func (c *DBusClient) RequestCommand(dest string, otype OutputType) (*os.File, *os.File, error) {
	var input, output dbus.UnixFD

	if dest == "" {
		return nil, nil, errors.New("Invalid destination")
	}

	obj := c.bus.Object(dest, "/com/firelizzard/teasvc/Server")
	err := obj.Call("com.firelizzard.teasvc.Server.RequestCommand", 0, byte(otype)).Store(&input, &output)
	if err != nil {
		return nil, nil, err
	}

	return os.NewFile(uintptr(input), "in pipe"), os.NewFile(uintptr(output), "out pipe"), nil
}

func (c *DBusClient) SendCommand(dest string, otype OutputType, command string) (*os.File, error) {
	var output dbus.UnixFD

	if dest == "" {
		return nil, errors.New("Invalid destination")
	}

	obj := c.bus.Object(dest, "/com/firelizzard/teasvc/Server")
	err := obj.Call("com.firelizzard.teasvc.Server.SendCommand", 0, byte(otype), command).Store(&output)
	if err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(output), "out pipe"), nil
}
