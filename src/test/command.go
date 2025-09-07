package test

import "github.com/kheina/openconman/src/cli"

type Command struct {
	cli.UnimplementedCommand
}

func (c *Command) Run() error {
	return nil
}

func RegisterCommand(i *cli.CLI) {
	i.NewCommand("test", func() (cli.Command, error) {
		return &Command{}, nil
	})
}
