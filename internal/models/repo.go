package models

// ScanStatus represents the result status of scanning a repository.
type ScanStatus string

const (
	ScanStatusSuccess ScanStatus = "success"
	ScanStatusNoPRs   ScanStatus = "no_prs"
	ScanStatusError   ScanStatus = "error"
	ScanStatusSkipped ScanStatus = "skipped"
)

// Repository represents a local Git repository that may have GitHub PRs.
type Repository struct {
	// Identity
	Name      string `json:"name"`       // e.g., "prt"
	Path      string `json:"path"`       // e.g., "/Users/jdoe/code/prt"
	RemoteURL string `json:"remote_url"` // e.g., "git@github.com:org/prt.git"
	Owner     string `json:"owner"`      // e.g., "org"

	// PRs associated with this repository
	PRs []*PR `json:"prs"`

	// Scan metadata
	// Note: ScanError is not JSON serialized because error interface doesn't marshal well
	ScanError  error      `json:"-"`
	ScanStatus ScanStatus `json:"scan_status"`
}

// FullName returns the repository's full name in "owner/name" format.
func (r *Repository) FullName() string {
	if r.Owner == "" {
		return r.Name
	}
	return r.Owner + "/" + r.Name
}

// HasPRs returns true if the repository has any pull requests.
func (r *Repository) HasPRs() bool {
	return len(r.PRs) > 0
}
