package commands

import (
	"fmt"
	"os"

	cwl "github.com/otiai10/cwl.go"
	"github.com/otiai10/yacle/core"
	"github.com/urfave/cli"
)

// Run ...
var Run = cli.Command{
	Name:        "run",
	Aliases:     []string{"r"},
	Usage:       "Run your workflow described by CWL",
	ArgsUsage:   "[file.cwl] [inputs.yaml]",
	Description: "Run the cwl",
	Action: func(ctx *cli.Context) error {

		r, err := os.Open(ctx.Args().First())
		if err != nil {
			return fmt.Errorf("failed to open CWL: %v", err)
		}

		root := cwl.NewCWL()
		if err = root.Decode(r); err != nil {
			return err
		}
		r.Close()

		// p, err := filepath.Abs(r.Name())
		// if err != nil {
		// 	return err
		// }

		if len(ctx.Args()) > 1 {
			r, err = os.Open(ctx.Args()[1])
			if err != nil {
				return err
			}
			r.Close()
		}

		if err = core.Run(root); err != nil {
			return err
		}

		return nil

	},
}
