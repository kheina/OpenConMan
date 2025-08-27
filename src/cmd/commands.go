package cmd

import (
	"github.com/kheina/openconman/src/daemon"
	"github.com/kheina/openconman/src/srv"
	"github.com/kheina/openconman/src/test"
	"github.com/kheina/openconman/src/version"
)

// TODO: prolly switch to using "github.com/mitchellh/cli" and "github.com/posener/complete"

type Command interface {
	// Description returns a short one-line descriptor of the command
	Description() string
	// Help returns a help message based upon the passed sub-command args
	Help(args []string) string
	// Run executes the given command based upon the passed sub-command args
	Run(args []string) error
}

type CommandFactory func() (Command, error)

var Commands map[string]CommandFactory

func initCommands() {
	// try to add new commands in a loosely alphabetic order
	Commands = map[string]CommandFactory{
		"daemon": func() (Command, error) {
			return &daemon.Command{}, nil
		},
		"serve": func() (Command, error) {
			return &srv.Command{}, nil
		},
		"version": func() (Command, error) {
			return &version.Command{}, nil
		},

		"test": func() (Command, error) {
			return &test.Command{}, nil
		},
	}
}
