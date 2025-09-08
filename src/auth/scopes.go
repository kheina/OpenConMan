package auth

type Scope uint

// Defines repository scopes like systemd, networking, and others.
const (
	Systemd Scope = iota
	Containers
)

type Action uint

// Classic CRUDL actions.
const (
	Create Action = iota
	Read
	Update
	Delete
	List
)
