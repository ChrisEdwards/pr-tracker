package config

import (
	"os"
	"path/filepath"
	"strings"
)

// KnownBots is a pre-populated list of common GitHub bot accounts.
// These are used to filter out bot PRs from the "team" and "other" categories.
var KnownBots = []string{
	"dependabot[bot]",
	"dependabot",
	"renovate[bot]",
	"renovate",
	"github-actions[bot]",
	"codecov[bot]",
	"codecov",
	"semantic-release-bot",
	"greenkeeper[bot]",
	"snyk-bot",
	"imgbot[bot]",
	"allcontributors[bot]",
	"mergify[bot]",
	"kodiakhq[bot]",
	"stale[bot]",
}

// DefaultConfig returns sensible default configuration values.
// Note: GitHubUsername and SearchPaths must be set by user or auto-detected.
var DefaultConfig = Config{
	GitHubUsername: "",             // Must be set or auto-detected
	TeamMembers:    []string{},     // No team members by default
	SearchPaths:    []string{},     // Must be set by user
	IncludeRepos:   []string{},     // Empty = match all repos
	ScanDepth:      3,              // Reasonable default depth
	Bots:           KnownBots,      // Pre-populated bot list
	DefaultGroupBy: GroupByProject, // Group by project by default
	DefaultSort:    SortOldest,     // Show oldest PRs first (needs attention)
	ShowBranchName: true,           // Show branch names
	ShowIcons:      true,           // Show status icons
	ShowOtherPRs:   false,          // Hide "Other PRs" by default
	MaxPRAgeDays:   0,              // No age limit by default (0 = show all)
}

// ConfigDir returns the path to the PRT configuration directory.
// Default: ~/.prt
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home is unavailable
		return ".prt"
	}
	return filepath.Join(home, ".prt")
}

// ConfigPath returns the path to the PRT configuration file.
// Default: ~/.prt/config.yaml
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

// ExpandPath expands ~ to the user's home directory in a path.
func ExpandPath(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if path == "~" {
		return home
	}

	// Handle ~/something
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}

	return path
}

// ExpandPaths expands ~ in all paths in the slice.
func ExpandPaths(paths []string) []string {
	result := make([]string, len(paths))
	for i, p := range paths {
		result[i] = ExpandPath(p)
	}
	return result
}
