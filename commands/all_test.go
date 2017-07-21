package commands

import (
	"flag"
	"testing"

	. "github.com/otiai10/mint"
	"github.com/urfave/cli"
)

func TestRun(t *testing.T) {
	set := flag.NewFlagSet("yacle", flag.ExitOnError)
	set.Parse([]string{
		"../testdata/001/1st-tool.cwl.yaml",
		"../testdata/001/echo-job.yml",
	})
	ctx := cli.NewContext(cli.NewApp(), set, nil)
	err := Run.Action.(func(*cli.Context) error)(ctx)
	Expect(t, err).ToBe(nil)

	When(t, "CWL file is not provided", func(t *testing.T) {
		set := flag.NewFlagSet("yacle", flag.ExitOnError)
		set.Parse([]string{"../testdata/not/existing"})
		ctx := cli.NewContext(cli.NewApp(), set, nil)
		err := Run.Action.(func(*cli.Context) error)(ctx)
		Expect(t, err).Not().ToBe(nil)
		Expect(t, err.Error()).ToBe("failed to open CWL: open ../testdata/not/existing: no such file or directory")
	})

	When(t, "Required params are not provided", func(t *testing.T) {
		set := flag.NewFlagSet("yacle", flag.ExitOnError)
		set.Parse([]string{"../testdata/001/1st-tool.cwl.yaml"})
		ctx := cli.NewContext(cli.NewApp(), set, nil)
		err := Run.Action.(func(*cli.Context) error)(ctx)
		Expect(t, err).Not().ToBe(nil)
		Expect(t, err.Error()).ToBe("Input `message` is required but not provided")
	})
}
