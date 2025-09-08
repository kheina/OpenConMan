/*
Authentication is the process of verifying a user's identity, while authorization
determines what resources or actions that authenticated user is allowed to access
or perform.
*/
package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"

	"github.com/kheina/openconman/src/errors"
)

func Authorize(ctx context.Context, scope Scope, action Action) error {
	const op = "auth.Authorize"
	authorization := ""
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if a, ok := md["authorization"]; ok {
			authorization = strings.Join(a, ",")
		}
	}
	if authorization == "" {
		return errors.New(errors.Unauthorized, op, "no authorization header found")
	}
	return nil
}
