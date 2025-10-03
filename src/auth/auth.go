/*
Authentication is the process of verifying a user's identity, while authorization
determines what resources or actions that authenticated user is allowed to access
or perform.
*/
package auth

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/hashicorp/go-secure-stdlib/base62"
	"github.com/kheina/openconman/src/errors"
	argonpb "github.com/kheina/openconman/src/gen/pbs/auth/argon2"
	pb "github.com/kheina/openconman/src/gen/pbs/auth/auth"
	"github.com/kheina/openconman/src/gen/pbs/auth/user"
	"github.com/kheina/openconman/src/users/config"
	"github.com/kheina/openconman/src/util"
)

const loginFailed = "LoginFailed"

func Login(name string, password []byte) (*config.UserConfig, error) {
	const op = "auth.Login"
	conf, err := config.Read(strings.ToLower(name))
	if err != nil {
		return nil, errors.Wrap(op, err, "failed to read user", errors.WithStatusCode(errors.Unauthorized), errors.WithErrorCode(loginFailed))
	}

	// check for all available authentication methods
	switch {
	case conf.Password.Hash != nil:
		hash := &argonpb.Argon2Hash{}
		hashdata, err := base64.RawURLEncoding.AppendDecode([]byte{}, conf.Password.Hash)
		if err != nil {
			return nil, errors.Wrap(op, err, "failed to decode user password data", errors.WithStatusCode(errors.Unauthorized), errors.WithErrorCode(loginFailed))
		}
		if err = proto.Unmarshal(hashdata, hash); err != nil {
			return nil, errors.Wrap(op, err, "failed to unmarshal user password hash", errors.WithStatusCode(errors.Unauthorized), errors.WithErrorCode(loginFailed))
		}
		rehash := argon2.IDKey(password, hash.Salt, hash.Config.Time, hash.Config.Memory*1024, uint8(hash.Config.Parallelism), hash.Config.HashLength)
		if subtle.ConstantTimeCompare(hash.Hash, rehash) != 1 {
			return nil, errors.New(errors.Unauthorized, op, "login failed", errors.WithErrorCode(loginFailed))
		}
		return conf, nil
	default:
		return nil, errors.New(errors.Unauthorized, op, "user config does not contain a valid authentication method", errors.WithErrorCode(loginFailed))
	}
}

type authKey struct {
	name string
	pub  ed25519.PublicKey
	priv ed25519.PrivateKey
}

var key *authKey = nil

type conf struct {
	Config  *config.UserConfig
	cleanup *time.Timer
}

type tokenMap map[string]conf

var tokens tokenMap = make(tokenMap)

func (m *tokenMap) GetUser(id string) *config.UserConfig {
	var conf *config.UserConfig

	if c, ok := (*m)[id]; ok {
		conf = c.Config
	}

	// tbh if a user needs fresh permissions, they can log out and back in.
	// in the future I would love to be able to listen for filesystem changes to
	// automatically do this, but I don't think it's worthwhile do it on every
	// request

	return conf
}

func (m *tokenMap) setUser(t *user.UserToken, c *config.UserConfig) {
	switch {
	case util.IsNil(t):
		return
	case util.IsNil(c):
		return
	case t.Id == "":
		return
	}
	// TODO: we should set a limit on the number of active tokens, and delete the oldest ones
	(*m)[t.Id] = conf{
		Config:  c,
		cleanup: time.AfterFunc(time.Until(t.Expires.AsTime()), func() { delete(*m, t.Id) }),
	}
}

// NewAuthToken generates a new auth token that has been signed, encoded, and is
// ready to be inserted directly into a request header.
func NewAuthToken(u *config.UserConfig, expires time.Time) ([]byte, error) {
	const op = "auth.NewAuthToken"
	if key == nil {
		pub, priv, err := ed25519.GenerateKey(nil)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to generate new private key: %w", op, err)
		}
		name, err := base62.Random(12)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to generate key name: %w", op, err)
		}
		key = &authKey{
			name: name,
			pub:  pub,
			priv: priv,
		}
		if !util.PathExists("./crypto") {
			if err = os.Mkdir("./crypto", 0644); err != nil {
				return nil, fmt.Errorf("%s: failed to persist public key: failed to create crypto dir: %w", op, err)
			}
		}
		if err = os.WriteFile(fmt.Sprintf("./crypto/%s.key", name), pub, 0644); err != nil {
			return nil, fmt.Errorf("%s: failed to persist public key: %w", op, err)
		}
	}
	tok := &user.UserToken{
		User:    u.User,
		Id:      rand.Text(),
		Expires: timestamppb.New(expires),
		Key:     key.name,
	}
	out, err := proto.Marshal(tok)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal user token: %w", op, err)
	}
	sig, err := key.priv.Sign(nil, out, crypto.Hash(0))
	if err != nil {
		return nil, fmt.Errorf("%s: failed to sign user token: %w", op, err)
	}

	// now that the token has been created, we need to store it locally so that it's valid for requests
	tokens.setUser(tok, u)

	return bytes.Join([][]byte{
		base64.RawURLEncoding.AppendEncode([]byte{}, out),
		base64.RawURLEncoding.AppendEncode([]byte{}, sig),
	}, []byte{'.'}), nil
}

const noToken = "NoBearerToken"
const tokenInvalid = "BearerTokenInvalid"

// Authorize accepts a request context, action, and scope list and parses out
// the authorization token and checks whethor or not the provided token has the
// capabilities implied by the provided scopes and action
//
// Returns nil if the user has access to the provided scope with the given token,
// otherwise returns an http 401 error if no authentication or invalid authentication
// was provided or an http 403 error with the phrase 'authenticated user does not
// have access to the scope "scope1:scope2:...:action"'
func Authorize(ctx context.Context, action pb.ACTION, scopes ...pb.SCOPE) error {
	const op = "auth.Authorize"
	bearer := ""
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if a, ok := md["authorization"]; ok {
			for _, authorization := range a {
				if len(authorization) > 7 && strings.ToLower(authorization[:7]) == "bearer " {
					bearer = authorization[7:]
					break
				}
			}
		}
	}
	if bearer == "" {
		return errors.New(
			errors.Unauthorized, op,
			"no bearer token found. make sure you are logged in and your auth token is sent via the authorization header preceded by \"Bearer \"",
			errors.WithErrorCode(noToken),
		)
	}
	parts := strings.Split(bearer, ".")
	if len(parts) != 2 {
		return errors.New(errors.Unauthorized, op, "failed to parse auth token", errors.WithErrorCode(tokenInvalid))
	}
	in, sig, err := []byte(nil), []byte(nil), error(nil)
	if in, err = base64.RawURLEncoding.DecodeString(parts[0]); err != nil {
		return errors.New(errors.Unauthorized, op, "failed to decode auth token. ensure auth token is base64 encoded", errors.WithErrorCode(tokenInvalid))
	}
	if sig, err = base64.RawURLEncoding.DecodeString(parts[1]); err != nil {
		return errors.New(errors.Unauthorized, op, "failed to decode auth signature. ensure auth token is base64 encoded", errors.WithErrorCode(tokenInvalid))
	}

	tok := &user.UserToken{}
	if err = proto.Unmarshal(in, tok); err != nil {
		return errors.New(errors.Unauthorized, op, "failed to unmarshal token data into user token", errors.WithErrorCode(tokenInvalid))
	}

	// gotta load the token public bytes
	pbytes, err := os.ReadFile(fmt.Sprintf("./crypto/%s.key", tok.Key))
	if err != nil {
		return fmt.Errorf("%s: failed to read public key: %w", op, err)
	}
	if !ed25519.Verify(ed25519.PublicKey(pbytes), in, sig) {
		return errors.New(errors.Unauthorized, op, "failed to verify token", errors.WithErrorCode(tokenInvalid))
	}

	return CheckPermission(tokens.GetUser(tok.GetId()), action, scopes...)
}

type Permissions interface {
	GetPermissions() []*user.Permission
	GetScope() pb.SCOPE
}

func hasScope(p Permissions, scope pb.SCOPE) Permissions {
	if p.GetScope() == pb.SCOPE_ALL_SCOPES {
		// we don't further traverse the tree when the PARENT matches on **
		return p
	}
	for _, pp := range p.GetPermissions() {
		switch pp.Scope {
		case scope:
		case pb.SCOPE_ANY_SCOPE:
		case pb.SCOPE_ALL_SCOPES:
		default:
			continue
		}
		return pp
	}
	return nil
}

func match(a []pb.ACTION, action pb.ACTION) bool {
	for _, a := range a {
		switch a {
		case action:
		case pb.ACTION_ANY_ACTION:
		default:
			continue
		}
		return true
	}
	return false
}

func newForbiddenErr(op string, action pb.ACTION, scopes []pb.SCOPE) error {
	scope := ""
	for _, s := range scopes {
		scope += ScopeToString(s) + ":"
	}
	scope += ActionToString(action)
	return errors.New(errors.Forbidden, op, fmt.Sprintf("authenticated user does not have access to the scope \"%s\"", scope))
}

func allowed(p *user.Permission, action pb.ACTION) bool {
	if match(p.Actions, action) {
		return true
	}
	// since all scopes matches no scope as well, we need to check if one is in there
	for _, p := range p.Permissions {
		if p.Scope == pb.SCOPE_ALL_SCOPES && match(p.Actions, action) {
			return true
		}
	}
	return false
}

// CheckPermission takes an action and an array of scopes and returns whether or
// not the loaded user config is capable of performing the given action in the
// provided scope.
// NOTE: the order of the scopes matters, and (action, scope1, scope2, ...) will
// be treated as the scope string scope1:scope2:...:action
//
// Returns nil if the user has access to the provided scope with the given action,
// otherwise returns an http 403 error with the phrase 'authenticated user does
// not have access to the scope "scope1:scope2:...:action"'
func CheckPermission(p Permissions, action pb.ACTION, scopes ...pb.SCOPE) error {
	const op = "auth.CheckPermission"
	if util.IsNil(p) {
		return errors.New(errors.Unauthorized, op, "user is nil")
	}
	for _, s := range scopes {
		if p = hasScope(p, s); p == nil {
			return newForbiddenErr(op, action, scopes)
		}
	}
	pp, ok := p.(*user.Permission)
	if !ok {
		return newForbiddenErr(op, action, scopes)
	}
	if allowed(pp, action) {
		return nil
	}
	return newForbiddenErr(op, action, scopes)
}
