// Package scanner provides functionality for discovering and inspecting
// Git repositories with GitHub remotes.
package scanner

import (
	"fmt"

	"github.com/gobwas/glob"
)

// RepoFilter filters repository names using glob patterns.
// Empty patterns match all repositories.
type RepoFilter struct {
	patterns []glob.Glob
}

// NewRepoFilter creates a RepoFilter from a list of glob pattern strings.
// Returns an error if any pattern is invalid.
//
// Pattern examples:
//   - "myorg-*" matches repos starting with "myorg-"
//   - "*-api" matches repos ending with "-api"
//   - "frontend" matches exactly "frontend"
//   - "*service*" matches repos containing "service"
func NewRepoFilter(patterns []string) (*RepoFilter, error) {
	globs := make([]glob.Glob, 0, len(patterns))
	for _, pattern := range patterns {
		g, err := glob.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
		}
		globs = append(globs, g)
	}
	return &RepoFilter{patterns: globs}, nil
}

// Matches returns true if the given name matches any of the filter patterns.
// If no patterns are configured, all names match (empty filter = include all).
// Matching is case-sensitive since GitHub repository names are case-sensitive.
func (f *RepoFilter) Matches(name string) bool {
	// No patterns = match all
	if len(f.patterns) == 0 {
		return true
	}

	// Match against any pattern (OR logic)
	for _, g := range f.patterns {
		if g.Match(name) {
			return true
		}
	}

	return false
}

// HasPatterns returns true if the filter has any patterns configured.
func (f *RepoFilter) HasPatterns() bool {
	return len(f.patterns) > 0
}
