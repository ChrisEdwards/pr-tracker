// Package config handles configuration loading and validation for PRT.
package config

// GroupBy constants define how PRs are grouped in the display.
const (
	GroupByProject = "project"
	GroupByAuthor  = "author"
)

// Sort constants define the order of PRs in the display.
const (
	SortOldest = "oldest"
	SortNewest = "newest"
)

// Config holds all configuration options for PRT.
type Config struct {
	// Identity - the current user's GitHub username
	GitHubUsername string `yaml:"github_username" mapstructure:"github_username"`

	// Team - list of GitHub usernames for team highlighting
	TeamMembers []string `yaml:"team_members" mapstructure:"team_members"`

	// Repository Discovery
	SearchPaths  []string `yaml:"search_paths" mapstructure:"search_paths"`   // Where to look for repos
	IncludeRepos []string `yaml:"include_repos" mapstructure:"include_repos"` // Glob patterns (empty = all)
	ScanDepth    int      `yaml:"scan_depth" mapstructure:"scan_depth"`       // Max directory depth

	// Known Bots - accounts to exclude from team/other categorization
	Bots []string `yaml:"bots" mapstructure:"bots"`

	// Display options
	DefaultGroupBy string `yaml:"default_group_by" mapstructure:"default_group_by"` // project | author
	DefaultSort    string `yaml:"default_sort" mapstructure:"default_sort"`         // oldest | newest
	ShowBranchName bool   `yaml:"show_branch_name" mapstructure:"show_branch_name"`
	ShowIcons      bool   `yaml:"show_icons" mapstructure:"show_icons"`
	ShowOtherPRs   bool   `yaml:"show_other_prs" mapstructure:"show_other_prs"` // Show "Other PRs" section

	// Filtering options
	MaxPRAgeDays int `yaml:"max_pr_age_days" mapstructure:"max_pr_age_days"` // Hide PRs older than N days (0 = no limit)
}

// IsValidGroupBy returns true if the given value is a valid GroupBy option.
func IsValidGroupBy(v string) bool {
	return v == GroupByProject || v == GroupByAuthor
}

// IsValidSort returns true if the given value is a valid Sort option.
func IsValidSort(v string) bool {
	return v == SortOldest || v == SortNewest
}
