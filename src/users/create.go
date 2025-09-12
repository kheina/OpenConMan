package users

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/argon2"
	"google.golang.org/protobuf/proto"

	"github.com/kheina/openconman/src/cli"
	pb "github.com/kheina/openconman/src/gen/pbs/auth/argon2"
	cfg "github.com/kheina/openconman/src/users/config"
	"github.com/kheina/openconman/src/util"
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

	scopes []string
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
		cli.NewByteStringArg("password", "The password for the new user. Please make it secure. tip: use env://<env var> to load an environment variable", &c.password, "--password", "-p"),
		cli.NewStringArg("work dir", "Creates the user in the provided working directory, in a format expected by a conman server running in the same directory.", &c.workDir, "--work-dir", "-w"),
		cli.NewUintArg("time", "How long a hash should take to create. Not in seconds, it's more metaphysical than that. 1-3 is good, depending on your hardware.", &c.time, "--time", "-t"),
		cli.NewUintArg("memory", "How much memory (RAM) to use, in megabytes (MiB), to use for each hash. This number should be fairly large for your device.", &c.memory, "--memory", "-m"),
		cli.NewUintArg("threads", "How many threads to use per hash. Dependent on hardware, but should probably be around a quarter of the available threads.", &c.threads, "--threads", "-T"),
		cli.NewUintArg("salt length", "The length, in bytes, of the salt to be used for the hash. If you don't know what this means, make it the same as the hash length.", &c.saltLength, "--salt", "-S"),
		cli.NewUintArg("hash length", "The length, in bytes, of the hash to be generated. 32 is good.", &c.hashLength, "--hash", "-H"),

		cli.NewStringSliceArg("scopes", "The scopes that this user has access to. Should be formatted as scope:scope:...:action", &c.scopes, "--scope", "-s"),
	}
}

func (c *createCommand) Run() error {
	const op = "users.(createCommand).Run"
	switch {
	case c.name == "":
		return fmt.Errorf("%s: user name missing", op)
	case c.password == nil:
		return fmt.Errorf("%s: user password missing", op)
	case c.time == 0:
		return fmt.Errorf("%s: time parameter cannot be 0", op)
	case c.memory == 0:
		return fmt.Errorf("%s: memory parameter cannot be 0", op)
	case c.threads == 0:
		return fmt.Errorf("%s: threads parameter cannot be 0", op)
	case c.saltLength == 0:
		return fmt.Errorf("%s: salt length cannot be 0", op)
	case c.hashLength == 0:
		return fmt.Errorf("%s: hash length cannot be 0", op)
	}
	config := &pb.Argon2Config{
		Time:        uint32(c.time),
		Memory:      uint32(c.memory),
		Parallelism: uint32(c.threads),
		HashLength:  uint32(c.hashLength),
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

	user := cfg.New(c.name)
	user.Password.Hash = base64.RawURLEncoding.AppendEncode([]byte{}, out)
	for _, s := range c.scopes {
		if err := user.AddScope(s); err != nil {
			return fmt.Errorf("%s: failed to add scope to user: %w", op, err)
		}
	}
	conf, err := user.Marshal()
	if err != nil {
		return fmt.Errorf("%s: failed to marshal user config: %w", op, err)
	}
	if c.workDir == "" {
		fmt.Println(string(conf))
	} else {
		path := strings.TrimRight(c.workDir, "/") + "/users"
		if !util.PathExists(path) {
			if err = os.Mkdir(path, 0644); err != nil {
				return fmt.Errorf("%s: failed to create users directory: %w", op, err)
			}
		}
		filename := fmt.Sprintf("%s/%s.conf", path, strings.ToLower(c.name))
		if err = os.WriteFile(filename, conf, 0644); err != nil {
			return fmt.Errorf("%s: failed to write user config to file: %w", op, err)
		}
		fmt.Printf("wrote userfile to %s\n", filename)
	}
	return nil
}
