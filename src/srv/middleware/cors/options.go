package cors

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
)

// getOpts - iterate the inbound Options and return a struct
func getOpts(opt ...Option) options {
	opts := options{
		withLogger: hclog.NewNullLogger(),
		withOrigin: []string{
			// for now, add localhost addresses to be always allowed. this isn't the best, but I can fix it later
			"localhost",
			"127.0.0.1",
		},
	}
	for _, o := range opt {
		if o != nil {
			o(&opts)
		}
	}
	return opts
}

// Option - how Options are passed as arguments
type Option func(*options)

// options = how options are represented
type options struct {
	withLogger   hclog.Logger
	withOrigin   []string
	withProtocol []string
}

// WithLogger allows passing an optional logger to the systemd server
func WithLogger(l hclog.Logger) Option {
	return func(o *options) {
		o.withLogger = l
	}
}

// WithOrigin passes a valid origin to cors to allow requests from. Can be used
// multiple times.
func WithOrigin(s string) Option {
	switch s {
	case "0.0.0.0":
		s = "*"
	case "::":
		s = "*"
	case "[::]":
		s = "*"
	}
	return func(o *options) {
		o.withOrigin = append(o.withOrigin, s)
	}
}

// WithProtocol allows traffic from a specific origin protocol. Only http or
// https is allowed. Can be used multiple times.
func WithProtocol(p string) Option {
	switch p {
	case "http":
	case "https":
	default:
		panic(fmt.Sprintf("protocol \"%s\" not allowed", p))
	}
	return func(o *options) {
		o.withProtocol = append(o.withProtocol, p)
	}
}
