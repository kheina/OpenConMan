package systemd

import (
	"context"
	stderr "errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/coreos/go-systemd/v22/sdjournal"
	"google.golang.org/grpc"

	"github.com/kheina/openconman/src/auth"
	"github.com/kheina/openconman/src/errors"
	pb "github.com/kheina/openconman/src/gen/pbs/api/systemd"
	srv "github.com/kheina/openconman/src/gen/srv/api/systemd"
	"github.com/kheina/openconman/src/util"
)

func (s *Server) GetService(req *srv.GetServiceRequest, stream grpc.ServerStreamingServer[srv.GetServiceResponse]) error {
	const op = "systemd.(Server).GetService"
	ctx := stream.Context()
	if err := auth.Authorize(ctx, auth.Read, auth.Systemd); err != nil {
		return errors.Wrap(op, err, "failed to authorize request")
	}

	switch {
	case req.Name == "":
		return errors.New(errors.BadRequest, op, "missing required field: name")
	}

	for {
		units, err := s.dbus.ListUnitsByNamesContext(ctx, []string{req.Name})
		switch {
		case err != nil:
			if stderr.Is(err, context.Canceled) {
				return nil
			}
			return errors.Wrap(op, err, "failed to list units")
		case len(units) != 1:
			return errors.New(errors.Internal, op, fmt.Sprintf("incorrect number of units found. expected 1, found %d", len(units)))
		}

		stream.Send(&srv.GetServiceResponse{
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
		})

		switch {
		case !req.Live:
			return nil
		case ctx.Err() != nil:
			return nil
		}
		time.Sleep(1 * time.Second)
	}
}

func (s *Server) ListAllServices(ctx context.Context, req *srv.ListAllServiceStatusesRequest) (*srv.GetServiceStatusesResponse, error) {
	const op = "systemd.(Server).ListServices"
	if err := auth.Authorize(ctx, auth.List, auth.Systemd); err != nil {
		return nil, errors.Wrap(op, err, "failed to authorize request")
	}

	files, err := s.dbus.ListUnitFilesContext(ctx)
	if err != nil {
		return nil, errors.Wrap(op, err, "failed to list systemd unit files")
	}
	units, err := s.dbus.ListUnitsContext(ctx)
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
	sort.Slice(items, func(a, b int) bool {
		return items[a].Name < items[b].Name
	})
	return &srv.GetServiceStatusesResponse{
		Items: items,
	}, nil
}

func (s *Server) ListServices(req *srv.GetServiceStatusesRequest, stream grpc.ServerStreamingServer[srv.GetServiceStatusesResponse]) error {
	const op = "systemd.(Server).ListServices"
	ctx := stream.Context()
	if err := auth.Authorize(ctx, auth.List, auth.Systemd); err != nil {
		return errors.Wrap(op, err, "failed to authorize request")
	}

	for {
		// fetch unit files (never started, permanently stopped, aliases)
		files, err := s.dbus.ListUnitFilesByPatternsContext(ctx, allUnitFileStateStrings(), []string{"ocm-*"})
		if err != nil {
			return errors.Wrap(op, err, "failed to list systemd unit files")
		}

		// fetch active units (stopped, currently running, etc)
		units, err := s.dbus.ListUnitsByPatternsContext(ctx, allServiceUnitStateStrings(), []string{"ocm-*"})
		if err != nil {
			return errors.Wrap(op, err, "failed to list systemd active units")
		}

		// resolve aliases
		aliases, err := s.dbus.ListUnitsByNamesContext(ctx, getAliasNames(files))
		if err != nil {
			return errors.Wrap(op, err, "failed to list systemd aliased units")
		}
		units = append(units, aliases...)

		amap := aliasMap(files)
		names := map[string]*pb.UnitStatus{}
		items := []*pb.UnitStatus{}
		for _, f := range files {
			if f.Type == string(alias) {
				continue
			}
			if i := strings.LastIndexByte(f.Path, '/'); i >= 0 {
				u := &pb.UnitStatus{
					Name:      f.Path[i+1:],
					LoadState: string(disabled),
					JobType:   f.Type,
					JobPath:   f.Path,
				}
				names[u.Name] = u
				items = append(items, u)
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
					Alias:       util.OptionalString(amap[u.Name]),
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
					Alias:       util.OptionalString(amap[u.Name]),
				})
			}
		}
		sort.Slice(items, func(a, b int) bool {
			return items[a].Name < items[b].Name
		})
		stream.Send(&srv.GetServiceStatusesResponse{
			Items: items,
		})
		switch {
		case !req.Live:
			return nil
		case ctx.Err() != nil:
			return nil
		}
		time.Sleep(1 * time.Second)
	}
}

func (s *Server) EnableService(ctx context.Context, req *srv.GetEnableServiceRequest) (*srv.GetServiceResponse, error) {
	const op = "systemd.(Server).EnableService"
	if err := auth.Authorize(ctx, auth.Update, auth.Systemd); err != nil {
		return nil, errors.Wrap(op, err, "failed to authorize request")
	}

	switch {
	case req.Name == "":
		return nil, errors.New(errors.BadRequest, op, "missing required field: name")
	}

	s.logger.Debug("enable", "op", op, "dbus", s.dbus)
	if _, _, err := s.dbus.EnableUnitFilesContext(ctx, []string{req.Name}, false, false); err != nil {
		return nil, errors.Wrap(op, err, fmt.Sprintf("failed to enable %s", req.Name))
	}

	s.logger.Debug("start", "op", op, "dbus", s.dbus)
	if _, err := s.dbus.StartUnitContext(ctx, req.Name, "replace", nil); err != nil {
		return nil, errors.Wrap(op, err, fmt.Sprintf("failed to start %s", req.Name))
	}

	s.logger.Debug("list", "op", op, "dbus", s.dbus)
	units, err := s.dbus.ListUnitsByNamesContext(ctx, []string{req.Name})
	switch {
	case err != nil:
		return nil, errors.Wrap(op, err, fmt.Sprintf("failed to fetch unit: %s", req.Name))
	case len(units) != 1:
		return nil, errors.New(errors.Internal, op, fmt.Sprintf("failed to fetch unit: %s", req.Name))
	}

	return &srv.GetServiceResponse{
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

func (s *Server) DisableService(ctx context.Context, req *srv.GetEnableServiceRequest) (*srv.GetServiceResponse, error) {
	const op = "systemd.(Server).DisableService"
	if err := auth.Authorize(ctx, auth.Update, auth.Systemd); err != nil {
		return nil, errors.Wrap(op, err, "failed to authorize request")
	}

	switch {
	case req.Name == "":
		return nil, errors.New(errors.BadRequest, op, "missing required field: name")
	}

	s.logger.Debug("stop", "op", op, "dbus", s.dbus)
	if _, err := s.dbus.StopUnitContext(ctx, req.Name, "fail", nil); err != nil {
		return nil, errors.Wrap(op, err, fmt.Sprintf("failed to stop %s", req.Name))
	}

	s.logger.Debug("disable", "op", op, "dbus", s.dbus)
	if _, err := s.dbus.DisableUnitFilesContext(ctx, []string{req.Name}, false); err != nil {
		return nil, errors.Wrap(op, err, fmt.Sprintf("failed to disable %s", req.Name))
	}

	s.logger.Debug("list", "op", op, "dbus", s.dbus)
	units, err := s.dbus.ListUnitsByNamesContext(ctx, []string{req.Name})
	switch {
	case err != nil:
		return nil, errors.Wrap(op, err, fmt.Sprintf("failed to fetch unit: %s", req.Name))
	case len(units) != 1:
		return nil, errors.New(errors.Internal, op, fmt.Sprintf("failed to fetch unit: %s", req.Name))
	}

	return &srv.GetServiceResponse{
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

func (s *Server) PutService(ctx context.Context, req *srv.PutServiceRequest) (*srv.PutServiceResponse, error) {
	const op = "systemd.(Server).PutService"
	if err := auth.Authorize(ctx, auth.Create, auth.Systemd); err != nil {
		return nil, errors.Wrap(op, err, "failed to authorize request")
	}

	switch {
	case req.Name == "":
		return nil, errors.New(errors.BadRequest, op, "missing required field: name")
	case req.Content == "":
		return nil, errors.New(errors.BadRequest, op, "missing required field: content")
	}

	name := req.Name
	if !strings.HasSuffix(req.Name, ".service") {
		name += ".service"
	}
	path := name
	if !strings.ContainsRune(req.Name, '/') {
		path = "/etc/systemd/system/" + path
	}

	if util.PathExists(path) {
		return nil, errors.New(errors.Conflict, op, fmt.Sprintf("file already exists at %s", path))
	}

	if err := os.WriteFile(path, []byte(req.Content), 0644); err != nil {
		return nil, errors.Wrap(op, err, "failed to write unit file")
	}
	if err := s.dbus.ReloadContext(ctx); err != nil {
		return nil, errors.Wrap(op, err, "failed to reload daemon")
	}

	units, err := s.dbus.ListUnitsByNamesContext(ctx, []string{name})
	switch {
	case err != nil:
		return nil, errors.Wrap(op, err, "failed to list units")
	case len(units) != 1:
		return nil, errors.New(errors.Internal, op, fmt.Sprintf("incorrect number of units found. expected 1, found %d", len(units)))
	}

	return &srv.PutServiceResponse{
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

func (s *Server) DeleteService(ctx context.Context, req *srv.DeleteServiceRequest) (*srv.DeleteServiceResponse, error) {
	const op = "systemd.(Server).DeleteService"
	if err := auth.Authorize(ctx, auth.Delete, auth.Systemd); err != nil {
		return nil, errors.Wrap(op, err, "failed to authorize request")
	}

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

	if units[0].Type == string(alias) {
		return nil, errors.New(errors.Internal, op, "unit name is an alias")
	}
	if path := parseAliasPath(units[0].Path); path != "" {
		return nil, errors.New(errors.Internal, op, "unit name is an alias")
	}

	if err = os.Remove(units[0].Path); err != nil {
		return nil, errors.Wrap(op, err, "failed to remove alias symlink")
	}
	if err = s.dbus.ReloadContext(ctx); err != nil {
		return nil, errors.Wrap(op, err, "failed to reload daemon")
	}

	return nil, nil
}

func (s *Server) GetServiceLogs(req *srv.GetServiceLogsRequest, stream grpc.ServerStreamingServer[srv.GetServiceLogsResponse]) error {
	const op = "systemd.(Server).GetServiceLogs"
	ctx := stream.Context()
	// the max number of lines able to be returned by the endpoint
	const maxlines = 1000
	if err := auth.Authorize(ctx, auth.Read, auth.Systemd, auth.Logs); err != nil {
		return errors.Wrap(op, err, "failed to authorize request")
	}

	switch {
	case req.Name == "":
		return errors.New(errors.BadRequest, op, "missing required field: name")
	}

	j, err := sdjournal.NewJournal()
	if err != nil {
		return errors.Wrap(op, err, "failed to create journal")
	}

	defer j.Close()
	j.AddMatch((&sdjournal.Match{
		Field: sdjournal.SD_JOURNAL_FIELD_SYSTEMD_UNIT,
		Value: req.Name,
	}).String())

	if req.Seek != nil {
		if err = j.SeekCursor(*req.Seek); err != nil {
			return errors.Wrap(op, err, fmt.Sprintf("failed to seek to \"%s\" within journal", *req.Seek))
		}
		l := false
		req.Live = &l
	} else if err = j.SeekTail(); err != nil {
		return errors.Wrap(op, err, "failed to seek to tail within journal")
	}

	res := &srv.GetServiceLogsResponse{
		Name: req.Name,
	}
	var lines uint32 = 100
	if req.Lines != nil {
		lines = min(*req.Lines, maxlines)
	}

	readEntry := func() error {
		entry, err := j.GetEntry()
		if err != nil {
			return errors.Wrap(op, err, "failed to retrieve journal entry")
		}
		log := &srv.LogEntry{
			Timestamp: entry.RealtimeTimestamp,
			Cursor:    entry.Cursor,
			Fields:    make(map[string]string),
		}
		for k, v := range entry.Fields {
			log.Fields[strings.ToLower(k)] = v
		}
		res.Logs = append(res.Logs, log)
		return nil
	}

journal:
	for range lines {
		n, err := j.Previous()
		switch {
		case err != nil:
			return errors.Wrap(op, err, "failed to iterate within journal")
		case n == 0:
			break journal
		}
		if err := readEntry(); err != nil {
			return err
		}
	}
	stream.Send(res)
	if !req.GetLive() {
		return nil
	}

	jch := make(chan int)
	go func() {
		for {
			jch <- j.Wait(time.Millisecond * 100)
			if ctx.Err() != nil {
				return
			}
		}
	}()
	latest := ""
	for {
		select {
		case <-ctx.Done():
			<-jch // make sure that the gofunc exits before returning otherwise it'll crash
			return nil
		case r := <-jch:
			switch r {
			case sdjournal.SD_JOURNAL_NOP:
			case sdjournal.SD_JOURNAL_INVALIDATE:
			case sdjournal.SD_JOURNAL_APPEND:
				if len(res.Logs) > 0 {
					latest = res.Logs[0].Cursor
					res.Logs = make([]*srv.LogEntry, 0)
				}
				if err = j.SeekTail(); err != nil {
					return errors.Wrap(op, err, "failed to seek to tail within journal")
				}
			j2:
				for range 10 {
					n, err := j.Previous()
					switch {
					case err != nil:
						return errors.Wrap(op, err, "failed to iterate within journal")
					case n == 0:
						break j2
					}
					if cur, err := j.GetCursor(); err != nil {
						return errors.Wrap(op, err, "failed to get journal cursor")
					} else if cur == latest {
						break j2
					}
					if err := readEntry(); err != nil {
						return err
					}
				}
				if len(res.Logs) > 0 {
					stream.Send(res)
				}
			default:
				return errors.New(500, op, "unexpected journal code received")
			}
		}
	}
}
