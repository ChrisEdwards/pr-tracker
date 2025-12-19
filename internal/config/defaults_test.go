package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKnownBots(t *testing.T) {
	// Verify the list is not empty
	if len(KnownBots) == 0 {
		t.Error("KnownBots should not be empty")
	}

	// Check for expected bots
	expectedBots := []string{
		"dependabot[bot]",
		"renovate[bot]",
		"github-actions[bot]",
	}

	for _, expected := range expectedBots {
		found := false
		for _, bot := range KnownBots {
			if bot == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("KnownBots should contain %q", expected)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	// Verify default values
	if DefaultConfig.ScanDepth != 3 {
		t.Errorf("ScanDepth = %d, want 3", DefaultConfig.ScanDepth)
	}
	if DefaultConfig.DefaultGroupBy != GroupByProject {
		t.Errorf("DefaultGroupBy = %q, want %q", DefaultConfig.DefaultGroupBy, GroupByProject)
	}
	if DefaultConfig.DefaultSort != SortOldest {
		t.Errorf("DefaultSort = %q, want %q", DefaultConfig.DefaultSort, SortOldest)
	}
	if !DefaultConfig.ShowBranchName {
		t.Error("ShowBranchName should be true by default")
	}
	if !DefaultConfig.ShowIcons {
		t.Error("ShowIcons should be true by default")
	}

	// Verify Bots is populated with KnownBots
	if len(DefaultConfig.Bots) != len(KnownBots) {
		t.Errorf("Bots length = %d, want %d", len(DefaultConfig.Bots), len(KnownBots))
	}

	// Verify empty slices are initialized
	if DefaultConfig.TeamMembers == nil {
		t.Error("TeamMembers should be initialized")
	}
	if DefaultConfig.SearchPaths == nil {
		t.Error("SearchPaths should be initialized")
	}
	if DefaultConfig.IncludeRepos == nil {
		t.Error("IncludeRepos should be initialized")
	}
}

func TestConfigDir(t *testing.T) {
	dir := ConfigDir()

	// Should end with .prt
	if !strings.HasSuffix(dir, ".prt") {
		t.Errorf("ConfigDir() = %q, should end with .prt", dir)
	}

	// Should be an absolute path (or .prt for fallback)
	if dir != ".prt" && !filepath.IsAbs(dir) {
		t.Errorf("ConfigDir() = %q, should be absolute or .prt", dir)
	}
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()

	// Should end with config.yaml
	if !strings.HasSuffix(path, "config.yaml") {
		t.Errorf("ConfigPath() = %q, should end with config.yaml", path)
	}

	// Should contain .prt directory
	if !strings.Contains(path, ".prt") {
		t.Errorf("ConfigPath() = %q, should contain .prt", path)
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no tilde",
			input: "/absolute/path",
			want:  "/absolute/path",
		},
		{
			name:  "tilde only",
			input: "~",
			want:  home,
		},
		{
			name:  "tilde with path",
			input: "~/code/project",
			want:  filepath.Join(home, "code/project"),
		},
		{
			name:  "relative path",
			input: "relative/path",
			want:  "relative/path",
		},
		{
			name:  "tilde without slash",
			input: "~user",
			want:  "~user", // Not expanded (different user)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandPath(tt.input)
			if got != tt.want {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExpandPaths(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	input := []string{"~/code", "/absolute", "relative"}
	result := ExpandPaths(input)

	if len(result) != 3 {
		t.Fatalf("ExpandPaths returned %d items, want 3", len(result))
	}

	if result[0] != filepath.Join(home, "code") {
		t.Errorf("result[0] = %q, want %q", result[0], filepath.Join(home, "code"))
	}
	if result[1] != "/absolute" {
		t.Errorf("result[1] = %q, want %q", result[1], "/absolute")
	}
	if result[2] != "relative" {
		t.Errorf("result[2] = %q, want %q", result[2], "relative")
	}
}
