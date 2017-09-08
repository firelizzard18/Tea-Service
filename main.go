package main

import (
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/firelizzard18/Tea-Service/cmds"
	maincmd "github.com/firelizzard18/Tea-Service/cmds/main"
)

func main() {
	app := cli.NewApp()
	app.Name = "teasvc"
	app.Usage = "Do teasvc-y things"
	app.Version = "0.3"
	app.Authors = []cli.Author{
		{
			Name:  "Ethan Reesor",
			Email: "ethan.reesor@gmail.com",
		},
	}

	maincmd.SetMain(app)
	app.Commands = cmds.Commands()

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
