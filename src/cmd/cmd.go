package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

func help(msg string) int {
	const op = "cmd.help"
	if msg == "" {
		msg = "Usage: conman <command> [args]"
	}
	fmt.Printf("%s\n\nAvailable commands:\n\n", msg)
	maxlen := 0
	keys := make([]string, 0, len(Commands))
	for n := range Commands {
		maxlen = max(maxlen, len(n))
		keys = append(keys, n)
	}
	slices.Sort(keys)
	for _, n := range keys {
		cmd, err := Commands[n]()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: error while producing command: %s\n", op, err.Error())
			return 1
		}
		fmt.Printf("\t%s%s - %s\n\n", n, strings.Repeat(" ", maxlen-len(n)), cmd.Description())
	}

	return 0
}

func Run(args []string) int {
	initCommands()
	if len(args) == 0 || slices.Contains([]string{"-h", "--help"}, args[0]) {
		return help("")
	}
	c, ok := Commands[args[0]]
	if !ok {
		return help(fmt.Sprintf("Unknown command: %s", args[0]))
	}
	cmd, err := c()
	switch {
	case err != nil:
		fmt.Fprintf(os.Stderr, "cli error: %s\n", err.Error())
		return 1
	case len(args) >= 2 && slices.ContainsFunc(args, func(e string) bool { return e == "-h" || e == "--help" }):
		fmt.Printf("%s\n\n", strings.TrimSuffix(cmd.Help(args[1:]), "\n"))
		return 0
	}
	if err = cmd.Run(args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	return 0
}
