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
	Description string `short:"d" long:"desc,description" description:"service {description} as presented to clients"`
	InputFile   string `short:"i" long:"input" description:"read stdin from a {file}"`
	OutputFile  string `short:"o" long:"output" description:"write stdout to a {file}"`
	ErrorFile   string `short:"e" long:"error" description:"write stderr to a {file}"`
	UseStdIn    bool   `short:"I" long:"stdin" description:"pipe stdin to the process"`
	UseStdOut   bool   `short:"O" long:"stdout" description:"pipe stdout from the process"`
	UseStdErr   bool   `short:"E" long:"stderr" description:"pipe stderr from the process"`
}

func (m *Command) Execute(c *cli.Context) (err error) {
	var fin, fout, ferr *os.File

	args := c.Args()
	bus := c.GlobalString("dbus-bus")

	if m.AsDaemon {
		if m.UseStdIn || m.UseStdOut || m.UseStdErr {
			return errors.New("Process IO cannot be piped in daemon mode")
		}
		return errors.New("Golang does not support daemonization")
	}

	if m.UseStdIn {
		fin = os.Stdin
	} else if m.InputFile != "" {
		if fin, err = os.Open(m.InputFile); err != nil {
			return
		}
	}

	if m.UseStdOut {
		fout = os.Stdout
	} else if m.OutputFile != "" {
		if fout, err = os.Create(m.OutputFile); err != nil {
			return
		}
	}

	if m.UseStdErr {
		ferr = os.Stderr
	} else if m.ErrorFile != "" {
		if ferr, err = os.Create(m.ErrorFile); err != nil {
			return
		}
	}

	proc, err := server.StartProcess(args[0], m.Description, bus, args[1:]...)
	if err != nil {
		return
	}
	defer proc.Cleanup()

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
