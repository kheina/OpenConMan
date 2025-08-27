package systemd

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/hashicorp/go-hclog"

	"github.com/kheina/openconman/src/errors"
	pb "github.com/kheina/openconman/src/gen/pbs/api/systemd"
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

func (s *Server) ListServices(ctx context.Context, req *srv.GetServiceStatusesRequest) (*srv.GetServiceStatusesResponse, error) {
	const op = "systemd.(Server).ListServices"
	files, err := s.dbus.ListUnitFilesByPatternsContext(ctx, allUnitFileStateStrings(), []string{"ocm-*"})
	if err != nil {
		return nil, errors.Wrap(op, err, "failed to list systemd unit files")
	}
	units, err := s.dbus.ListUnitsByPatternsContext(ctx, allServiceUnitStateStrings(), []string{"ocm-*"})
	if err != nil {
		return nil, errors.Wrap(op, err, "failed to list systemd active units")
	}
	names := map[string]*pb.UnitStatus{}
	items := []*pb.UnitStatus{}
	for _, f := range files {
		i := strings.LastIndexByte(f.Path, '/')
		if i >= 0 {
			i := &pb.UnitStatus{
				Name:      f.Path[i+1:],
				LoadState: string(disabled),
				JobType:   f.Type,
				JobPath:   f.Path,
			}
			names[i.Name] = i
			items = append(items, i)
		} else {
			// eh?
		}
	}
	for _, u := range units {
		if _, ok := names[u.Name]; ok {
			*(names[u.Name]) = pb.UnitStatus{
				Name:        u.Name,
				Description: u.Description,
				LoadState:   u.LoadState,
				ActiveState: u.ActiveState,
				SubState:    u.SubState,
				Followed:    u.Followed,
				Path:        string(u.Path),
				JobId:       u.JobId,
				JobType:     u.JobType,
				JobPath:     string(u.JobPath),
			}
		} else {
			items = append(items, &pb.UnitStatus{
				Name:        u.Name,
				Description: u.Description,
				LoadState:   u.LoadState,
				ActiveState: u.ActiveState,
				SubState:    u.SubState,
				Followed:    u.Followed,
				Path:        string(u.Path),
				JobId:       u.JobId,
				JobType:     u.JobType,
				JobPath:     string(u.JobPath),
			})
		}
	}
	return &srv.GetServiceStatusesResponse{
		Items: items,
	}, nil
	// return s.dbus.ListUnitsByNamesContext(ctx, unitNames)
}

func (s *Server) EnableService(ctx context.Context, req *srv.GetEnableServiceRequest) (*srv.GetEnableServiceResponse, error) {
	const op = "systemd.(Server).StartService"
	s.logger.Debug("init", "op", op, "server", s)
	switch {
	case req.Name == "":
		return nil, errors.New(400, op, "missing required field: name")
	}
	s.logger.Debug("enable", "op", op, "dbus", s.dbus)
	if _, _, err := s.dbus.EnableUnitFilesContext(ctx, []string{req.Name}, false, false); err != nil {
		return nil, errors.Wrap(op, err, "failed to enable %s", req.Name)
	}
	s.logger.Debug("start", "op", op, "dbus", s.dbus)
	if _, err := s.dbus.StartUnitContext(ctx, req.Name, "replace", nil); err != nil {
		return nil, errors.Wrap(op, err, "failed to start %s", req.Name)
	}
	s.logger.Debug("list", "op", op, "dbus", s.dbus)
	units, err := s.dbus.ListUnitsByNamesContext(ctx, []string{req.Name})
	switch {
	case err != nil:
		return nil, errors.Wrap(op, err, "failed to fetch unit: %s", req.Name)
	case len(units) != 1:
		return nil, errors.New(500, op, fmt.Sprintf("failed to fetch unit: %s", req.Name))
	}
	return &srv.GetEnableServiceResponse{
		Item: &pb.UnitStatus{
			Name:        units[0].Name,
			Description: units[0].Description,
			LoadState:   units[0].LoadState,
			ActiveState: units[0].ActiveState,
			SubState:    units[0].SubState,
			Followed:    units[0].Followed,
			Path:        string(units[0].Path),
			JobId:       units[0].JobId,
			JobType:     units[0].JobType,
			JobPath:     string(units[0].JobPath),
		},
	}, nil
}

func (s *Server) DisableService(ctx context.Context, req *srv.GetEnableServiceRequest) (*srv.GetEnableServiceResponse, error) {
	const op = "systemd.(Server).StopService"
	s.logger.Debug("init", "op", op, "server", s)
	switch {
	case req.Name == "":
		return nil, errors.New(400, op, "missing required field: name")
	}
	s.logger.Debug("stop", "op", op, "dbus", s.dbus)
	if _, err := s.dbus.StopUnitContext(ctx, req.Name, "fail", nil); err != nil {
		return nil, errors.Wrap(op, err, "failed to stop %s", req.Name)
	}
	s.logger.Debug("disable", "op", op, "dbus", s.dbus)
	if _, err := s.dbus.DisableUnitFilesContext(ctx, []string{req.Name}, false); err != nil {
		return nil, errors.Wrap(op, err, "failed to disable %s", req.Name)
	}
	s.logger.Debug("list", "op", op, "dbus", s.dbus)
	units, err := s.dbus.ListUnitsByNamesContext(ctx, []string{req.Name})
	switch {
	case err != nil:
		return nil, errors.Wrap(op, err, "failed to fetch unit: %s", req.Name)
	case len(units) != 1:
		return nil, errors.New(500, op, fmt.Sprintf("failed to fetch unit: %s", req.Name))
	}
	return &srv.GetEnableServiceResponse{
		Item: &pb.UnitStatus{
			Name:        units[0].Name,
			Description: units[0].Description,
			LoadState:   units[0].LoadState,
			ActiveState: units[0].ActiveState,
			SubState:    units[0].SubState,
			Followed:    units[0].Followed,
			Path:        string(units[0].Path),
			JobId:       units[0].JobId,
			JobType:     units[0].JobType,
			JobPath:     string(units[0].JobPath),
		},
	}, nil
}
