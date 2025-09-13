package pkg

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kheina/openconman/src/auth"
	"github.com/kheina/openconman/src/errors"
	srv "github.com/kheina/openconman/src/gen/srv/api/pkg"
	"github.com/kheina/openconman/src/version"
)

type Server struct {
	srv.UnimplementedPackagesServer
}

func New() (*Server, error) {
	return &Server{}, nil
}

const releases = "https://api.github.com/repos/kheina/openconman/releases"

func (s *Server) GetDaemonUpdate(ctx context.Context, req *srv.GetDaemonUpdateRequest) (*srv.GetDaemonUpdateResponse, error) {
	const op = "pkg.(Server).GetDaemonUpdate"
	if err := auth.Authorize(ctx, auth.Read, auth.Daemon); err != nil {
		return nil, errors.Wrap(op, err, "failed to authorize request")
	}

	v, err := version.New()
	if err != nil {
		return nil, errors.Wrap(op, err, "failed to init version thingy")
	}

	r, err := http.Get(releases)
	if err != nil {
		return nil, errors.Wrap(op, err, "failed to retrieve releases")
	}

	b := make([]byte, r.ContentLength)
	n, err := r.Body.Read(b)
	if err != nil {
		return nil, errors.Wrap(op, err, "failed to read releases")
	}

	releases := []GithubRelease{}
	if err = json.Unmarshal(b[:n], &releases); err != nil {
		return nil, errors.Wrap(op, err, "failed to unmarshal releases")
	}

	res := &srv.GetDaemonUpdateResponse{
		Current: v.String(),
		Dev:     v.Dev,
	}
releases:
	for _, r := range releases {
		if r.Prerelease == v.Dev {
			res.Latest = r.Name
			for _, a := range r.Assets {
				if a.Name == "conman" {
					res.Asset = a.Id
					break releases
				}
			}
			break releases
		}
	}

	return res, nil
}
