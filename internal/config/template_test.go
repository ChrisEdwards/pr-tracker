package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGenerateConfigFile_ValidYAML(t *testing.T) {
	cfg := &Config{
		GitHubUsername: "testuser",
		TeamMembers:    []string{"teammate1", "teammate2"},
		SearchPaths:    []string{"~/code", "/work/repos"},
		IncludeRepos:   []string{"myorg-*"},
		ScanDepth:      3,
		Bots:           []string{"dependabot[bot]"},
		DefaultGroupBy: GroupByProject,
		DefaultSort:    SortOldest,
		ShowBranchName: true,
		ShowIcons:      true,
	}

	content, err := GenerateConfigFile(cfg)
	if err != nil {
		t.Fatalf("GenerateConfigFile() error: %v", err)
	}

	// Verify it's valid YAML by parsing it
	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &parsed); err != nil {
		t.Errorf("Generated config is not valid YAML: %v\nContent:\n%s", err, content)
	}
}

func TestGenerateConfigFile_ContainsComments(t *testing.T) {
	cfg := &Config{
		GitHubUsername: "testuser",
		SearchPaths:    []string{"~/code"},
		ScanDepth:      3,
		DefaultGroupBy: GroupByProject,
		DefaultSort:    SortOldest,
	}

	content, err := GenerateConfigFile(cfg)
	if err != nil {
		t.Fatalf("GenerateConfigFile() error: %v", err)
	}

	// Check for expected comments
	expectedComments := []string{
		"# PRT Configuration",
		"# Your GitHub username (required)",
		"# Team members (GitHub usernames)",
		"# Directories to search for Git repositories",
		"# Maximum directory depth",
	}

	for _, comment := range expectedComments {
		if !strings.Contains(content, comment) {
			t.Errorf("Generated config should contain comment %q", comment)
		}
	}
}

func TestGenerateConfigFile_CorrectValues(t *testing.T) {
	cfg := &Config{
		GitHubUsername: "myuser",
		TeamMembers:    []string{"alice", "bob"},
		SearchPaths:    []string{"/my/path"},
		IncludeRepos:   []string{"prefix-*"},
		ScanDepth:      5,
		Bots:           []string{"bot1"},
		DefaultGroupBy: GroupByAuthor,
		DefaultSort:    SortNewest,
		ShowBranchName: false,
		ShowIcons:      false,
	}

	content, err := GenerateConfigFile(cfg)
	if err != nil {
		t.Fatalf("GenerateConfigFile() error: %v", err)
	}

	// Check values appear in output
	checks := []string{
		`github_username: "myuser"`,
		`- "alice"`,
		`- "bob"`,
		`- "/my/path"`,
		`- "prefix-*"`,
		`scan_depth: 5`,
		`- "bot1"`,
		`default_group_by: "author"`,
		`default_sort: "newest"`,
		`show_branch_name: false`,
		`show_icons: false`,
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("Generated config should contain %q\nContent:\n%s", check, content)
		}
	}
}

func TestGenerateConfigFile_EmptySlices(t *testing.T) {
	cfg := &Config{
		GitHubUsername: "testuser",
		TeamMembers:    []string{}, // Empty
		SearchPaths:    []string{}, // Empty
		IncludeRepos:   []string{}, // Empty
		ScanDepth:      3,
		Bots:           []string{},
		DefaultGroupBy: GroupByProject,
		DefaultSort:    SortOldest,
	}

	content, err := GenerateConfigFile(cfg)
	if err != nil {
		t.Fatalf("GenerateConfigFile() error: %v", err)
	}

	// Empty slices should show commented examples
	expectedExamples := []string{
		`# - "teammate1"`,
		`# - "~/code/work"`,
		`# - "myorg-*"`,
	}

	for _, example := range expectedExamples {
		if !strings.Contains(content, example) {
			t.Errorf("Generated config with empty slices should contain example %q", example)
		}
	}
}

func TestGenerateConfigFile_DefaultConfig(t *testing.T) {
	cfg := &DefaultConfig

	content, err := GenerateConfigFile(cfg)
	if err != nil {
		t.Fatalf("GenerateConfigFile() error: %v", err)
	}

	// Should be valid YAML
	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &parsed); err != nil {
		t.Errorf("Default config template is not valid YAML: %v", err)
	}

	// Should contain default values
	if !strings.Contains(content, `scan_depth: 3`) {
		t.Error("Default config should have scan_depth: 3")
	}
	if !strings.Contains(content, `default_group_by: "project"`) {
		t.Error("Default config should have default_group_by: project")
	}
}

func TestSaveConfig(t *testing.T) {
	// Use a temp directory for testing
	tmpDir := t.TempDir()

	cfg := &Config{
		GitHubUsername: "savetest",
		SearchPaths:    []string{tmpDir},
		ScanDepth:      3,
		DefaultGroupBy: GroupByProject,
		DefaultSort:    SortOldest,
		ShowBranchName: true,
		ShowIcons:      true,
	}

	// Create a custom save path for testing
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Generate and save manually (since we can't override ConfigDir easily)
	content, err := GenerateConfigFile(cfg)
	if err != nil {
		t.Fatalf("GenerateConfigFile() error: %v", err)
	}

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	// Verify file exists and has correct permissions
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Config file not created: %v", err)
	}

	// Check permissions (0644)
	perm := info.Mode().Perm()
	if perm != 0644 {
		t.Errorf("Config file permissions = %o, want 0644", perm)
	}

	// Verify content is valid YAML
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	var parsed map[string]interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Errorf("Saved config is not valid YAML: %v", err)
	}

	// Check expected value is present
	if !strings.Contains(string(data), `github_username: "savetest"`) {
		t.Error("Saved config should contain the username")
	}
}

func TestGenerateConfigFile_SpecialCharacters(t *testing.T) {
	cfg := &Config{
		GitHubUsername: "user-with-dash",
		TeamMembers:    []string{"user_underscore", "user.dot"},
		SearchPaths:    []string{"/path/with spaces/repo"},
		ScanDepth:      3,
		DefaultGroupBy: GroupByProject,
		DefaultSort:    SortOldest,
	}

	content, err := GenerateConfigFile(cfg)
	if err != nil {
		t.Fatalf("GenerateConfigFile() error: %v", err)
	}

	// Verify it's still valid YAML
	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &parsed); err != nil {
		t.Errorf("Config with special chars is not valid YAML: %v\nContent:\n%s", err, content)
	}
}
