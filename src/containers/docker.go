package containers

import (
	"context"
	"fmt"
	"time"

	"github.com/kheina/openconman/src/errors"
	pb "github.com/kheina/openconman/src/gen/pbs/api/containers"
	"github.com/kheina/openconman/src/util"
	"github.com/moby/moby/api/types/container"
	moby "github.com/moby/moby/client"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type docker struct {
	client *moby.Client
}

func newDocker() (ContainerClient, error) {
	const op = "containers.newDocker"
	d := &docker{}
	var err error
	d.client, err = moby.NewClientWithOpts(moby.FromEnv, moby.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("%s: failed to fetch docker client: %w", op, err)
	}
	return d, nil
}

func (d *docker) listContainers(ctx context.Context) ([]*pb.Container, error) {
	const op = "containers.(docker).listContainers"
	c, err := d.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, errors.Wrap(op, err, "failed to list containers")
	}

	items := []*pb.Container{}
	for _, con := range c {
		item := &pb.Container{
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
			HostConfig: &pb.HostConfig{
				NetworkMode: con.HostConfig.NetworkMode,
				Annotations: con.HostConfig.Annotations,
			},
		}
		if con.Health != nil {
			item.Health = &pb.Health{
				Status:        con.Health.Status,
				FailingStreak: uint64(con.Health.FailingStreak),
			}
		}
		for _, m := range con.Mounts {
			item.Mounts = append(item.Mounts, &pb.MountPoint{
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
			item.Networks = append(item.Networks, &pb.Port{
				Ip:          util.OptionalString(m.IP),
				PrivatePort: uint32(m.PrivatePort),
				PublicPort:  &pub,
				Type:        m.Type,
			})
		}
		items = append(items, item)
	}

	return items, nil
}
