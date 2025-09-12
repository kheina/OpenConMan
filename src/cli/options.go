package cli

// getOpts - iterate the inbound Options and return a struct
func getOpts(opt ...Option) options {
	opts := options{}
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
	withEnvPrefix string
}

// WithEnvPrefix allows retrieving cli arg values from environment variables
// automatically where the env variable is uppercase(prefix)+uppercase(arg.key).
// all whitespace will be replaced with underscores such that if the argument
// key is "user name" the env var becomes "PREFIX_USER_NAME". note that any args
// passed in directly through the commandline will override env variables
func WithEnvPrefix(prefix string) Option {
	return func(o *options) {
		o.withEnvPrefix = prefix
	}
}
