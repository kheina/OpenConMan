package version

import (
	"fmt"
)

type Command struct{}

func (c *Command) Help(args []string) string {
	return `Usage: conman version [options]

	$ conman version

Available Options:

	--raw, -R
		returns only the version number in the format v0.0.0+dev`
}

func (c *Command) Description() string {
	return "Prints the current version info"
}

const timeFormat = "2006-01-02T15:04:05Z"

type vArgs struct {
	raw bool
}

func (c *Command) parseArgs(args []string) (*vArgs, error) {
	a := &vArgs{
		raw: false,
	}
	for i := range len(args) {
		switch arg := args[i]; arg {
		case "--raw", "-R":
			a.raw = true
		}
	}
	return a, nil
}

func (c *Command) Run(args []string) error {
	v, err := New()
	if err != nil {
		return err
	}
	pargs, _ := c.parseArgs(args)
	if pargs.raw {
		_, err = fmt.Printf("v%s", v)
	} else {
		_, err = fmt.Printf("version number:    %s\ncommit hash:       %s\nbuild date:        %s\n\n", v, v.Rev, v.BuildDate.UTC().Format(timeFormat))
	}
	return err
}
