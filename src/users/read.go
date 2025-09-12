package users

import (
	"fmt"

	"github.com/kheina/openconman/src/cli"
	"github.com/kheina/openconman/src/users/config"
)

type readCommand struct {
	cli.UnimplementedCommand

	in []byte
}

func newReadCommand() *readCommand {
	return &readCommand{}
}

func (c *readCommand) Synopsis() string {
	return "Read a userfile for syntax and parsing verification"
}

func (c *readCommand) Args() cli.Args {
	return cli.Args{
		cli.NewByteStringArg("in file", "The user file to be read", &c.in, "--in", "-i"),
	}
}

func (c *readCommand) Run() error {
	user := &config.UserConfig{}
	user.Unmarshal(c.in)
	out, err := user.Marshal()
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}
