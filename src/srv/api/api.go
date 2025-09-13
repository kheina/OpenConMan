package api

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/kheina/openconman/src/auth"
	"github.com/kheina/openconman/src/containers"
	"github.com/kheina/openconman/src/errors"
	srvauth "github.com/kheina/openconman/src/gen/srv/api/auth"
	"github.com/kheina/openconman/src/gen/srv/api/docker"
	srvpkg "github.com/kheina/openconman/src/gen/srv/api/pkg"
	srvsys "github.com/kheina/openconman/src/gen/srv/api/systemd"
	"github.com/kheina/openconman/src/pkg"
	"github.com/kheina/openconman/src/systemd"
)

func Handler(ctx context.Context, gs *grpc.Server, grpcAddr string, logger hclog.Logger, conf *tls.Config) (http.Handler, error) {
	const op = "api.Handler"
	grpcopts := []grpc.DialOption{}
	name := "conman"
	if conf == nil {
		grpcopts = append(grpcopts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		pool := x509.NewCertPool()
		for _, c := range conf.Certificates {
			for _, cc := range c.Certificate {
				parsed, err := x509.ParseCertificate(cc)
				if err != nil {
					return nil, fmt.Errorf("%s: failed to parse tls cert: %w", op, err)
				}
				pool.AddCert(parsed)
				name = parsed.Subject.CommonName
				logger.Debug("cert added to gateway ca pool")
			}
		}
		grpcopts = append(grpcopts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, name)))
	}

	conn, err := grpc.NewClient(grpcAddr, grpcopts...)
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
		runtime.WithOutgoingHeaderMatcher(func(key string) (string, bool) {
			switch key {
			case "Set-Cookie":
				return key, true
			default:
				return runtime.DefaultHeaderMatcher(key)
			}
		}),
		runtime.WithErrorHandler(errors.ApiErrorHandler(logger)),
	)

	// register all of the different grpc servers
	// NOTE: this MUST be done at the same time as registering the handlers
	if srv, err := containers.NewServer(); err != nil {
		return nil, fmt.Errorf("%s: failed to create containers server: %w", op, err)
	} else {
		docker.RegisterContainerServer(gs, srv)
	}
	if srv, err := systemd.New(ctx, systemd.WithLogger(logger)); err != nil {
		return nil, fmt.Errorf("%s: failed to create systemd server: %w", op, err)
	} else {
		srvsys.RegisterSystemdServer(gs, srv)
	}
	if srv, err := auth.New(auth.WithLogger(logger)); err != nil {
		return nil, fmt.Errorf("%s: failed to create auth server: %w", op, err)
	} else {
		srvauth.RegisterAuthServer(gs, srv)
	}
	if srv, err := pkg.New(); err != nil {
		return nil, fmt.Errorf("%s: failed to create pkg server: %w", op, err)
	} else {
		srvpkg.RegisterPackagesServer(gs, srv)
	}

	// register all the different grpc handlers
	// NOTE: this MUST be done at the same time as registering the servers
	if err = docker.RegisterContainerHandler(ctx, gwmux, conn); err != nil {
		return nil, fmt.Errorf("%s: failed to register container gateway: %w", op, err)
	}
	if err = srvsys.RegisterSystemdHandler(ctx, gwmux, conn); err != nil {
		return nil, fmt.Errorf("%s: failed to register systemd gateway: %w", op, err)
	}
	if err = srvauth.RegisterAuthHandler(ctx, gwmux, conn); err != nil {
		return nil, fmt.Errorf("%s: failed to register auth gateway: %w", op, err)
	}
	if err = srvpkg.RegisterPackagesHandler(ctx, gwmux, conn); err != nil {
		return nil, fmt.Errorf("%s: failed to register pkg gateway: %w", op, err)
	}

	return gwmux, nil
}
