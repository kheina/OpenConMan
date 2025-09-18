package containers

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/kheina/openconman/src/auth"
	"github.com/kheina/openconman/src/errors"
	"github.com/kheina/openconman/src/gen/pbs/api/containers"
	srv "github.com/kheina/openconman/src/gen/srv/api/docker"
)

type Server struct {
	srv.UnimplementedContainerServer

	client ContainerClient
}

type ContainerClient interface {
	listContainers(context.Context) ([]*containers.Container, error)
}

func NewServer(logger hclog.Logger) (*Server, error) {
	const op = "containers.NewServer"

	// attempt to bind to podman
	c, perr := newPodman()
	if perr == nil {
		if _, perr = c.listContainers(context.Background()); perr == nil {
			return &Server{
				client: c,
			}, nil
		}
	}
	logger.Info("podman unavailable", "err", perr)

	// attempt to bind to docker
	c, derr := newDocker()
	if derr == nil {
		if _, derr = c.listContainers(context.Background()); derr == nil {
			return &Server{
				client: c,
			}, nil
		}
	}
	logger.Info("docker unavailable", "err", derr)

	return nil, fmt.Errorf("%s: no viable container client found\npodman err: %w\ndocker err: %w", op, perr, derr)
}

func (s *Server) GetContainers(ctx context.Context, req *srv.GetContainerStatusesRequest) (*srv.GetContainerStatusesResponse, error) {
	const op = "containers.(Server).GetContainers"
	if err := auth.Authorize(ctx, auth.List, auth.Containers); err != nil {
		return nil, errors.Wrap(op, err, "failed to authorize request")
	}

	items, err := s.client.listContainers(ctx)
	if err != nil {
		return nil, errors.Wrap(op, err, "client failed to list containers")
	}

	return &srv.GetContainerStatusesResponse{
		Items: items,
	}, nil
}
