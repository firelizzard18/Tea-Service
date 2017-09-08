package cmd

import (
	"errors"
	"os"

	"github.com/urfave/cli"

	"github.com/firelizzard18/Tea-Service/cmds/_"
	"github.com/firelizzard18/Tea-Service/server"
)

func init() {
	cmds.Register("launch", "launch a service").Executor(&Command{})
}

type Command struct {
	AsDaemon    bool   `short:"D" long:"daemon,background" description:"run as a daemon in the background"`
	Description string `short:"d" long:"desc,description" description:"service {description} as presented to clients`
	StdInFile   string `short:"i" long:"input" description:"read stdin from a {file}"`
	StdOutFile  string `short:"o" long:"output" description:"write stdout to a {file}"`
	StdErrFile  string `short:"e" long:"error" description:"write stderr to a {file}"`
}

func (m *Command) Execute(c *cli.Context) (err error) {
	var fin, fout, ferr *os.File

	args := c.Args()
	bus := c.GlobalString("dbus-bus")

	if m.AsDaemon {
		return errors.New("Golang does not support daemonization")
	}

	open := func(path string, file **os.File, flag int) bool {
		if path == "" {
			return true
		}
		if *file, err = os.OpenFile(path, flag, 0644); err != nil {
			return false
		}
		return true
	}

	if !open(m.StdInFile, &fin, os.O_RDONLY) ||
		!open(m.StdOutFile, &fout, os.O_CREATE|os.O_EXCL) ||
		!open(m.StdErrFile, &ferr, os.O_CREATE|os.O_EXCL) {
		return
	}

	proc, err := server.StartProcess(args[0], m.Description, bus, args[1:]...)
	if err != nil {
		return
	}

	if fin != nil {
		if err = proc.ConnectInput(fin); err != nil {
			return
		}
	}

	if fout != nil {
		proc.ConnectOutput(fout)
	}

	if ferr != nil {
		proc.ConnectError(ferr)
	}

	if err = proc.Start(); err != nil {
		return
	}

	if err = proc.Wait(); err != nil {
		return
	}

	return
}
