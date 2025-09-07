package cmd

import (
	"github.com/kheina/openconman/src/certificate"
	"github.com/kheina/openconman/src/cli"
	"github.com/kheina/openconman/src/daemon"
	"github.com/kheina/openconman/src/srv"
	"github.com/kheina/openconman/src/test"
	"github.com/kheina/openconman/src/version"
)

func initCLI() *cli.CLI {
	i := cli.NewCLI("conman")
	srv.RegisterCommand(i)
	daemon.RegisterCommand(i)
	test.RegisterCommand(i)
	version.RegisterCommand(i)
	certificate.RegisterCommand(i)
	return i
}

func Run(args []string) int {
	return initCLI().Run(args)
}
