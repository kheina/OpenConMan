package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/kheina/openconman/src/containers"
	"github.com/kheina/openconman/src/errors"
	"github.com/kheina/openconman/src/gen/srv/api/docker"
	srvsys "github.com/kheina/openconman/src/gen/srv/api/systemd"
	"github.com/kheina/openconman/src/systemd"
)

func Handler(ctx context.Context, grpcAddr string, logger hclog.Logger) (http.Handler, error) {
	const op = "api.Handler"
	conn, err := grpc.NewClient(
		grpcAddr,
		// TODO: update once we implement credentials
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to dial gRPC server: %w", op, err)
	}

	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
		runtime.WithErrorHandler(errors.ApiErrorHandler(logger)),
	)
	// register all the different handlers
	if err = docker.RegisterContainerHandler(ctx, gwmux, conn); err != nil {
		return nil, fmt.Errorf("%s: failed to register gateway: %w", op, err)
	}
	if err = srvsys.RegisterSystemdHandler(ctx, gwmux, conn); err != nil {
		return nil, fmt.Errorf("%s: failed to register gateway: %w", op, err)
	}

	return gwmux, nil
}

func RegisterGrpcServices(ctx context.Context, gs *grpc.Server, l hclog.Logger) error {
	const op = "api.RegisterGrpcServices"
	cSrv, err := containers.NewServer()
	if err != nil {
		return fmt.Errorf("%s: failed to create containers server: %w", op, err)
	}
	docker.RegisterContainerServer(gs, cSrv)

	sSrv, err := systemd.New(ctx, systemd.WithLogger(l))
	if err != nil {
		return fmt.Errorf("%s: failed to create systemd server: %w", op, err)
	}
	srvsys.RegisterSystemdServer(gs, sSrv)
	return nil
}
