package config

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/kheina/openconman/src/errors"
	"github.com/kheina/openconman/src/gen/pbs/auth/auth"
	"github.com/kheina/openconman/src/gen/pbs/auth/user"
	"github.com/kheina/openconman/src/util"
)

func actionToString(a auth.ACTION) string {
	switch a {
	case auth.ACTION_ANY_ACTION:
		return "*"
	default:
		return strings.ToLower(a.String())
	}
}

func scopeToString(s auth.SCOPE) string {
	switch s {
	case auth.SCOPE_ANY_SCOPE:
		return "*"
	case auth.SCOPE_ALL_SCOPES:
		return "**"
	default:
		return strings.ToLower(s.String())
	}
}

func (u *UserConfig) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	_, err := buf.WriteString("[user]\n")
	if err != nil {
		return nil, err
	}

	if _, err = buf.WriteString(fmt.Sprintf("id = %s\n", u.User.Id)); err != nil {
		return nil, err
	}
	if _, err = buf.WriteString(fmt.Sprintf("name = %s\n", u.User.Name)); err != nil {
		return nil, err
	}
	if _, err = buf.WriteString(fmt.Sprintf("hash = %s\n", u.Password.Hash)); err != nil {
		return nil, err
	}

	if _, err = buf.WriteString("\n[scopes]\n"); err != nil {
		return nil, err
	}
	for _, s := range u.GetScopes() {
		if _, err = buf.WriteString(s + "\n"); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

var userRegex = regexp.MustCompile(`\[user\]\n([\s\S]*?)(?:\n\[|$)`)
var idRegex = regexp.MustCompile(`id\s*=\s*(.*?)(?:\s|\n|$)`)
var nameRegex = regexp.MustCompile(`name\s*=\s*(.*?)(?:\s|\n|$)`)
var hashRegex = regexp.MustCompile(`hash\s*=\s*(.*?)(?:\s|\n|$)`)
var scopesRegex = regexp.MustCompile(`\[scopes\]\n([\s\S]*?)(?:\n\[|$)`)

func (u *UserConfig) Unmarshal(b []byte) error {
	const op = "config.(UserConfig).Unmarshal"
	udata := userRegex.FindSubmatch(b)
	if udata == nil {
		return fmt.Errorf("%s: no user data found in config", op)
	}
	sdata := scopesRegex.FindSubmatch(b)
	if sdata == nil {
		return fmt.Errorf("%s: no scopes data found in config", op)
	}

	id := idRegex.FindSubmatch(udata[1])
	name := nameRegex.FindSubmatch(udata[1])
	switch {
	case id == nil:
		return fmt.Errorf("%s: could not find id in user config", op)
	case name == nil:
		return fmt.Errorf("%s: could not find name in user config", op)
	case u.User == nil:
		u.User = &user.UserData{}
	}
	u.User.Id = string(id[1])
	u.User.Name = string(name[1])

	hash := hashRegex.FindSubmatch(udata[1])
	switch {
	case hash == nil:
		return fmt.Errorf("%s: could not find password hash in user config", op)
	case u.Password == nil:
		u.Password = &Password{}
	}
	u.Password.Hash = hash[1]

	for _, s := range strings.Split(string(sdata[1]), "\n") {
		if s == "" {
			continue
		}
		if err := u.AddScope(s); err != nil {
			return fmt.Errorf("%s: failed to add scope to user: %w", op, err)
		}
	}
	return nil
}

func Read(id string) (*UserConfig, error) {
	const op = "config.Read"
	filename := fmt.Sprintf("./users/%s.conf", id)
	if !util.PathExists(filename) {
		return nil, fmt.Errorf("%s: user does not exist", op)
	}
	udata, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(op, err, "failed to read userdata")
	}
	conf := New("")
	if err = conf.Unmarshal(udata); err != nil {
		return nil, errors.Wrap(op, err, "failed to unmarshal userdata")
	}
	return conf, nil
}
