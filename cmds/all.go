package cmds

import (
	"github.com/urfave/cli"

	"github.com/firelizzard18/Tea-Service/cmds/_"

	// register commands
	_ "github.com/firelizzard18/Tea-Service/cmds/command"
	_ "github.com/firelizzard18/Tea-Service/cmds/connect"
	_ "github.com/firelizzard18/Tea-Service/cmds/interactive"
	_ "github.com/firelizzard18/Tea-Service/cmds/launch"
	_ "github.com/firelizzard18/Tea-Service/cmds/list"
)

func Commands() []cli.Command {
	return cmds.Commands()
}
