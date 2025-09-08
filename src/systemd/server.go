package systemd

import (
	"context"
	"runtime"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/hashicorp/go-hclog"

	"github.com/kheina/openconman/src/errors"
	srv "github.com/kheina/openconman/src/gen/srv/api/systemd"
)

type Server struct {
	srv.UnimplementedSystemdServer

	dbus   *dbus.Conn
	logger hclog.Logger
}

func New(ctx context.Context, opt ...Option) (*Server, error) {
	const op = "daemon.New"
	opts := getOpts(opt...)
	dconn, err := dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		return nil, errors.Wrap(op, err, "could not connect to systemd bus")
	}
	c := &Server{
		dbus:   dconn,
		logger: opts.withLogger,
	}
	runtime.AddCleanup(c, func(dc *dbus.Conn) { dc.Close() }, c.dbus)
	return c, nil
}
