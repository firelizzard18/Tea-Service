package cmd

import (
	"errors"

	"github.com/urfave/cli"

	"github.com/firelizzard18/Tea-Service/client"
	"github.com/firelizzard18/Tea-Service/cmds/_"
)

func init() {
	cmds.Register("connect", "connect to a service").Executor(&Command{Command: client.CmdDefaults})
}

type Command struct {
	client.Command
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

	out, err := cl.RequestOutput(args[0], m.GetOutType(), !m.Direct, m.Timeout)
	if err != nil {
		return err
	}
	defer out.Close()

	client.DumpOutput(out)
	return nil
}
