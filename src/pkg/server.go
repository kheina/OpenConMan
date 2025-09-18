package pkg

import (
	"context"
	"encoding/json"
	"io"
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

const relAllUrl = "https://api.github.com/repos/kheina/openconman/releases"
const relLatestUrl = "https://api.github.com/repos/kheina/openconman/releases/latest"

func (s *Server) GetDaemonUpdate(ctx context.Context, req *srv.GetDaemonUpdateRequest) (*srv.GetDaemonUpdateResponse, error) {
	const op = "pkg.(Server).GetDaemonUpdate"
	if err := auth.Authorize(ctx, auth.Read, auth.Daemon); err != nil {
		return nil, errors.Wrap(op, err, "failed to authorize request")
	}

	v, err := version.New()
	if err != nil {
		return nil, errors.Wrap(op, err, "failed to init version thingy")
	}

	res := &srv.GetDaemonUpdateResponse{
		Current: v.String(),
		Dev:     v.Dev,
	}

	if v.Dev {
		r, err := http.Get(relAllUrl)
		if err != nil {
			return nil, errors.Wrap(op, err, "failed to retrieve releases")
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, errors.Wrap(op, err, "failed to read releases")
		}

		releases := []GithubRelease{}
		if err = json.Unmarshal(b, &releases); err != nil {
			return nil, errors.Wrap(op, err, "failed to unmarshal releases")
		}

	releases:
		for _, r := range releases {
			if r.Prerelease != v.Dev {
				continue
			}
			res.Latest = r.Name
			for _, a := range r.Assets {
				if a.Name == "conman" {
					res.Asset = &a.Id
					break releases
				}
			}
			break
		}
	} else {
		r, err := http.Get(relLatestUrl)
		if err != nil {
			return nil, errors.Wrap(op, err, "failed to retrieve latest release")
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, errors.Wrap(op, err, "failed to read latest release")
		}

		release := GithubRelease{}
		if err = json.Unmarshal(b, &release); err != nil {
			return nil, errors.Wrap(op, err, "failed to unmarshal release")
		}

		res.Latest = release.Name
		for _, a := range release.Assets {
			if a.Name == "conman" {
				res.Asset = &a.Id
				break
			}
		}
	}

	if res.Newer, err = v.Compare(res.Latest); err != nil {
		return nil, errors.Wrap(op, err, "could not compare current to latest")
	}

	return res, nil
}
