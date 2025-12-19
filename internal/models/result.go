package models

import "time"

// ScanResult aggregates all categorized PRs and metadata from a scan.
// This is the final output of the scan pipeline:
// 1. Scanner finds repos
// 2. GitHub client fetches PRs
// 3. Stack detector builds trees
// 4. Categorizer sorts PRs
// 5. All assembled into ScanResult
type ScanResult struct {
	// Categorized PRs
	MyPRs            []*PR `json:"my_prs"`
	NeedsMyAttention []*PR `json:"needs_my_attention"`
	TeamPRs          []*PR `json:"team_prs"`
	OtherPRs         []*PR `json:"other_prs"`

	// Repository information
	ReposWithPRs    []*Repository `json:"repos_with_prs"`
	ReposWithoutPRs []*Repository `json:"repos_without_prs"`
	ReposWithErrors []*Repository `json:"repos_with_errors"`

	// Stack information (keyed by repo full name, e.g., "org/repo")
	Stacks map[string]*Stack `json:"stacks"`

	// Metadata
	TotalReposScanned int           `json:"total_repos_scanned"`
	TotalPRsFound     int           `json:"total_prs_found"`
	ScanDuration      time.Duration `json:"scan_duration_ns"`
	Username          string        `json:"username"`
}

// NewScanResult creates a new ScanResult with all slices and maps initialized.
func NewScanResult() *ScanResult {
	return &ScanResult{
		MyPRs:            make([]*PR, 0),
		NeedsMyAttention: make([]*PR, 0),
		TeamPRs:          make([]*PR, 0),
		OtherPRs:         make([]*PR, 0),
		ReposWithPRs:     make([]*Repository, 0),
		ReposWithoutPRs:  make([]*Repository, 0),
		ReposWithErrors:  make([]*Repository, 0),
		Stacks:           make(map[string]*Stack),
	}
}

// TotalPRs returns the total count of all categorized PRs.
func (r *ScanResult) TotalPRs() int {
	return len(r.MyPRs) + len(r.NeedsMyAttention) + len(r.TeamPRs) + len(r.OtherPRs)
}

// HasPRs returns true if there are any PRs in any category.
func (r *ScanResult) HasPRs() bool {
	return r.TotalPRs() > 0
}

// HasErrors returns true if any repositories had scan errors.
func (r *ScanResult) HasErrors() bool {
	return len(r.ReposWithErrors) > 0
}

// TotalRepos returns the total count of all repositories (with/without PRs, with errors).
func (r *ScanResult) TotalRepos() int {
	return len(r.ReposWithPRs) + len(r.ReposWithoutPRs) + len(r.ReposWithErrors)
}

// ScanDurationString returns a human-readable duration string.
func (r *ScanResult) ScanDurationString() string {
	if r.ScanDuration < time.Second {
		return r.ScanDuration.Round(time.Millisecond).String()
	}
	return r.ScanDuration.Round(time.Second).String()
}
