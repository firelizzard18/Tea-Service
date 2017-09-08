package cmd

import (
	"log"

	"github.com/urfave/cli"

	"github.com/firelizzard18/Tea-Service/client"
	"github.com/firelizzard18/Tea-Service/cmds/_"
)

func init() {
	cmds.Register("list", "list available services").Executor(&Command{Timeout: 100})
}

type Command struct {
	Timeout int `short:"t" long:"timeout" description:"timeout for listing available servers"`
}

func (m *Command) Execute(c *cli.Context) error {
	cl, err := client.ConnectToDBus(c.GlobalString("dbus-bus"))
	if err != nil {
		return err
	}

	for server := range cl.ListServers(m.Timeout) {
		log.Print(server)
	}

	return nil
}
