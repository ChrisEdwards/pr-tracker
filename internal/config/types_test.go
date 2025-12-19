package config

import (
	"testing"
)

func TestIsValidGroupBy(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"project", GroupByProject, true},
		{"author", GroupByAuthor, true},
		{"invalid", "invalid", false},
		{"empty", "", false},
		{"uppercase", "PROJECT", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidGroupBy(tt.value); got != tt.want {
				t.Errorf("IsValidGroupBy(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestIsValidSort(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"oldest", SortOldest, true},
		{"newest", SortNewest, true},
		{"invalid", "invalid", false},
		{"empty", "", false},
		{"uppercase", "OLDEST", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidSort(tt.value); got != tt.want {
				t.Errorf("IsValidSort(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	// Verify constant values match expected strings
	if GroupByProject != "project" {
		t.Errorf("GroupByProject = %q, want %q", GroupByProject, "project")
	}
	if GroupByAuthor != "author" {
		t.Errorf("GroupByAuthor = %q, want %q", GroupByAuthor, "author")
	}
	if SortOldest != "oldest" {
		t.Errorf("SortOldest = %q, want %q", SortOldest, "oldest")
	}
	if SortNewest != "newest" {
		t.Errorf("SortNewest = %q, want %q", SortNewest, "newest")
	}
}

func TestConfigStruct(t *testing.T) {
	// Test that Config can be instantiated with expected fields
	cfg := Config{
		GitHubUsername: "testuser",
		TeamMembers:    []string{"teammate1", "teammate2"},
		SearchPaths:    []string{"~/code", "~/projects"},
		IncludeRepos:   []string{"org/*", "personal-*"},
		ScanDepth:      3,
		Bots:           []string{"dependabot[bot]", "renovate[bot]"},
		DefaultGroupBy: GroupByProject,
		DefaultSort:    SortOldest,
		ShowBranchName: true,
		ShowIcons:      true,
	}

	if cfg.GitHubUsername != "testuser" {
		t.Errorf("GitHubUsername = %q, want %q", cfg.GitHubUsername, "testuser")
	}
	if len(cfg.TeamMembers) != 2 {
		t.Errorf("TeamMembers length = %d, want %d", len(cfg.TeamMembers), 2)
	}
	if cfg.ScanDepth != 3 {
		t.Errorf("ScanDepth = %d, want %d", cfg.ScanDepth, 3)
	}
	if !cfg.ShowBranchName {
		t.Error("ShowBranchName should be true")
	}
	if !cfg.ShowIcons {
		t.Error("ShowIcons should be true")
	}
}
