package errors

// getOpts - iterate the inbound Options and return a struct
func getOpts(opt ...Option) options {
	opts := options{
		withStatusCode: Internal,
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
	withErrorCode  string
	withStatusCode Status
}

// WithErrorCode provides a non-standard error code in place of the name of the error
//
// ex: OriginNotAllowed for a Bad Request error response
func WithErrorCode(c string) Option {
	return func(o *options) {
		o.withErrorCode = c
	}
}

// WithStatusCode assigns the specified status code during errors.Wrap. Does not override an underlying ApiError's status code.
func WithStatusCode(s Status) Option {
	return func(o *options) {
		o.withStatusCode = s
	}
}
