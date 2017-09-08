package cmds

import (
	"github.com/firelizzard18/go-cli-cmds"
	"github.com/urfave/cli"
)

var commands = subcmds.NewCollection()

func Commands() []cli.Command {
	return commands.Commands()
}

func RegisterCommand(cmd cli.Command) {
	commands.RegisterCommand(cmd)
}

func Register(name, usage string, flags ...cli.Flag) subcmds.RegistrationContext {
	return commands.Register(name, usage, flags...)
}
