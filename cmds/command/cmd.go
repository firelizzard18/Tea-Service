package cmd

import (
	"errors"

	"github.com/urfave/cli"

	"github.com/firelizzard18/Tea-Service/client"
	"github.com/firelizzard18/Tea-Service/cmds/_"
)

func init() {
	cmds.Register("command", "send a command to a service").Executor(&Command{Command: client.CmdDefaults})
}

type Command struct {
	client.Command
	SendCmd string `long:"send" description:"send a {command} to a server and connect to it's output"`
}

func (m *Command) Execute(c *cli.Context) error {
	args := c.Args()
	if len(args) != 1 {
		return errors.New("Expected a single argument of the service to connect to")
	}

	cl, err := client.ConnectToDBus(c.GlobalString("dbus-bus"))
	if err != nil {
		return err
	}
	defer cl.Close()

	if m.SendCmd != "" {
		out, err := cl.SendCommand(args[0], m.GetOutType(), m.SendCmd)
		if err != nil {
			return err
		}

		client.DumpOutput(out)
	} else {
		in, out, err := cl.RequestCommand(args[0], m.GetOutType())
		if err != nil {
			return nil
		}

		go client.SendInput(in)
		client.DumpOutput(out)
	}

	return nil
}
