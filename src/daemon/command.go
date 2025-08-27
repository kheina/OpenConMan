package daemon

import (
	"context"
	"fmt"
	"slices"

	"github.com/kheina/openconman/src/errors"
)

type Command struct{}

func (c *Command) Help(args []string) string {
	const help = `%s

	$ conman daemon create

Available Subcommands:

	create  - Creates a systemctl unit file and adds it to your
	          system. Requires root permissions.

	delete  - Disables and removes the conman systemctl server
	          daemon, should it exist. Requires root permissions.

	enable  - Enables and starts the conman systemctl server
	          daemon. Must be run after create to start the
	          server. Requires root permissions.

	disable - Stops and prevents the conman systemctl server
	          daemon from running without being enabled again.
	          Requires root permissions.

	restart - Restarts the conman systemctl daemon. Requires
	          root permissions.

	reload  - Equivalent to running systemctl daemon-reload.
	          Requires root permissions.`

	if len(args) == 0 || slices.Contains([]string{"-h", "--help"}, args[0]) {
		return fmt.Sprintf(help, "Usage: conman daemon <subcommand> [options]")
	}

	switch args[0] {
	case "create":
		return `Usage: conman daemon create [options]

	$ conman daemon create --port 443 --host 0.0.0.0

Creates a conman server daemon systemctl unit file and adds it
to your system. Passed options are copied directly to conman
serve within the unit file. Requires root permissions.`

	case "delete":
		return `Usage: conman daemon delete

Disables and removes the conman systemctl server daemon, if it
exists. Requires root permissions.`

	case "enable":
		return `Usage: conman daemon enable

Enables the conman systemctl daemon, if it exists. This command
must be run in order to start the conman daemon after creating
it. This command allows the conman daemon to start automatically
on system restarts and reboots. Requires root permissions.`

	case "disable":
		return `Usage: conman daemon disable

Stops and disables the conman systemctl daemon, but does not
delete it. Requires root permissions.`

	case "restart":
		return `Usage: conman daemon restart

Restarts the conman systemctl daemon, useful for updating the
daemon. Requires root permissions.`

	case "reload":
		return `Usage: conman daemon reload

Reloads the systemctl daemon, useful for development, not really
necessary for every day use. Requires root permissions.`

	default:
		return fmt.Sprintf(help, fmt.Sprintf("Unknown subcommand: %s", args[0]))
	}
}

func (c *Command) Description() string {
	return "Add or modify the conman systemctl daemon"
}

func (c *Command) Run(args []string) error {
	const op = "daemon.Run"
	ctx := context.Background()
	if len(args) == 0 {
		fmt.Printf("%s\n\n", c.Help(args))
		return nil
	}
	con, err := New(ctx)
	if err != nil {
		return errors.Wrap(op, err, "failed to initialize daemon controller")
	}
	switch args[0] {
	case "create":
		return con.Create(ctx, args[1:])
	case "enable":
		return con.Enable(ctx)
	case "disable":
		return con.Disable(ctx)
	case "delete":
		return con.Delete(ctx)
	case "restart":
		return con.Restart(ctx)
	case "reload":
		return con.Reload(ctx)
	default:
		fmt.Printf("%s\n\n", c.Help(args))
		return nil
	}
}
