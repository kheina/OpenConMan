package certificate

import (
	"fmt"
	"os"

	"github.com/kheina/openconman/src/cli"
	"github.com/kheina/openconman/src/srv/certs"
	"github.com/kheina/openconman/src/util"
)

type Command struct {
	cli.UnimplementedCommand

	certfile string
	keyfile  string
	force    bool
}

func (c *Command) Synopsis() string {
	return "Generates a TLS certificate"
}

func (c *Command) Args() cli.Args {
	return cli.Args{
		cli.NewStringArg("certfile", "The file to write the certificate to", &c.certfile, "--certfile", "-c"),
		cli.NewStringArg("keyfile", "The file to write the certificate's private key to", &c.keyfile, "--keyfile", "-k"),
		cli.NewBoolArg("force", "Forces the command to overwrite any existing files", &c.force, "--force", "-f"),
	}
}

func (c *Command) Run() error {
	const op = "certificate.(Command).Run"
	switch {
	case c.certfile == "":
		return fmt.Errorf("%s: missing certfile", op)
	case c.keyfile == "":
		return fmt.Errorf("%s: missing keyfile", op)
	}

	if !c.force {
		switch {
		case util.PathExists(c.certfile):
			return fmt.Errorf("%s: file already exists at %s", op, c.certfile)
		case util.PathExists(c.keyfile):
			return fmt.Errorf("%s: file already exists at %s", op, c.keyfile)
		}
	}

	cert, pkey, err := certs.GenerateCertificate()
	if err != nil {
		return fmt.Errorf("%s: failed to generate certificate: %w", op, err)
	}
	if err := os.WriteFile(c.certfile, cert, 0644); err != nil {
		return fmt.Errorf("%s: failed to write certificate: %w", op, err)
	}
	if err := os.WriteFile(c.keyfile, pkey, 0644); err != nil {
		return fmt.Errorf("%s: failed to write private key: %w", op, err)
	}
	return nil
}

func RegisterCommand(i *cli.CLI) {
	i.NewCommand("gen-cert", func() (cli.Command, error) {
		return &Command{}, nil
	})
}
