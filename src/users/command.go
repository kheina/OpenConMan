package users

import (
	"github.com/kheina/openconman/src/cli"
)

type Command struct {
	cli.UnimplementedCommand

	sub subcommand
}

type subcommand uint8

const (
	subcommandUnknown subcommand = iota
	subcommandCreate
	subcommandRead
	subcommandUpdate
	subcommandDelete
	subcommandList
)

func (c *Command) Synopsis() string {
	switch c.sub {
	case subcommandCreate:
		return "Creates a systemctl unit file and adds it to your system. Requires root permissions."
	default:
		return "Create or modify conman users"
	}
}

func (c *Command) Description() string {
	switch c.sub {
	case subcommandUnknown:
		return "Create, modify, delete or otherwise interact with conman users. Useful for initial setup or creating additional users without the webui."
	default:
		return ""
	}
}

func RegisterCommand(i *cli.CLI) {
	c := i.NewCommand("users", func() (cli.Command, error) {
		return &Command{}, nil
	})
	c.NewSubCommand("create", func() (cli.Command, error) {
		return newCreateCommand(), nil
	})
	c.NewSubCommand("read", func() (cli.Command, error) {
		return newReadCommand(), nil
	})
	c.NewSubCommand("login", func() (cli.Command, error) {
		return newLoginCommand(), nil
	})
	// c.NewSubCommand("update", func() (cli.Command, error) {
	// 	return &Command{sub: subcommandUpdate}, nil
	// })
	// c.NewSubCommand("delete", func() (cli.Command, error) {
	// 	return &Command{sub: subcommandDelete}, nil
	// })
	// c.NewSubCommand("list", func() (cli.Command, error) {
	// 	return &Command{sub: subcommandList}, nil
	// })
}
