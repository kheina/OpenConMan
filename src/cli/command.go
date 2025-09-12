package cli

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/kheina/openconman/src/util"
	"github.com/mitchellh/go-wordwrap"
)

type command struct {
	key    string
	cli    *CLI
	parent *command

	cmd      CommandFactory
	commands map[string]*command
}

type Command interface {
	mustEmbedUnimplementedCommand()

	// Synopsis returns a short one-line descriptor of the command
	Synopsis() string
	// Description optionally returns a longer description of the commend. Only
	// shown in individual command/subcommand menus
	Description() string
	// Args returns the command line args to be parsed for this command
	Args() Args
	// Run executes the given command based upon the pre-parsed sub-command args
	Run() error
}

var UnimplementedCommandError = fmt.Errorf("command not implemented")

type UnimplementedCommand struct{}

func (UnimplementedCommand) mustEmbedUnimplementedCommand() {}
func (c *UnimplementedCommand) Synopsis() string {
	return "<no description available>"
}
func (UnimplementedCommand) Description() string {
	return ""
}
func (UnimplementedCommand) Args() Args {
	return Args{}
}

func (c *UnimplementedCommand) Run() error {
	return UnimplementedCommandError
}

type CommandFactory func() (Command, error)

type CLI struct {
	name      string
	commands  map[string]*command
	args      []*arg
	envPrefix string

	help_ bool
}

// New returns a new CLI with the help command already attached
func New(name string, opt ...Option) *CLI {
	opts := getOpts(opt...)
	cli := &CLI{
		name:      name,
		commands:  make(map[string]*command),
		envPrefix: opts.withEnvPrefix,
	}
	cli.NewGlobalArg(NewBoolArg("help", "Prints this help message", &cli.help_, "--help", "-h"))
	return cli
}

// NewGlobalArg creates an argument for the entire cli, and will be parsed for
// any and all commands and subcommands within the cli
func (i *CLI) NewGlobalArg(args ...*arg) {
	i.args = append(i.args, args...)
}

// NewCommand add a command to the cli. arguments should be added to the returned
// command directly
func (i *CLI) NewCommand(key string, cmd CommandFactory) *command {
	if _, ok := i.commands[key]; ok {
		panic(fmt.Sprintf("command \"%s\" already exists", key))
	}
	c := &command{
		key: key,
		cli: i,
		cmd: cmd,
	}
	i.commands[key] = c
	return c
}

func (i *CLI) envVariable(arg *arg) string {
	switch {
	case i.envPrefix == "":
		return ""
	case arg.key == "":
		return ""
	}
	ws := regexp.MustCompile(`\s+`)
	return strings.ToUpper(i.envPrefix + "_" + ws.ReplaceAllString(arg.key, "_"))
}

// NewSubCommand adds a nested command within a pre-existing command. these are
// called by providing all parent command keys within the cli, followed by the
// subcommand key
func (c *command) NewSubCommand(key string, cmd CommandFactory) *command {
	if c.commands == nil {
		c.commands = make(map[string]*command)
	}
	if _, ok := c.commands[key]; ok {
		panic("subcommand already exists")
	}
	s := &command{
		key:    key,
		cli:    c.cli,
		parent: c,
		cmd:    cmd,
	}
	c.commands[key] = s
	return s
}

func printArgs(args Args) {
	for _, arg := range args {
		v := arg.Value()
		t := arg.Type()
		switch {
		case util.IsNil(v):
			fallthrough
		case t == argTypeFlag:
			fmt.Printf(
				"\n\n\t%s\n\t\t%s",
				fmt.Sprintf("%s <%s>", strings.Join(arg.flags, ", "), arg.Type()),
				strings.ReplaceAll(wordwrap.WrapString(arg.desc, 56), "\n", "\n\t\t"),
			)
		default:
			fmt.Printf(
				"\n\n\t%-40s [default: %v]\n\t\t%s",
				fmt.Sprintf("%s <%s>", strings.Join(arg.flags, ", "), arg.Type()),
				v,
				strings.ReplaceAll(wordwrap.WrapString(arg.desc, 56), "\n", "\n\t\t"),
			)
		}
	}
}

func (c *command) help() int {
	const op = "cli.(command).help"
	cmd, err := c.cmd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: cli error: %s\n", op, err.Error())
		return 1
	}

	chain := []*command{}
	cc := c.parent
	for cc != nil {
		chain = append(chain, cc)
		cc = cc.parent
	}
	slices.Reverse(chain)
	fmt.Printf("Usage: %s ", c.cli.name)
	for _, cc := range chain {
		fmt.Printf("%s ", cc.key)
	}
	if c.commands == nil {
		fmt.Printf("%s [options]\n\n\t$ %s ", c.key, c.cli.name)
		for _, cc := range chain {
			fmt.Printf("%s ", cc.key)
		}
		fmt.Printf("%s", c.key)
		for i, arg := range cmd.Args() {
			if util.IsNil(arg.Value()) {
				continue
			}
			if i >= 3 {
				break
			}
			switch arg.Type() {
			case argTypeFlag:
				fmt.Printf(" %s", arg.flags[0])
			default:
				fmt.Printf(" %s %v", arg.flags[0], arg.Value())
			}
		}
		if desc := cmd.Description(); desc != "" {
			fmt.Printf("\n\n%s ", wordwrap.WrapString(desc, 72))
		}
	} else {
		fmt.Printf("%s <subcommand> [options]", c.key)
		if desc := cmd.Description(); desc != "" {
			fmt.Printf("\n\n%s ", wordwrap.WrapString(desc, 72))
		}
		fmt.Printf("\n\nAvailable Subcommands:")

		maxlen := 0
		keys := make([]string, 0, len(c.commands))
		for n := range c.commands {
			maxlen = max(maxlen, len(n))
			keys = append(keys, n)
		}
		slices.Sort(keys)
		for _, n := range keys {
			cmd, err := c.commands[n].cmd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: error while producing command: %s\n", op, err.Error())
				return 1
			}
			desc := cmd.Synopsis()
			lim := 64 - maxlen
			if len(desc) > lim {
				buf := "\n\t   " + strings.Repeat(" ", maxlen)
				if i := strings.LastIndexFunc(desc[:lim], unicode.IsSpace); i > 0 {
					desc = desc[:i] + buf + strings.ReplaceAll(wordwrap.WrapString(desc[i+1:], uint(lim)), "\n", buf)
				} else {
					desc = strings.ReplaceAll(wordwrap.WrapString(desc, uint(lim)), "\n", buf)
				}
			}
			fmt.Printf("\n\n\t%s%s - %s", n, strings.Repeat(" ", maxlen-len(n)), desc)
		}
	}
	args := cmd.Args()
	if len(args) != 0 {
		fmt.Printf("\n\nAvailable Options:")
		printArgs(cmd.Args())
	}
	fmt.Printf("\n\nGlobal Options:")
	printArgs(c.cli.args)

	fmt.Print("\n\n")
	return 0
}

func (i *CLI) help(msg string) int {
	const op = "cli.(CLI).help"
	if msg == "" {
		msg = fmt.Sprintf("Usage: %s <command> [options]", i.name)
	}
	fmt.Printf("%s\n\nAvailable Commands:\n\n", msg)
	maxlen := 0
	keys := make([]string, 0, len(i.commands))
	for n := range i.commands {
		maxlen = max(maxlen, len(n))
		keys = append(keys, n)
	}
	slices.Sort(keys)
	for _, n := range keys {
		cmd, err := i.commands[n].cmd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: error while producing command: %s\n", op, err.Error())
			return 1
		}
		fmt.Printf("\t%s%s - %s\n\n", n, strings.Repeat(" ", maxlen-len(n)), cmd.Synopsis())
	}
	fmt.Printf("Global Options:")
	printArgs(i.args)

	fmt.Print("\n\n")
	return 0
}

// Run runs the CLI with the provided commands and arguments. args should be
// os.Args[1:] unless some other form of preprocessing has been done to them
func (i *CLI) Run(args []string) int {
	if len(args) == 0 || slices.Contains([]string{"-h", "--help"}, args[0]) {
		return i.help("")
	}
	c, ok := i.commands[args[0]]
	if !ok {
		return i.help(fmt.Sprintf("Unknown command: %s", args[0]))
	}
	args = args[1:]
	var cmd Command
	var err error
	for {
		if len(args) == 0 {
			cmd, err = c.cmd()
			break
		}
		if cc, ok := c.commands[args[0]]; !ok {
			cmd, err = c.cmd()
			break
		} else {
			c = cc
			args = args[1:]
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "cli error: %s\n", err.Error())
		return 1
	}
	if err := i.ParseArgs(args, append(i.args, cmd.Args()...)...); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 1
	}
	if i.help_ {
		return c.help()
	}
	if err = cmd.Run(); err != nil {
		if errors.Is(err, UnimplementedCommandError) {
			return c.help()
		}
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	return 0
}
