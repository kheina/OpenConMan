package auth

import (
	"strings"

	"github.com/kheina/openconman/src/gen/pbs/auth/auth"
)

// re-list the pb enums here for ease of use

// Classic CRUDL actions
const (
	Create = auth.ACTION_CREATE
	Read   = auth.ACTION_READ
	Update = auth.ACTION_UPDATE
	Delete = auth.ACTION_DELETE
	List   = auth.ACTION_LIST
)

const (
	Alias      = auth.SCOPE_ALIAS
	Containers = auth.SCOPE_CONTAINERS
	Systemd    = auth.SCOPE_SYSTEMD
	Daemon     = auth.SCOPE_DAEMON
	Logs       = auth.SCOPE_LOGS
)

func ActionToString(a auth.ACTION) string {
	switch a {
	case auth.ACTION_ANY_ACTION:
		return "*"
	default:
		return strings.ToLower(a.String())
	}
}

func ScopeToString(s auth.SCOPE) string {
	switch s {
	case auth.SCOPE_ANY_SCOPE:
		return "*"
	case auth.SCOPE_ALL_SCOPES:
		return "**"
	default:
		return strings.ToLower(s.String())
	}
}
