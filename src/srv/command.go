package srv

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/go-hclog"
)

type Command struct{}

func (c *Command) Help(args []string) string {
	return `Usage: conman serve [options]

	$ conman serve --port 443 --host 0.0.0.0

Available Options:

	--port, -P <uint>
		Runs openconman on the provided port number

	--host, -H <ip>
		Runs openconman on the provided ip address

	--grpc-port, -GP <uint>
		Runs the openconman gRPC api on the provided port number

	--certfile <filename>    [NOT IMPLEMENTED]
		SSL certificate file to be used for https connections.
		--keyfile must also be provided to encrypt connections

	--keyfile <filename>     [NOT IMPLEMENTED]
		SSL key file to be used for https connections.
		--certfile must also be provided to encrypt connections

	-v, -vv, -vvv
		Sets the verbosity of the current command where -v is
		somewhat verbose and -vvv is very verbose`
}

func (c *Command) Description() string {
	return "Runs the openconman server"
}

type srvArgs struct {
	host     string
	port     uint
	grpcPort uint
	logLevel hclog.Level
}

func (c *Command) parseArgs(args []string) (*srvArgs, error) {
	const op = "srv.(Command).parseArgs"
	a := &srvArgs{
		host:     "0.0.0.0",
		port:     5050,
		grpcPort: 4949,
		logLevel: hclog.Info,
	}
	for i := range len(args) {
		switch arg := args[i]; arg {
		case "-H", "--host":
			i++
			if i == len(args) {
				return nil, fmt.Errorf("%s: received \"host\" arg but no value provided", op)
			}
			a.host = args[i]
		case "-P", "--port":
			i++
			if i == len(args) {
				return nil, fmt.Errorf("%s: received \"port\" arg but no value provided", op)
			}
			p, err := strconv.ParseUint(args[i], 10, 0)
			if err != nil {
				return nil, fmt.Errorf("%s: could not parse port value: %s", op, args[i])
			}
			a.port = uint(p)
		case "-GP", "--grpc-port":
			i++
			if i == len(args) {
				return nil, fmt.Errorf("%s: received \"grpc-port\" arg but no value provided", op)
			}
			p, err := strconv.ParseUint(args[i], 10, 0)
			if err != nil {
				return nil, fmt.Errorf("%s: could not parse port value: %s", op, args[i])
			}
			a.grpcPort = uint(p)
		case "-v":
			a.logLevel = hclog.Info
		case "-vv":
			a.logLevel = hclog.Debug
		case "-vvv":
			a.logLevel = hclog.Trace
		default:
			return nil, fmt.Errorf("%s: unknown option: %s", op, args[i])
		}
	}
	return a, nil
}

func (c *Command) Run(args []string) error {
	const op = "srv.(Command).Run"
	pargs, err := c.parseArgs(args)
	if err != nil {
		return fmt.Errorf("%s: failed to parse args: %w", op, err)
	}
	r, err := NewRouter(pargs.host, pargs.port, pargs.grpcPort, pargs.logLevel)
	if err != nil {
		return err
	}
	if err = r.Serve(); err != nil {
		return fmt.Errorf("%s: failed to run server: %w", op, err)
	}
	return nil
}
