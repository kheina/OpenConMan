package test

import (
	"fmt"

	"github.com/kheina/openconman/src/cli"
	"github.com/kheina/openconman/src/errors"
)

type Command struct {
	cli.UnimplementedCommand
}

type multierr struct {
	msg  string
	errs []error
	Dump map[string]string
	A    int
}

func (e *multierr) Error() string {
	return e.msg
}

func (e *multierr) Unwrap() []error {
	return e.errs
}

func (c *Command) Run() error {
	err := &multierr{
		msg: "test3",
		errs: []error{
			&multierr{
				msg: "test5",
				errs: []error{
					errors.New(123, "op6", "test6"),
					errors.New(123, "op7", "test7"),
				},
			},
			errors.New(123, "op1", "test1"),
			errors.New(123, "op2", "test2"),
		},
	}

	fmt.Println(errors.Tree(errors.Wrap("op4", err, "test4")))

	return nil
}

func RegisterCommand(i *cli.CLI) {
	i.NewCommand("test", func() (cli.Command, error) {
		return &Command{}, nil
	})
}
