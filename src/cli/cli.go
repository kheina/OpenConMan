package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func newArg(op, key, desc string, dest any, flags []string) *arg {
	if len(flags) == 0 {
		panic(fmt.Sprintf("%s: failed to create cli arg, at least one flag required", op))
	}
	return &arg{
		desc:  desc,
		dest:  dest,
		flags: flags,
		key:   key,
	}
}

type arg struct {
	// Description
	desc string
	// flags used to parse the command line strings
	flags []string
	// where the value is stored afterwards
	dest any
	// name of the arg, for helpful errors
	key string
}

type Args []*arg

func (a *arg) Value() any {
	// there's GOT to be a better way to do this with reflection
	switch dest := a.dest.(type) {
	case *string:
		if *dest == "" {
			return nil
		}
		return *dest
	case *int:
		return *dest
	case *uint:
		return *dest
	case *bool:
		return *dest
	default:
		return nil
	}
}

func (a *arg) SetValue(v string) error {
	const op = "cli.(arg).SetValue"
	switch dest := a.dest.(type) {
	case *bool:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("%s: could not parse bool value for %s: %s", op, a.key, v)
		}
		*dest = b
	case *string:
		*dest = v
	case *int:
		i, err := strconv.ParseInt(v, 10, 0)
		if err != nil {
			return fmt.Errorf("%s: could not parse int value for %s: %s", op, a.key, v)
		}
		*dest = int(i)
	case *uint:
		u, err := strconv.ParseUint(v, 10, 0)
		if err != nil {
			return fmt.Errorf("%s: could not parse uint value for %s: %s", op, a.key, v)
		}
		*dest = uint(u)
	case *[]byte:
		if dest == nil {
			// this is to avoid a nil pointer dereference
			// we need to populate the pointer first
			dest = &[]byte{}
		}
		*dest = []byte(v)
	case *[]string:
		if dest == nil {
			// this is to avoid a nil pointer dereference
			// we need to populate the pointer first
			dest = &[]string{}
		}
		*dest = append(*dest, v)
	default:
		return fmt.Errorf("%s: found unexpected destination type for %s: %T", op, a.key, dest)
	}
	return nil
}

type argType int

const (
	argTypeUnknown argType = iota
	argTypeString
	argTypeInt
	argTypeUint
	argTypeFlag
	argTypeByteString
	argTypeStringSlice
)

func (t argType) String() string {
	switch t {
	case argTypeByteString:
		return "string"
	case argTypeString:
		return "string"
	case argTypeInt:
		return "int"
	case argTypeUint:
		return "uint"
	case argTypeFlag:
		return "flag"
	case argTypeStringSlice:
		return "string"
	default:
		return "unknown"
	}
}

func (a *arg) Type() argType {
	// there's GOT to be a better way to do this with reflection
	switch a.dest.(type) {
	case *string:
		return argTypeString
	case *int:
		return argTypeInt
	case *uint:
		return argTypeUint
	case *bool:
		return argTypeFlag
	case *[]byte:
		return argTypeByteString
	case *[]string:
		return argTypeStringSlice
	default:
		return argTypeUnknown
	}
}

// NewStringArg returns a new arg with the provided values to be parsed by
// ParseArgs. panics if you do something stupid like not provide any flags.
// flags should be formatted WITH preceding dashes.
func NewStringArg(key, description string, destination *string, flags ...string) *arg {
	const op = "cli.NewStringArg"
	return newArg(op, key, description, destination, flags)
}

// NewByteStringArg returns a new arg with the provided values to be parsed by
// ParseArgs. panics if you do something stupid like not provide any flags.
// flags should be formatted WITH preceding dashes.
func NewByteStringArg(key, description string, destination *[]byte, flags ...string) *arg {
	const op = "cli.NewByteStringArg"
	return newArg(op, key, description, destination, flags)
}

// NewUintArg returns a new arg with the provided values to be parsed by
// ParseArgs. panics if you do something stupid like not provide any flags.
// flags should be formatted WITH preceding dashes.
func NewUintArg(key, description string, destination *uint, flags ...string) *arg {
	const op = "cli.NewUintArg"
	return newArg(op, key, description, destination, flags)
}

// NewIntArg returns a new arg with the provided values to be parsed by
// ParseArgs. panics if you do something stupid like not provide any flags.
// flags should be formatted WITH preceding dashes.
func NewIntArg(key, description string, destination *int, flags ...string) *arg {
	const op = "cli.NewIntArg"
	return newArg(op, key, description, destination, flags)
}

// NewBoolArg returns a new arg with the provided values to be parsed by
// ParseArgs. panics if you do something stupid like not provide any flags.
// flags should be formatted WITH preceding dashes.
func NewBoolArg(key, description string, destination *bool, flags ...string) *arg {
	const op = "cli.NewBoolArg"
	return newArg(op, key, description, destination, flags)
}

// NewStringSliceArg returns a new arg with the provided values to be parsed by
// ParseArgs. panics if you do something stupid like not provide any flags. flags
// should be formatted WITH preceding dashes. NOTE: initial array should be non-nil
// and empty
func NewStringSliceArg(key, description string, destination *[]string, flags ...string) *arg {
	const op = "cli.NewStringSliceArg"
	return newArg(op, key, description, destination, flags)
}

// ParseArgs parses all of the passed command line arguments using the args
// created by other exported cli functions
func (i *CLI) ParseArgs(args []string, a ...*arg) error {
	const op = "cli.ParseArgs"
	found := make(map[string]bool)
	mapper := make(map[string]*arg)
	for _, arg := range a {
		if _, ok := found[arg.key]; ok {
			return fmt.Errorf("%s: duplicate arg keys given: %s", op, arg.key)
		}
		found[arg.key] = false
		for _, n := range arg.flags {
			if _, ok := mapper[n]; ok {
				return fmt.Errorf("%s: duplicate arg flags given: %s", op, n)
			}
			mapper[n] = arg
		}
	}
	// pull possible values from the env first
	for _, arg := range a {
		envVar := i.envVariable(arg)
		if envVar == "" {
			continue
		}
		if v, ok := os.LookupEnv(envVar); ok {
			if err := arg.SetValue(v); err != nil {
				return err
			}
		}
	}
	// then parse all passed cli args
	for i := 0; i < len(args); i++ {
		arg, v, _ := strings.Cut(args[i], "=")
		value, ok := mapper[arg]

		if !ok {
			return fmt.Errorf("%s: unexpected argument received: %s", op, arg)
		}
		if f, ok := found[value.key]; ok && f {
			return fmt.Errorf("%s: duplicate argument received: %s", op, value.key)
		}
		switch value.dest.(type) {
		case *[]string:
		default:
			found[value.key] = true
		}

		// gotta parse bool early since it can be a toggle and not a passed value
		if v == "" {
			if dest, ok := value.dest.(*bool); ok {
				*dest = !*dest
				continue
			} else {
				i++
				if i == len(args) {
					return fmt.Errorf("%s: received \"%s\" arg but no value provided", op, value.key)
				}
				v = args[i]
			}
		}

		// if the value of the passed arg indicates it's an env variable or file, swap the contents
		switch {
		case strings.HasPrefix(v, "env://"):
			v, _ = os.LookupEnv(v[6:])
		case strings.HasPrefix(v, "file://"):
			file := v[7:]
			vb, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("%s: failed to read file for %s: %s", op, value.key, file)
			}
			v = string(vb)
		}

		if err := value.SetValue(v); err != nil {
			return err
		}
	}
	return nil
}

// TODO: implement autocomplete?
