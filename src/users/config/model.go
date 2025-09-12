package config

import (
	"fmt"
	"strings"

	pb "github.com/kheina/openconman/src/gen/pbs/auth/auth"
	"github.com/kheina/openconman/src/gen/pbs/auth/user"
	"github.com/kheina/openconman/src/scopes"
)

type (
	UserConfig struct {
		User        *user.UserData
		Permissions []*user.Permission
		Password    *Password
	}

	Password struct {
		Hash []byte
	}

	perms interface {
		GetPermissions() []*user.Permission
	}
)

func New(name string) *UserConfig {
	return &UserConfig{
		Password: &Password{
			// this must be initialized for storing the hash later
		},
		User: &user.UserData{
			Id:   strings.ToLower(name),
			Name: name,
		},
	}
}

func (*UserConfig) GetScope() pb.SCOPE                   { return pb.SCOPE_UNKNOWN_SCOPE }
func (u *UserConfig) GetPermissions() []*user.Permission { return u.Permissions }

func trav(scopes *[]string, scope string, p *user.Permission) {
	scope = scope + scopeToString(p.Scope) + ":"
	for _, a := range p.Actions {
		*scopes = append(*scopes, scope+actionToString(a))
	}
	for _, pp := range p.Permissions {
		trav(scopes, scope, pp)
	}
}

func (u *UserConfig) GetScopes() []string {
	scopes := []string{}
	for _, p := range u.Permissions {
		trav(&scopes, "", p)
	}
	return scopes
}

// AddScope takes a scope string in the format of scope:scope...:action gives
// the user that permission
func (u *UserConfig) AddScope(scopeStr string) error {
	const op = "config.(UserConfig).AddScope"
	s := strings.Split(scopeStr, ":")
	switch {
	case len(s) == 1:
		return fmt.Errorf("%s, scope missing action", op)
	}

	action := scopes.NewAction(s[len(s)-1])
	if action == pb.ACTION_UNKNOWN_ACTION {
		return fmt.Errorf("%s, unknown action key: %s", op, s[len(s)-1])
	}
	var p perms = u
search:
	for _, sc := range s[:len(s)-1] {
		s := scopes.NewScope(sc)
		if s == pb.SCOPE_UNKNOWN_SCOPE {
			return fmt.Errorf("%s, unknown scope key: %s", op, sc)
		}
		for _, pp := range p.GetPermissions() {
			if pp.Scope == s {
				p = pp
				continue search
			}
		}
		// this must be a POINTER to the slice
		var pp *[]*user.Permission
		switch p := p.(type) {
		case *UserConfig:
			pp = &p.Permissions
		case *user.Permission:
			pp = &p.Permissions
		default:
			return fmt.Errorf("%s: unexpected perms value: %T", op, p)
		}
		ppp := &user.Permission{
			Scope: s,
		}
		*pp = append(*pp, ppp)
		p = ppp
	}
	pp, ok := p.(*user.Permission)
	if !ok {
		return fmt.Errorf("%s: resultant Permission var not of type (*Permission)", op)
	}
	pp.Actions = append(pp.Actions, action)
	return nil
}
