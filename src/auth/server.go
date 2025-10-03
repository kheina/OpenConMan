package auth

import (
	"context"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/kheina/openconman/src/errors"
	srv "github.com/kheina/openconman/src/gen/srv/api/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	srv.UnimplementedAuthServer

	logger hclog.Logger
}

func New(opt ...Option) (*Server, error) {
	opts := getOpts(opt...)
	s := &Server{
		logger: opts.withLogger,
	}
	return s, nil
}

func (s *Server) Login(ctx context.Context, req *srv.PostLoginRequest) (*srv.PostLoginResponse, error) {
	const op = "auth.(Server).Login"
	switch {
	case req.Username == "":
		return nil, errors.New(errors.BadRequest, op, "username empty")
	case req.Password == "":
		return nil, errors.New(errors.BadRequest, op, "password empty")
	}
	user, err := Login(req.Username, []byte(req.Password))
	if err != nil {
		return nil, errors.Wrap(op, err, "login failed")
	}
	exp := time.Now().Add(time.Hour * 24) // 1 day
	tok, err := NewAuthToken(user, exp)
	if err != nil {
		return nil, errors.Wrap(op, err, "failed to create auth token")
	}
	stok := string(tok)
	if err = grpc.SetHeader(ctx, metadata.Pairs("Set-Cookie", "ocm-auth="+stok)); err != nil {
		// failed to set header ig
	}
	return &srv.PostLoginResponse{
		Token:   stok,
		Scopes:  user.GetScopes(),
		Expires: timestamppb.New(exp),
	}, nil
}
