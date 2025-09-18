package containers

import (
	"context"
	"fmt"
	"strings"

	"github.com/containers/podman/v5/pkg/bindings"
	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/kheina/openconman/src/errors"
	pb "github.com/kheina/openconman/src/gen/pbs/api/containers"
	"github.com/kheina/openconman/src/util"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type podman struct {
	connCtx context.Context
}

func newPodman() (ContainerClient, error) {
	p := &podman{}
	var err error
	p.connCtx, err = bindings.NewConnection(context.Background(), "unix:///run/podman/podman.sock")
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *podman) listContainers(ctx context.Context) ([]*pb.Container, error) {
	const op = "containers.(podman).listContainers"
	c, err := containers.List(p.connCtx, nil)
	if err != nil {
		return nil, errors.Wrap(op, err, "failed to list containers")
	}
	items := []*pb.Container{}
	for _, con := range c {
		item := &pb.Container{
			Id:      con.ID,
			Image:   con.Image,
			ImageId: con.ImageID,
			Command: strings.Join(con.Command, " "),
			Created: timestamppb.New(con.Created),
			Labels:  con.Labels,
			Names:   con.Names,
			State:   con.State,
			Status:  con.Status,
		}
		if con.Size != nil {
			item.SizeRw = con.Size.RwSize
			item.SizeRootFs = con.Size.RootFsSize
		}
		for i, m := range con.Mounts {
			fmt.Printf("mount[%d]: %s\n", i, m)
			// item.Mounts = append(item.Mounts, &pb.MountPoint{
			// 	Type:        util.OptionalString(string(m.Type)),
			// 	Name:        util.OptionalString(m.Name),
			// 	Source:      m.Source,
			// 	Destination: m.Destination,
			// 	Driver:      util.OptionalString(m.Driver),
			// 	Mode:        m.Mode,
			// 	Rw:          m.RW,
			// 	Propagation: util.OptionalString(string(m.Propagation)),
			// })
		}
		for _, m := range con.Ports {
			pub := uint32(m.HostPort)
			item.Networks = append(item.Networks, &pb.Port{
				Ip:          util.OptionalString(m.HostIP),
				PrivatePort: uint32(m.ContainerPort),
				PublicPort:  &pub,
				Type:        m.Protocol,
			})
		}
		items = append(items, item)
	}

	return items, nil
}
