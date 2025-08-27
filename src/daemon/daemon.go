package daemon

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/coreos/go-systemd/v22/dbus"

	"github.com/kheina/openconman/src/errors"
)

type controller struct {
	dbus *dbus.Conn
}

func New(ctx context.Context) (*controller, error) {
	const op = "daemon.New"
	dconn, err := dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		return nil, errors.Wrap(op, err, "could not connect to systemd bus")
	}
	c := &controller{
		dbus: dconn,
	}
	runtime.AddCleanup(c, func(dc *dbus.Conn) { dc.Close() }, c.dbus)
	return c, nil
}

const service = "openconman.service"

func (c *controller) Create(ctx context.Context, args []string) error {
	const op = "daemon.(controller).Create"
	exe, err := os.Executable()
	if err != nil {
		return errors.Wrap(op, err, "failed to create openconman.service file")
	}
	if exe == "" || strings.HasPrefix(exe, "/tmp") {
		return fmt.Errorf("%s: cannot set a temporary file as daemon executable. exec: %s", op, exe)
	}
	if err = newUnitFile(exe, service, args); err != nil {
		return errors.Wrap(op, err, "failed to create openconman.service file")
	}
	if err = c.dbus.ReloadContext(ctx); err != nil {
		return errors.Wrap(op, err, "failed to reload systemctl context")
	}
	if c.dbus, err = dbus.NewSystemdConnectionContext(ctx); err != nil {
		return errors.Wrap(op, err, "could not reconnect to systemd bus")
	}
	return nil
}

func (c *controller) Enable(ctx context.Context) error {
	const op = "daemon.(controller).Enable"
	if _, _, err := c.dbus.EnableUnitFilesContext(ctx, []string{service}, false, false); err != nil {
		return errors.Wrap(op, err, "failed to enable %s", service)
	}
	if _, err := c.dbus.StartUnitContext(ctx, service, "replace", nil); err != nil {
		return errors.Wrap(op, err, "failed to start %s", service)
	}
	return nil
}

func (c *controller) Disable(ctx context.Context) error {
	const op = "daemon.(controller).Disable"
	if _, err := c.dbus.StopUnitContext(ctx, service, "fail", nil); err != nil {
		return errors.Wrap(op, err, "failed to start %s", service)
	}
	if _, err := c.dbus.DisableUnitFilesContext(ctx, []string{service}, false); err != nil {
		return errors.Wrap(op, err, "failed to disable %s", service)
	}
	return nil
}

func (c *controller) Delete(ctx context.Context) error {
	const op = "daemon.(controller).Delete"
	if err := c.Disable(ctx); err != nil {
		return err
	}
	if err := deleteUnitFile(service); err != nil {
		return errors.Wrap(op, err, "failed to delete openconman.service file")
	}
	return nil
}

func (c *controller) Restart(ctx context.Context) error {
	const op = "daemon.(controller).Restart"
	var err error
	if err = c.dbus.ReloadContext(ctx); err != nil {
		return errors.Wrap(op, err, "failed to reload systemctl context")
	}
	if c.dbus, err = dbus.NewSystemdConnectionContext(ctx); err != nil {
		return errors.Wrap(op, err, "could not reconnect to systemd bus")
	}
	if _, err = c.dbus.RestartUnitContext(ctx, service, "replace", nil); err != nil {
		return errors.Wrap(op, err, "failed to restart conman daemon")
	}
	return nil
}

func (c *controller) Reload(ctx context.Context) error {
	const op = "daemon.(controller).Restart"
	var err error
	if err = c.dbus.ReloadContext(ctx); err != nil {
		return errors.Wrap(op, err, "failed to reload systemctl context")
	}
	if c.dbus, err = dbus.NewSystemdConnectionContext(ctx); err != nil {
		return errors.Wrap(op, err, "could not reconnect to systemd bus")
	}
	return nil
}
