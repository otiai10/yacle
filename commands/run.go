package commands

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	Action: func(ctx *cli.Context) (err error) {

		cwlfile, err := os.Open(ctx.Args().First())
		if err != nil {
			return fmt.Errorf("failed to open CWL: %v", err)
		}
		defer cwlfile.Close()

		root := cwl.NewCWL()
		if err = root.Decode(cwlfile); err != nil {
			return fmt.Errorf("failed to decode CWL file: %v", err)
		}
		if root.Path, err = filepath.Abs(cwlfile.Name()); err != nil {
			return err
		}

		paramfile, err := os.Open(ctx.Args().Get(1))
		if err != nil {
			return fmt.Errorf("failed to open parameter file: %v", err)
		}
		defer paramfile.Close()

		params := cwl.NewParameters()
		if err = params.Decode(paramfile); err != nil {
			return fmt.Errorf("failed to decode parameter file: %v", err)
		}

		handler, err := core.NewHandler(root)
		if err != nil {
			return fmt.Errorf("failed to instantiate yacle.Handler: %v", err)
		}

		// Optionals
		handler.Outdir = ctx.String("outdir")
		handler.Quiet = ctx.Bool("quiet")

		// Debug ...
		handler.Quiet = false
		handler.SetLogger(log.New(os.Stdout, "[DEBUG]\t", 0))

		return handler.Handle(*params)

	},
}
