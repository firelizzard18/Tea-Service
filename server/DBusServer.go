package server

import (
	"errors"
	"log"

	"github.com/godbus/dbus"

	"github.com/firelizzard18/Tea-Service/common"
)

type DBusServer struct {
	proc *Process
	path dbus.ObjectPath
	bus  *dbus.Conn
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

	s.path = dbus.ObjectPath("/com/firelizzard/teasvc/Server")
	go s.handleSignals()

	err = s.bus.Export(s, s.path, "com.firelizzard.teasvc.Server")
	if err != nil {
		return nil, err
	}

	c := s.bus.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, "type='signal',interface='com.firelizzard.teasvc',member='Ping'")
	if c.Err != nil {
		return nil, c.Err
	}
	return s, nil
}

func (s *DBusServer) handleSignals() {
	ch := make(chan *dbus.Signal, 50)
	s.bus.Signal(ch)

	for sig := range ch {
		if sig.Name == "com.firelizzard.teasvc.Ping" {
			err := s.bus.Emit(s.path, "com.firelizzard.teasvc.Pong", s.proc.Description)
			if err != nil {
				log.Print(err)
			}
		}
	}
}

func newError(name string, body ...interface{}) *dbus.Error {
	return dbus.NewError(name, body)
}

func (s *DBusServer) RequestOutput(sender dbus.Sender, otype byte) (output dbus.UnixFD, derr *dbus.Error) {
	output = -1

	outPipe, err := s.proc.RequestOutput(common.OutputType(otype))
	if err != nil {
		derr = newError("com.firelizzard.teasvc.Server.RequestOutputFailure", err.Error())
		return
	}

	if outPipe != nil {
		output = dbus.UnixFD(outPipe.Fd())
	}

	return
}

func (s *DBusServer) RequestCommand(sender dbus.Sender, otype byte) (input, output dbus.UnixFD, derr *dbus.Error) {
	input = -1
	output = -1

	inPipe, outPipe, err := s.proc.RequestCommand(common.OutputType(otype))
	if err != nil {
		derr = newError("com.firelizzard.teasvc.Server.RequestCommandFailure", err.Error())
		return
	}

	if inPipe != nil {
		input = dbus.UnixFD(inPipe.Fd())
	}

	if outPipe != nil {
		output = dbus.UnixFD(outPipe.Fd())
	}

	return
}

func (s *DBusServer) SendCommand(sender dbus.Sender, otype byte, command string) (output dbus.UnixFD, derr *dbus.Error) {
	output = -1

	outPipe, err := s.proc.SendCommand(common.OutputType(otype), command)
	if err != nil {
		derr = newError("com.firelizzard.teasvc.Server.SendCommandFailure", err.Error())
		return
	}

	if outPipe != nil {
		output = dbus.UnixFD(outPipe.Fd())
	}

	return
}
