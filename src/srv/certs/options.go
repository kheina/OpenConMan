package certs

// getOpts - iterate the inbound Options and return a struct
func getOpts(opt ...Option) options {
	opts := options{
		withName:         []string{},
		withCountry:      []string{"US"},
		withProvince:     nil,
		withLocality:     nil,
		withOrganization: nil,
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
	withName         []string
	withCountry      []string
	withProvince     []string
	withLocality     []string
	withOrganization []string
}

// WithName sets the TLS certificate's CommonName and DNSNames. The first name will
// be used as the certificate CommonName, all passed names will be used as DNSNames.
func WithName(n string) Option {
	return func(o *options) {
		o.withName = append(o.withName, n)
	}
}

// WithCountry sets the TLS certificate's country. Should be a two-digit country
// code. EX: US or AU
func WithCountry(c string) Option {
	return func(o *options) {
		o.withCountry = []string{c}
	}
}

// WithProvince sets the TLS certificate's state or province. Should be the full
// name of the state or province. EX: Massachusetts
func WithProvince(p string) Option {
	return func(o *options) {
		o.withProvince = []string{p}
	}
}

// WithLocality sets the TLS certificate's locality. Should be a city or town.
// EX: Boston
func WithLocality(l string) Option {
	return func(o *options) {
		o.withLocality = []string{l}
	}
}

// WithOrganization sets the TLS certificate's organization
func WithOrganization(org string) Option {
	return func(o *options) {
		o.withOrganization = []string{org}
	}
}
