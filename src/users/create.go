package users

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
	"google.golang.org/protobuf/proto"

	"github.com/kheina/openconman/src/cli"
	pb "github.com/kheina/openconman/src/gen/pbs/auth/argon2"
)

type createCommand struct {
	cli.UnimplementedCommand

	name     string
	password []byte
	workDir  string

	// config
	// config   *pb.Argon2Config
	time       uint
	memory     uint
	threads    uint
	saltLength uint
	hashLength uint
}

func newCreateCommand() *createCommand {
	return &createCommand{
		time:       3,
		memory:     256,
		threads:    4,
		saltLength: 32,
		hashLength: 32,
	}
}

func (c *createCommand) Synopsis() string {
	return "Creates a userfile that can be used by the conman server"
}

func (c *createCommand) Args() cli.Args {
	return cli.Args{
		cli.NewStringArg("name", "The name to be given to the new user. Used for logins and identification. Must be unique within a directory.", &c.name, "--name", "-n"),
		cli.NewStringArg("password", "The password for the new user. Please make it secure. tip: use env://<env var> to load an environment variable", &c.name, "--password", "-p"),
		cli.NewStringArg("work dir", "Creates the user in the provided working directory, in a format expected by a conman server running in the same directory.", &c.workDir, "--work-dir", "-w"),
		cli.NewUintArg("time", "How long a hash should take to create. Not in seconds, it's more metaphysical than that. 1-3 is good, depending on your hardware.", &c.time, "--time", "-t"),
		cli.NewUintArg("memory", "How much memory (RAM) to use, in megabytes (MiB), to use for each hash. This number should be fairly large for your device.", &c.memory, "--memory", "-m"),
		cli.NewUintArg("threads", "How many threads to use per hash. Dependent on hardware, but should probably be around a quarter of the available threads.", &c.threads, "--threads", "-T"),
		cli.NewUintArg("salt length", "The length, in bytes, of the salt to be used for the hash. If you don't know what this means, make it the same has the hash length.", &c.saltLength, "--salt", "-s"),
		cli.NewUintArg("hash length", "The length, in bytes, of the hash to be generated. 32 is good.", &c.saltLength, "--hash", "-H"),
	}
}

func (c *createCommand) Run() error {
	const op = "users.(createCommand).Run"
	switch {
	case c.name == "":
	}
	config := &pb.Argon2Config{
		Time:        uint32(c.time),
		Memory:      uint32(c.memory),
		Parallelism: uint32(c.threads),
		// SaltLength:  uint32(c.saltLength),
		HashLength: uint32(c.hashLength),
	}
	salt := make([]byte, c.saltLength)
	if _, err := rand.Reader.Read(salt); err != nil {
		return fmt.Errorf("%s: failed to read salt: %w", op, err)
	}
	hash := &pb.Argon2Hash{
		Config: config,
		Salt:   salt,
	}
	hash.Hash = argon2.IDKey(c.password, salt, config.Time, config.Memory*1024, uint8(config.Parallelism), config.HashLength)

	out, err := proto.Marshal(hash)
	if err != nil {
		return fmt.Errorf("%s: failed to marshal password hash: %w", op, err)
	}

	enc := base64.RawURLEncoding.EncodeToString(out)
	fmt.Printf("hash: %s (%d/%d)\n", enc, len(enc), len(out))
	return nil
}
