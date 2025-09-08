package srv

import (
	"fmt"

	"github.com/hashicorp/go-hclog"

	"github.com/kheina/openconman/src/cli"
)

type Command struct {
	cli.UnimplementedCommand

	host     string
	port     uint
	grpcPort uint
	workDir  string
	logLevel string

	insecure bool
	cert     []byte
	key      []byte
}

func (c *Command) Synopsis() string {
	return "Runs the openconman server"
}

func (c *Command) Args() cli.Args {
	return cli.Args{
		cli.NewStringArg("host", "Runs openconman on the provided ip address", &c.host, "--host", "-H"),
		cli.NewUintArg("port", "Runs openconman on the provided port number", &c.port, "--port", "-p"),
		cli.NewUintArg("gRPC port", "Runs the openconman gRPC api on the provided port number", &c.grpcPort, "--grpc-port", "-gp"),
		cli.NewStringArg("work dir", "Sets the working directory of the server. Used for configuration and state persistence. Defaults to the current working directory.", &c.workDir, "--work-dir", "-w"),
		cli.NewStringArg("log level", "Sets the log level of the server logger. Possible values, in order of increasing verbosity: Off, Error, Warn, Info, Debug, Trace.", &c.logLevel, "--log-level", "-l"),
		cli.NewByteStringArg("cert", "TLS certificate bytes to be used for https connections. Key must also be provided to encrypt connections. tip: use file://<filename> to load a certfile directly", &c.cert, "--cert", "-c"),
		cli.NewByteStringArg("key", "TLS key bytes to be used for https connections. Certificate must also be provided to encrypt connections. tip: use file://<filename> to load a keyfile directly", &c.key, "--key", "-k"),
		cli.NewBoolArg("insecure", "Forces the server to disable TLS certificate generation when no certificate is provided.", &c.insecure, "--insecure"),
	}
}

func (c *Command) Run() error {
	const op = "srv.(Command).Run"
	logLevel := hclog.LevelFromString(c.logLevel)
	if logLevel == hclog.NoLevel {
		return fmt.Errorf("%s: invalid log level supplied: %s", op, c.logLevel)
	}
	r, err := NewRouter(c.host, c.port, c.grpcPort, c.workDir, logLevel, c.cert, c.key, c.insecure)
	if err != nil {
		return err
	}
	if err = r.Serve(); err != nil {
		return fmt.Errorf("%s: failed to run server: %w", op, err)
	}
	return nil
}

func RegisterCommand(i *cli.CLI) {
	i.NewCommand("serve", func() (cli.Command, error) {
		return &Command{
			host:     "0.0.0.0",
			port:     5050,
			grpcPort: 4949,
			workDir:  "",
			logLevel: "info",
			cert:     nil,
			key:      nil,
		}, nil
	})
}
