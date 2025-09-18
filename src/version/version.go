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
		vstr = fmt.Sprintf("%s+dev-%d", vstr, v.BuildDate.Unix())
	}
	return vstr
}

// Compare returns true if the parsed version ver is newer than the struct version v
func (v *Version) Compare(ver string) (bool, error) {
	const op = "version.(Version).Compare"
	x, dev, _ := strings.Cut(ver, "+")
	vstr := strings.Split(strings.TrimSpace(x), ".")
	if len(vstr) != 3 {
		return false, errors.New(errors.Internal, op, fmt.Sprintf("incorrent number of version sub components, should be exactly 3, got %d (%s)", len(vstr), x))
	}

	var err error
	parsed := &Version{}
	if parsed.Major, err = strconv.ParseInt(vstr[0], 10, 64); err != nil {
		return false, errors.Wrap(op, err, "failed to parse major version component")
	}
	if parsed.Minor, err = strconv.ParseInt(vstr[1], 10, 64); err != nil {
		return false, errors.Wrap(op, err, "failed to parse minor version component")
	}
	if parsed.Patch, err = strconv.ParseInt(vstr[2], 10, 64); err != nil {
		return false, errors.Wrap(op, err, "failed to parse patch version component")
	}

	if dev != "" {
		vstr = strings.Split(dev, "-")
		if len(vstr) != 2 {
			return false, errors.New(errors.Internal, op, fmt.Sprintf("incorrent number of dev sub components, should be exactly 2, got %d (%s)", len(vstr), dev))
		}
		if vstr[0] != "dev" {
			return false, errors.New(errors.Internal, op, fmt.Sprintf("invalid version dev string, should be formatted \"dev-{unix-timestamp}\", but was %s", dev))
		}
		var buildtime int64
		if buildtime, err = strconv.ParseInt(vstr[1], 10, 64); err != nil {
			return false, errors.Wrap(op, err, "failed to parse dev string timestamp")
		}
		parsed.BuildDate = time.Unix(buildtime, 0)
	}

	switch {
	case parsed.Major < v.Major:
		return false, nil
	case parsed.Major > v.Major:
		return true, nil
	// state: parsed.Major == v.Major
	case parsed.Minor < v.Minor:
		return false, nil
	case parsed.Minor > v.Minor:
		return true, nil
	// state: parsed.Minor == v.Minor
	case parsed.Patch < v.Patch:
		return false, nil
	case parsed.Patch > v.Patch:
		return true, nil
	// state: parsed.Patch == v.Patch
	case parsed.BuildDate.Unix() > v.BuildDate.Unix():
		return true, nil
	// state: parsed.BuildDate <= v.BuildDate
	default:
		return false, nil
	}
}

func New() (*Version, error) {
	const op = "version.New"
	if VersionStr == "" {
		// currently being run via `go run`
		vb, err := os.ReadFile("VERSION")
		if err != nil {
			// set it to 0, so it doesn't crash
			VersionStr = "0.0.0"
		} else {
			VersionStr = string(vb)
		}
	}
	vstr := strings.Split(strings.TrimSpace(VersionStr), ".")
	if len(vstr) != 3 {
		return nil, errors.New(errors.Internal, op, fmt.Sprintf("incorrent number of version sub components, should be exactly 3, got %d (%s)", len(vstr), VersionStr))
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
