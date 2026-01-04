// Package version provides build version info.
package version

const (
	// Version is the release version (RC frozen)
	Version = "1.0.0-rc1"

	// Commit is the git commit hash (optional)
	Commit = ""

	// BuildDate is the build timestamp
	BuildDate = ""
)

// Info holds version metadata
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit,omitempty"`
	BuildDate string `json:"build_date,omitempty"`
}

// Get returns current version info
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
	}
}
