package commands

import (
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	"github.com/otiai10/gulo/gulo"
	"github.com/urfave/cli"
)

// Run ...
var Run = cli.Command{
	Name:        "run",
	Description: "Run the cwl",
	Action: func(ctx *cli.Context) error {

		r, err := os.Open(ctx.Args().First())
		if err != nil {
			return err
		}

		b, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		r.Close()

		root := gulo.NewCWL()
		if err = yaml.Unmarshal(b, root); err != nil {
			return err
		}

		p, err := filepath.Abs(r.Name())
		if err != nil {
			return err
		}
		root.Path = p

		if len(ctx.Args()) > 1 {
			r, err = os.Open(ctx.Args()[1])
			if err != nil {
				return err
			}
			if err = gulo.ParseProvidedInputs(r, root.ProvidedInputs); err != nil {
				return err
			}
			r.Close()
		}

		if err = root.Run(); err != nil {
			return err
		}

		return nil

	},
}
