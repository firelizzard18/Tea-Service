package cmd

import (
	"errors"

	"github.com/urfave/cli"

	"github.com/firelizzard18/Tea-Service/client"
	"github.com/firelizzard18/Tea-Service/cmds/_"
)

func init() {
	cmds.Register("send", "send a command to a service").Executor(&Command{Command: client.CmdDefaults})
}

type Command struct {
	client.Command
}

func (m *Command) Execute(c *cli.Context) error {
	args := c.Args()
	if len(args) != 2 {
		return errors.New("Expected two arguments of the service to connect to and the command to send")
	}

	cl, err := client.ConnectToDBus(c.GlobalString("dbus-bus"))
	if err != nil {
		return err
	}
	defer cl.Close()

	out, err := cl.SendCommand(args[0], m.GetOutType(), args[1], !m.Direct, m.Timeout)
	if err != nil {
		return err
	}
	defer out.Close()

	return nil
}
