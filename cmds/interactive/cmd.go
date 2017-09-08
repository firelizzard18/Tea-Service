package cmd

import (
	"errors"

	"github.com/urfave/cli"

	"github.com/firelizzard18/Tea-Service/cmds/_"
)

func init() {
	cmds.Register("interactive", "open up an interactive session with a service").Executor(&Command{})
}

type Command struct {
}

func (m *Command) Execute(c *cli.Context) error {
	return errors.New("not implemented")
}
