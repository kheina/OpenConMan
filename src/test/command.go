package test

import (
	"context"
	"fmt"

	"github.com/kheina/openconman/src/errors"
	"github.com/kheina/openconman/src/systemd"
)

type Command struct{}

func (c *Command) Help(args []string) string {
	return "does test things"
}

func (c *Command) Description() string {
	return "does test things"
}

func (c *Command) Run(args []string) error {
	const op = "test.Run"
	ctx := context.Background()
	srv, err := systemd.New(ctx)
	if err != nil {
		return errors.Wrap(op, err, "failed to create systemd server")
	}
	list, err := srv.ListServices(ctx, nil)
	if err != nil {
		return errors.Wrap(op, err, "failed to list systemd services")
	}
	fmt.Println("srv.List:", list.Items)
	return nil
}
