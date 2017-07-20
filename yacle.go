package main

import (
	"log"
	"os"

	"github.com/otiai10/yacle/commands"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "yacle"
	app.Commands = []cli.Command{
		commands.Run,
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
