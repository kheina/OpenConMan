package users

import (
	"fmt"
	"os"
	"time"

	"github.com/kheina/openconman/src/auth"
	"github.com/kheina/openconman/src/cli"
)

type loginCommand struct {
	cli.UnimplementedCommand

	name     string
	password []byte
	workDir  string
}

func newLoginCommand() *loginCommand {
	return &loginCommand{}
}

func (c *loginCommand) Synopsis() string {
	return "Read a userfile for syntax and parsing verification"
}

func (c *loginCommand) Args() cli.Args {
	return cli.Args{
		cli.NewStringArg("user name", "The user to login as", &c.name, "--name", "-n"),
		cli.NewByteStringArg("password", "The user's password for authentication", &c.password, "--password", "-p"),
		cli.NewStringArg("work dir", "Retrieves the user from the provided working directory, in a format expected by a conman server running in the same directory.", &c.workDir, "--work-dir", "-w"),
	}
}

func (c *loginCommand) Run() error {
	const op = "users.(loginCommand).Run"
	switch {
	case c.name == "":
		return fmt.Errorf("%s: user name missing", op)
	case c.password == nil:
		return fmt.Errorf("%s: user password missing", op)
	case c.workDir == "":
		return fmt.Errorf("%s: working directory missing", op)
	}
	if err := os.Chdir(c.workDir); err != nil {
		return fmt.Errorf("%s: failed to set working directory", op)
	}
	user, err := auth.Login(c.name, c.password)
	if err != nil {
		return fmt.Errorf("%s: login failed: %w", op, err)
	}
	tok, err := auth.NewAuthToken(user, time.Now().Add(time.Hour*24*365)) // 1 year ig
	if err != nil {
		return fmt.Errorf("%s: failed to create token: %w", op, err)
	}
	fmt.Println(string(tok))
	return nil
}
