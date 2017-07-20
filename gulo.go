package main

import (
	"log"
	"os"

	"github.com/otiai10/gulo/commands"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "gulo"
	app.Commands = []cli.Command{
		commands.Run,
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
