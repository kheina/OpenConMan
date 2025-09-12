package containers

import (
	"context"
	"fmt"
	"time"

	"github.com/moby/moby/api/types/container"
	docker "github.com/moby/moby/client"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/kheina/openconman/src/auth"
	"github.com/kheina/openconman/src/errors"
	"github.com/kheina/openconman/src/gen/pbs/api/containers"
	srv "github.com/kheina/openconman/src/gen/srv/api/docker"
	"github.com/kheina/openconman/src/util"
)

type Server struct {
	srv.UnimplementedContainerServer

	client *docker.Client
}

func NewServer() (*Server, error) {
	const op = "containers.NewServer"
	client, err := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("%s: failed to fetch docker client: %w", op, err)
	}

	return &Server{
		client: client,
	}, nil
}

func (s *Server) GetContainers(ctx context.Context, req *srv.GetContainerStatusesRequest) (*srv.GetContainerStatusesResponse, error) {
	const op = "containers.(Server).GetContainers"
	if err := auth.Authorize(ctx, auth.List, auth.Containers); err != nil {
		return nil, errors.Wrap(op, err, "failed to authorize request")
	}

	c, err := s.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, errors.Wrap(op, err, "failed to list containers")
	}

	items := []*containers.Container{}
	for _, con := range c {
		item := &containers.Container{
			Id:         con.ID,
			Image:      con.Image,
			ImageId:    con.ImageID,
			Command:    con.Command,
			Created:    timestamppb.New(time.Unix(con.Created, 0)),
			Labels:     con.Labels,
			Names:      con.Names,
			SizeRw:     con.SizeRw,
			SizeRootFs: con.SizeRootFs,
			State:      con.State,
			Status:     con.Status,
			HostConfig: &containers.HostConfig{
				NetworkMode: con.HostConfig.NetworkMode,
				Annotations: con.HostConfig.Annotations,
			},
		}
		if con.Health != nil {
			item.Health = &containers.Health{
				Status:        con.Health.Status,
				FailingStreak: uint64(con.Health.FailingStreak),
			}
		}
		for _, m := range con.Mounts {
			item.Mounts = append(item.Mounts, &containers.MountPoint{
				Type:        util.OptionalString(string(m.Type)),
				Name:        util.OptionalString(m.Name),
				Source:      m.Source,
				Destination: m.Destination,
				Driver:      util.OptionalString(m.Driver),
				Mode:        m.Mode,
				Rw:          m.RW,
				Propagation: util.OptionalString(string(m.Propagation)),
			})
		}
		for _, m := range con.Ports {
			pub := uint32(m.PublicPort)
			item.Networks = append(item.Networks, &containers.Port{
				Ip:          util.OptionalString(m.IP),
				PrivatePort: uint32(m.PrivatePort),
				PublicPort:  &pub,
				Type:        m.Type,
			})
		}
		items = append(items, item)
	}

	return &srv.GetContainerStatusesResponse{
		Items: items,
	}, nil
}
