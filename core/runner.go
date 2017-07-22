package core

import (
	"os/exec"

	cwl "github.com/otiai10/cwl.go"
)

// Run TODO: Refactor
func Run(root *cwl.Root) error {
	args, err := root.Args()
	if err != nil {
		return err
	}
	cmd := exec.Command(root.BaseCommand, args...)
	return cmd.Run()
}
