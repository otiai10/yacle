package main

import (
	"log"
	"os"

	"github.com/otiai10/yacle/commands"
	"github.com/urfave/cli"
)

const version = "0.0.1"

func main() {
	app := cli.NewApp()
	app.Name = "yacle"
	app.Usage = "Yet Another CWL Engine"
	app.Version = version
	app.Commands = []cli.Command{
		commands.Run, // yacle run
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "outdir",
			Usage: "Output Directory",
		},
		cli.BoolFlag{
			Name:  "quiet",
			Usage: "compress stdout log",
		},
	}
	app.Action = commands.Run.Action // yacle (without subcommand)
	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
