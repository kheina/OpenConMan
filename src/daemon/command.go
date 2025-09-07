package daemon

import (
	"context"

	"github.com/kheina/openconman/src/cli"
	"github.com/kheina/openconman/src/errors"
)

type Command struct {
	cli.UnimplementedCommand

	key    daemonCommandKey
	config string
}

type daemonCommandKey uint8

const (
	daemonCommandUnknown daemonCommandKey = iota
	daemonCommandCreate
	daemonCommandDelete
	daemonCommandEnable
	daemonCommandDisable
	daemonCommandRestart
	daemonCommandReload
)

func (c *Command) Synopsis() string {
	switch c.key {
	case daemonCommandCreate:
		return "Creates a systemctl unit file and adds it to your system. Requires root permissions."
	case daemonCommandDelete:
		return "Disables and removes the conman systemctl server daemon, should it exist. Requires root permissions."
	case daemonCommandEnable:
		return "Enables and starts the conman systemctl server daemon. Must be run after create to start the server. Requires root permissions."
	case daemonCommandDisable:
		return "Stops and prevents the conman systemctl server daemon from running without being enabled again. Requires root permissions."
	case daemonCommandRestart:
		return "Restarts the conman systemctl daemon. Requires root permissions."
	case daemonCommandReload:
		return "Equivalent to running systemctl daemon-reload. Requires root permissions."
	default:
		return "Add or modify the conman systemctl daemon"
	}
}

func (c *Command) Description() string {
	switch c.key {
	case daemonCommandCreate:
		return "Creates a conman server daemon systemctl unit file and adds it to your system. Server config file will be copied to /etc/conman.d, the working directory of the systemd daemon. Requires root permissions."
	case daemonCommandDelete:
		return "Disables and removes the conman systemctl server daemon, if it exists. Requires root permissions."
	case daemonCommandEnable:
		return "Enables the conman systemctl daemon, if it exists. This command must be run in order to start the conman daemon after creating it. This command allows the conman daemon to start automatically on system restarts and reboots. Requires root permissions."
	case daemonCommandDisable:
		return "Stops and disables the conman systemctl daemon, but does not delete it. Requires root permissions."
	case daemonCommandRestart:
		return "Restarts the conman systemctl daemon, useful for updating the daemon. Requires root permissions."
	case daemonCommandReload:
		return "Reloads the systemctl daemon, useful for development, not really necessary for every day use. Requires root permissions."
	default:
		return ""
	}
}

func (c *Command) Args() cli.Args {
	switch c.key {
	case daemonCommandCreate:
		return cli.Args{
			cli.NewStringArg("config", "Specifies a config file to be used for the conman server. Should be a filepath: /tmp/example.cfg", &c.config, "--config-file", "-c"),
		}
	default:
		return cli.Args{}
	}
}

func (c *Command) Run() error {
	const op = "daemon.Run"
	if c.key == daemonCommandUnknown {
		// default to standard behavior (return help)
		return cli.UnimplementedCommandError
	}
	ctx := context.Background()
	con, err := New(ctx)
	if err != nil {
		return errors.Wrap(op, err, "failed to initialize daemon controller")
	}
	switch c.key {
	case daemonCommandCreate:
		return con.Create(ctx, []string{})
	case daemonCommandEnable:
		return con.Enable(ctx)
	case daemonCommandDisable:
		return con.Disable(ctx)
	case daemonCommandDelete:
		return con.Delete(ctx)
	case daemonCommandRestart:
		return con.Restart(ctx)
	case daemonCommandReload:
		return con.Reload(ctx)
	default:
		// default to standard behavior (return help)
		return cli.UnimplementedCommandError
	}
}

func RegisterCommand(i *cli.CLI) {
	c := i.NewCommand("daemon", func() (cli.Command, error) {
		return &Command{}, nil
	})
	c.NewSubCommand("create", func() (cli.Command, error) {
		return &Command{key: daemonCommandCreate}, nil
	})
	c.NewSubCommand("enable", func() (cli.Command, error) {
		return &Command{key: daemonCommandEnable}, nil
	})
	c.NewSubCommand("disable", func() (cli.Command, error) {
		return &Command{key: daemonCommandDisable}, nil
	})
	c.NewSubCommand("delete", func() (cli.Command, error) {
		return &Command{key: daemonCommandDelete}, nil
	})
	c.NewSubCommand("restart", func() (cli.Command, error) {
		return &Command{key: daemonCommandRestart}, nil
	})
	c.NewSubCommand("reload", func() (cli.Command, error) {
		return &Command{key: daemonCommandReload}, nil
	})
}
