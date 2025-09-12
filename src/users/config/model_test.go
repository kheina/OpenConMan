package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kheina/openconman/src/auth"
	authpb "github.com/kheina/openconman/src/gen/pbs/auth/auth"
	userpb "github.com/kheina/openconman/src/gen/pbs/auth/user"
	"github.com/kheina/openconman/src/users/config"
)

func Test_Authorized(t *testing.T) {
	defaultPerm := &userpb.Permission{
		Scope: authpb.SCOPE_UNKNOWN_SCOPE,
		Actions: []authpb.ACTION{
			authpb.ACTION_UNKNOWN_ACTION,
		},
	}
	tests := []struct {
		name     string
		perms    []*userpb.Permission
		scope    []authpb.SCOPE
		action   authpb.ACTION
		expected bool
	}{
		{
			name:   "systemd:create, userpb has explicit permission",
			scope:  []authpb.SCOPE{authpb.SCOPE_SYSTEMD},
			action: authpb.ACTION_CREATE,
			perms: []*userpb.Permission{
				defaultPerm,
				{
					Scope: authpb.SCOPE_SYSTEMD,
					Actions: []authpb.ACTION{
						authpb.ACTION_CREATE,
					},
				},
			},
			expected: true,
		},
		{
			name:   "systemd:create, userpb has no Permission",
			scope:  []authpb.SCOPE{authpb.SCOPE_SYSTEMD},
			action: authpb.ACTION_CREATE,
			perms: []*userpb.Permission{
				defaultPerm,
			},
			expected: false,
		},
		{
			name:   "systemd:alias:create, userpb has *:* Permission",
			scope:  []authpb.SCOPE{authpb.SCOPE_SYSTEMD, authpb.SCOPE_ALIAS},
			action: authpb.ACTION_CREATE,
			perms: []*userpb.Permission{
				defaultPerm,
				{
					Scope: authpb.SCOPE_ANY_SCOPE,
					Actions: []authpb.ACTION{
						authpb.ACTION_ANY_ACTION,
					},
				},
			},
			expected: false,
		},
		{
			name:   "systemd:create, userpb has *:* Permission",
			scope:  []authpb.SCOPE{authpb.SCOPE_SYSTEMD},
			action: authpb.ACTION_CREATE,
			perms: []*userpb.Permission{
				defaultPerm,
				{
					Scope: authpb.SCOPE_ANY_SCOPE,
					Actions: []authpb.ACTION{
						authpb.ACTION_ANY_ACTION,
					},
				},
			},
			expected: true,
		},
		{
			name:   "systemd:create, userpb has **:* Permission",
			scope:  []authpb.SCOPE{authpb.SCOPE_SYSTEMD},
			action: authpb.ACTION_CREATE,
			perms: []*userpb.Permission{
				defaultPerm,
				{
					Scope: authpb.SCOPE_ALL_SCOPES,
					Actions: []authpb.ACTION{
						authpb.ACTION_ANY_ACTION,
					},
				},
			},
			expected: true,
		},
		{
			name:   "systemd:alias:create, userpb has **:* Permission",
			scope:  []authpb.SCOPE{authpb.SCOPE_SYSTEMD, authpb.SCOPE_ALIAS},
			action: authpb.ACTION_CREATE,
			perms: []*userpb.Permission{
				defaultPerm,
				{
					Scope: authpb.SCOPE_ALL_SCOPES,
					Actions: []authpb.ACTION{
						authpb.ACTION_ANY_ACTION,
					},
				},
			},
			expected: true,
		},
		{
			name:   "systemd:alias:create, userpb has explicit Permission",
			scope:  []authpb.SCOPE{authpb.SCOPE_SYSTEMD, authpb.SCOPE_ALIAS},
			action: authpb.ACTION_CREATE,
			perms: []*userpb.Permission{
				defaultPerm,
				{
					Scope: authpb.SCOPE_SYSTEMD,
					Permissions: []*userpb.Permission{
						{
							Scope: authpb.SCOPE_ALIAS,
							Actions: []authpb.ACTION{
								authpb.ACTION_CREATE,
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name:   "systemd:alias:create, userpb has systemd:**:* Permission",
			scope:  []authpb.SCOPE{authpb.SCOPE_SYSTEMD, authpb.SCOPE_ALIAS},
			action: authpb.ACTION_CREATE,
			perms: []*userpb.Permission{
				defaultPerm,
				{
					Scope: authpb.SCOPE_SYSTEMD,
					Permissions: []*userpb.Permission{
						{
							Scope: authpb.SCOPE_ALL_SCOPES,
							Actions: []authpb.ACTION{
								authpb.ACTION_ANY_ACTION,
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name:   "systemd:create, userpb has systemd:**:* Permission",
			scope:  []authpb.SCOPE{authpb.SCOPE_SYSTEMD},
			action: authpb.ACTION_CREATE,
			perms: []*userpb.Permission{
				defaultPerm,
				{
					Scope: authpb.SCOPE_SYSTEMD,
					Permissions: []*userpb.Permission{
						{
							Scope: authpb.SCOPE_ALL_SCOPES,
							Actions: []authpb.ACTION{
								authpb.ACTION_ANY_ACTION,
							},
						},
					},
				},
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &config.UserConfig{
				Permissions: tt.perms,
			}
			if tt.expected {
				assert.Nil(t, auth.CheckPermission(u, tt.action, tt.scope...))
			} else {
				assert.NotNil(t, auth.CheckPermission(u, tt.action, tt.scope...))
			}
		})
	}
}
