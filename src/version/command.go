package version

import (
	"fmt"

	"github.com/kheina/openconman/src/cli"
)

type Command struct {
	cli.UnimplementedCommand

	raw bool
}

func (c *Command) Synopsis() string {
	return "Prints the current version info"
}

const timeFormat = "2006-01-02T15:04:05Z"

func (c *Command) Args() cli.Args {
	return cli.Args{
		cli.NewBoolArg("raw", "returns only the version number in the format v0.0.0+dev", &c.raw, "--raw", "-r"),
	}
}

func (c *Command) Run() error {
	v, err := New()
	if err != nil {
		return err
	}
	if c.raw {
		_, err = fmt.Printf("v%s", v)
	} else {
		_, err = fmt.Printf("version number:    %s\ncommit hash:       %s\nbuild date:        %s\n\n", v, v.Rev, v.BuildDate.UTC().Format(timeFormat))
	}
	return err
}

func RegisterCommand(i *cli.CLI) {
	i.NewCommand("version", func() (cli.Command, error) {
		return &Command{}, nil
	})
}
