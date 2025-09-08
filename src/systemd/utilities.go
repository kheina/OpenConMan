package systemd

import (
	"io/fs"
	"os"
	"strings"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/kheina/openconman/src/errors"
)

func newAliasPath(unit *dbus.UnitFile) (string, error) {
	const op = "systemd.newAliasPath"
	i := strings.LastIndexByte(unit.Path, '/')
	if i < 0 {
		return "", errors.New(500, op, "unit path did not contain slash")
	}

	return unit.Path[:i+1] + "ocm-" + unit.Path[i+1:], nil
}

func pathToName(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}

func getAliasNames(files []dbus.UnitFile) []string {
	aliases := []string{}
	for _, f := range files {
		if f.Type != string(alias) {
			continue
		}
		aliases = append(aliases, pathToName(f.Path))
	}
	return aliases
}

// parseAliasPath checks if a given path is an alias and returns the filepath of
// the aliased unit. returns "" when path is not an alias
func parseAliasPath(path string) string {
	res := ""
	for {
		s, err := os.Lstat(path)
		if err != nil || s.Mode()&fs.ModeSymlink == 0 {
			break
		}
		// f.Path is a symlink (alias)

		path, err = os.Readlink(path)
		if err != nil {
			break
		}
		res = path
	}
	return res
}

// aliasMap returns a map of unit names to their alias names.
// map[name] -> alias
func aliasMap(files []dbus.UnitFile) map[string]string {
	aliases := make(map[string]string)
	for _, f := range files {
		if f.Type != string(alias) {
			continue
		}
		// f is an alias

		res := parseAliasPath(f.Path)
		if res == "" {
			continue
		}

		// f.Path -> res
		aliases[pathToName(res)] = pathToName(f.Path)
	}
	return aliases
}
