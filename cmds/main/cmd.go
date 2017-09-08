package cmd

import (
	"errors"

	"github.com/firelizzard18/go-cli-cmds"
	"github.com/urfave/cli"
)

func SetMain(app *cli.App) {
	cmds := subcmds.NewCollection()
	cmds.Register("main", "main").Executor(&Command{})

	main := cmds.Commands()[0]
	// app.Action = main.Action
	app.Flags = main.Flags
}

type Command struct {
	Bus string `short:"b" long:"bus,dbus-bus" description:"Configure the DBus {session} to use"`
}

func (m *Command) Execute(c *cli.Context) error {
	return errors.New("not implemented")
}
