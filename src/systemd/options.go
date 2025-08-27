package systemd

import "github.com/hashicorp/go-hclog"

// getOpts - iterate the inbound Options and return a struct
func getOpts(opt ...Option) options {
	opts := options{
		withLogger: hclog.NewNullLogger(),
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
	withLogger hclog.Logger
}

// WithLogger allows passing an optional logger to the systemd server
func WithLogger(l hclog.Logger) Option {
	return func(o *options) {
		o.withLogger = l
	}
}
