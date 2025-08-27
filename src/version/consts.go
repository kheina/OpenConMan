package version

var (
	// these are all set at compile time using `make build`
	Commit     string
	Branch     string
	VersionStr string
	Timestamp  string
)
