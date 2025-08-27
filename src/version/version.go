package version

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kheina/openconman/src/errors"
)

type Version struct {
	Major     int64
	Minor     int64
	Patch     int64
	Rev       string
	Dev       bool
	BuildDate time.Time
}

func (v *Version) String() string {
	vstr := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Dev {
		vstr = fmt.Sprintf("%s+dev", vstr)
	}
	return vstr
}

func New() (*Version, error) {
	const op = "version.New"
	if VersionStr == "" {
		// currently being run via `go run`
		vb, err := os.ReadFile("VERSION")
		if err != nil {
			return nil, errors.Wrap(op, err, "failed to read version file")
		}
		VersionStr = string(vb)
	}
	vstr := strings.Split(strings.TrimSpace(VersionStr), ".")
	if len(vstr) != 3 {
		return nil, fmt.Errorf("%s: incorrent number of version sub components, should be exactly 3, got %d (%s)", op, len(vstr), VersionStr)
	}
	var err error
	v := &Version{}

	if v.Major, err = strconv.ParseInt(vstr[0], 10, 64); err != nil {
		return nil, errors.Wrap(op, err, "failed to parse major version component")
	}
	if v.Minor, err = strconv.ParseInt(vstr[1], 10, 64); err != nil {
		return nil, errors.Wrap(op, err, "failed to parse minor version component")
	}
	if v.Patch, err = strconv.ParseInt(vstr[2], 10, 64); err != nil {
		return nil, errors.Wrap(op, err, "failed to parse patch version component")
	}
	if Timestamp == "" {
		// currently being run via `go run`
		v.BuildDate = time.Now()
	} else {
		if ts, err := strconv.ParseInt(Timestamp, 10, 64); err != nil {
			return nil, errors.Wrap(op, err, "failed to parse build timestamp")
		} else {
			v.BuildDate = time.Unix(ts, 0)
		}
	}

	v.Rev = Commit
	v.Dev = Branch != "main"
	return v, nil
}
