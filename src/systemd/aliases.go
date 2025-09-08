package systemd

import (
	"context"
	"fmt"
	"os"

	"github.com/kheina/openconman/src/errors"
	pb "github.com/kheina/openconman/src/gen/pbs/api/systemd"
	srv "github.com/kheina/openconman/src/gen/srv/api/systemd"
)

func (s *Server) PutServiceAlias(ctx context.Context, req *srv.PutServiceAliasRequest) (*srv.PutServiceAliasResponse, error) {
	const op = "systemd.(Server).PutServiceAlias"
	switch {
	case req.Name == "":
		return nil, errors.New(errors.BadRequest, op, "missing required field: name")
	}

	units, err := s.dbus.ListUnitFilesByPatternsContext(ctx, allUnitFileStateStrings(), []string{req.Name})
	switch {
	case err != nil:
		return nil, errors.Wrap(op, err, "failed to list units")
	case len(units) != 1:
		return nil, errors.New(errors.Internal, op, fmt.Sprintf("incorrect number of units found. expected 1, found %d", len(units)))
	}

	symlink, err := newAliasPath(&units[0])
	if err != nil {
		return nil, errors.Wrap(op, err, "failed to create symlink path")
	}
	if err = os.Symlink(units[0].Path, symlink); err != nil {
		return nil, errors.Wrap(op, err, "failed to create symlink alias")
	}
	if err = s.dbus.ReloadContext(ctx); err != nil {
		return nil, errors.Wrap(op, err, "failed to reload daemon")
	}

	aliases, err := s.dbus.ListUnitsByNamesContext(ctx, []string{pathToName(symlink)})
	switch {
	case err != nil:
		return nil, errors.Wrap(op, err, "failed to list units from alias")
	case len(aliases) != 1:
		// what
		return nil, errors.New(errors.Internal, op, fmt.Sprintf("incorrect number of aliases found. expected 1, found %d", len(units)))
	}

	return &srv.PutServiceAliasResponse{
		Item: &pb.UnitStatus{
			Name:        aliases[0].Name,
			Description: aliases[0].Description,
			LoadState:   aliases[0].LoadState,
			ActiveState: aliases[0].ActiveState,
			SubState:    aliases[0].SubState,
			Followed:    aliases[0].Followed,
			Path:        string(aliases[0].Path),
			JobId:       aliases[0].JobId,
			JobType:     aliases[0].JobType,
			JobPath:     string(aliases[0].JobPath),
		},
	}, nil
}

func (s *Server) DeleteServiceAlias(ctx context.Context, req *srv.DeleteServiceRequest) (*srv.DeleteServiceResponse, error) {
	const op = "systemd.(Server).DeleteServiceAlias"
	switch {
	case req.Name == "":
		return nil, errors.New(errors.BadRequest, op, "missing required field: name")
	}

	units, err := s.dbus.ListUnitFilesByPatternsContext(ctx, allUnitFileStateStrings(), []string{req.Name})
	switch {
	case err != nil:
		return nil, errors.Wrap(op, err, "failed to list units")
	case len(units) != 1:
		return nil, errors.New(errors.Internal, op, fmt.Sprintf("incorrect number of units found. expected 1, found %d", len(units)))
	}

	if units[0].Type != string(alias) {
		return nil, errors.New(errors.Internal, op, "unit name is not an alias")
	}
	if path := parseAliasPath(units[0].Path); path == "" {
		return nil, errors.New(errors.Internal, op, "unit name is not an alias")
	}

	if err = os.Remove(units[0].Path); err != nil {
		return nil, errors.Wrap(op, err, "failed to remove alias symlink")
	}
	if err = s.dbus.ReloadContext(ctx); err != nil {
		return nil, errors.Wrap(op, err, "failed to reload daemon")
	}

	return nil, nil
}
