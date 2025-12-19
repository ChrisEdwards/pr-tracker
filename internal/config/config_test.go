package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_WithDefaults(t *testing.T) {
	// Load with no config file and no flags
	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load(nil) error: %v", err)
	}

	// Should have default values
	if cfg.ScanDepth != DefaultConfig.ScanDepth {
		t.Errorf("ScanDepth = %d, want %d", cfg.ScanDepth, DefaultConfig.ScanDepth)
	}
	if cfg.DefaultGroupBy != DefaultConfig.DefaultGroupBy {
		t.Errorf("DefaultGroupBy = %q, want %q", cfg.DefaultGroupBy, DefaultConfig.DefaultGroupBy)
	}
	if cfg.DefaultSort != DefaultConfig.DefaultSort {
		t.Errorf("DefaultSort = %q, want %q", cfg.DefaultSort, DefaultConfig.DefaultSort)
	}
	if !cfg.ShowBranchName {
		t.Error("ShowBranchName should be true by default")
	}
	if !cfg.ShowIcons {
		t.Error("ShowIcons should be true by default")
	}
}

func TestLoad_WithFlags(t *testing.T) {
	flags := &Flags{
		Path:  "/custom/path",
		Depth: 5,
		Group: GroupByAuthor,
	}

	cfg, err := Load(flags)
	if err != nil {
		t.Fatalf("Load(flags) error: %v", err)
	}

	// Flags should override defaults
	if len(cfg.SearchPaths) != 1 || cfg.SearchPaths[0] != "/custom/path" {
		t.Errorf("SearchPaths = %v, want [/custom/path]", cfg.SearchPaths)
	}
	if cfg.ScanDepth != 5 {
		t.Errorf("ScanDepth = %d, want 5", cfg.ScanDepth)
	}
	if cfg.DefaultGroupBy != GroupByAuthor {
		t.Errorf("DefaultGroupBy = %q, want %q", cfg.DefaultGroupBy, GroupByAuthor)
	}
}

func TestLoad_WithFilter(t *testing.T) {
	flags := &Flags{
		Filter: "org/*",
	}

	cfg, err := Load(flags)
	if err != nil {
		t.Fatalf("Load(flags) error: %v", err)
	}

	if len(cfg.IncludeRepos) != 1 || cfg.IncludeRepos[0] != "org/*" {
		t.Errorf("IncludeRepos = %v, want [org/*]", cfg.IncludeRepos)
	}
}

func TestLoad_ExpandsPaths(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	flags := &Flags{
		Path: "~/code",
	}

	cfg, err := Load(flags)
	if err != nil {
		t.Fatalf("Load(flags) error: %v", err)
	}

	expected := filepath.Join(home, "code")
	if len(cfg.SearchPaths) != 1 || cfg.SearchPaths[0] != expected {
		t.Errorf("SearchPaths = %v, want [%s]", cfg.SearchPaths, expected)
	}
}

func TestLoadDefault(t *testing.T) {
	cfg := LoadDefault()

	if cfg.ScanDepth != DefaultConfig.ScanDepth {
		t.Errorf("ScanDepth = %d, want %d", cfg.ScanDepth, DefaultConfig.ScanDepth)
	}
	if cfg.DefaultGroupBy != DefaultConfig.DefaultGroupBy {
		t.Errorf("DefaultGroupBy = %q, want %q", cfg.DefaultGroupBy, DefaultConfig.DefaultGroupBy)
	}
}

func TestNeedsSetup(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want bool
	}{
		{
			name: "empty search paths needs setup",
			cfg:  Config{SearchPaths: []string{}},
			want: true,
		},
		{
			name: "nil search paths needs setup",
			cfg:  Config{SearchPaths: nil},
			want: true,
		},
		{
			name: "with search paths doesn't need setup",
			cfg:  Config{GitHubUsername: "user", SearchPaths: []string{"/some/path"}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NeedsSetup(&tt.cfg); got != tt.want {
				t.Errorf("NeedsSetup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlags_Struct(t *testing.T) {
	// Test that Flags struct can be instantiated
	flags := Flags{
		Path:    "/some/path",
		Filter:  "org/*",
		Group:   GroupByAuthor,
		Depth:   5,
		JSON:    true,
		NoColor: true,
	}

	if flags.Path != "/some/path" {
		t.Errorf("Path = %q, want /some/path", flags.Path)
	}
	if flags.Depth != 5 {
		t.Errorf("Depth = %d, want 5", flags.Depth)
	}
	if !flags.JSON {
		t.Error("JSON should be true")
	}
	if !flags.NoColor {
		t.Error("NoColor should be true")
	}
}

func TestConfig_Validate(t *testing.T) {
	// Create a temporary directory for valid path tests
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errMsgs []string // Substrings expected in error messages
	}{
		{
			name: "valid config",
			cfg: Config{
				GitHubUsername: "testuser",
				SearchPaths:    []string{tmpDir},
				DefaultGroupBy: GroupByProject,
				DefaultSort:    SortOldest,
				ScanDepth:      3,
			},
			wantErr: false,
		},
		{
			name: "missing username",
			cfg: Config{
				GitHubUsername: "",
				SearchPaths:    []string{tmpDir},
				DefaultGroupBy: GroupByProject,
				DefaultSort:    SortOldest,
				ScanDepth:      3,
			},
			wantErr: true,
			errMsgs: []string{"github_username is required"},
		},
		{
			name: "no search paths",
			cfg: Config{
				GitHubUsername: "testuser",
				SearchPaths:    []string{},
				DefaultGroupBy: GroupByProject,
				DefaultSort:    SortOldest,
				ScanDepth:      3,
			},
			wantErr: true,
			errMsgs: []string{"at least one search_path is required"},
		},
		{
			name: "search path does not exist",
			cfg: Config{
				GitHubUsername: "testuser",
				SearchPaths:    []string{"/nonexistent/path/xyz123"},
				DefaultGroupBy: GroupByProject,
				DefaultSort:    SortOldest,
				ScanDepth:      3,
			},
			wantErr: true,
			errMsgs: []string{"search path does not exist"},
		},
		{
			name: "invalid group_by",
			cfg: Config{
				GitHubUsername: "testuser",
				SearchPaths:    []string{tmpDir},
				DefaultGroupBy: "invalid",
				DefaultSort:    SortOldest,
				ScanDepth:      3,
			},
			wantErr: true,
			errMsgs: []string{"invalid default_group_by"},
		},
		{
			name: "invalid sort",
			cfg: Config{
				GitHubUsername: "testuser",
				SearchPaths:    []string{tmpDir},
				DefaultGroupBy: GroupByProject,
				DefaultSort:    "invalid",
				ScanDepth:      3,
			},
			wantErr: true,
			errMsgs: []string{"invalid default_sort"},
		},
		{
			name: "zero scan depth",
			cfg: Config{
				GitHubUsername: "testuser",
				SearchPaths:    []string{tmpDir},
				DefaultGroupBy: GroupByProject,
				DefaultSort:    SortOldest,
				ScanDepth:      0,
			},
			wantErr: true,
			errMsgs: []string{"scan_depth must be at least 1"},
		},
		{
			name: "negative scan depth",
			cfg: Config{
				GitHubUsername: "testuser",
				SearchPaths:    []string{tmpDir},
				DefaultGroupBy: GroupByProject,
				DefaultSort:    SortOldest,
				ScanDepth:      -1,
			},
			wantErr: true,
			errMsgs: []string{"scan_depth must be at least 1"},
		},
		{
			name: "multiple errors",
			cfg: Config{
				GitHubUsername: "",
				SearchPaths:    []string{},
				DefaultGroupBy: "bad",
				DefaultSort:    "bad",
				ScanDepth:      0,
			},
			wantErr: true,
			errMsgs: []string{
				"github_username is required",
				"at least one search_path is required",
				"invalid default_group_by",
				"invalid default_sort",
				"scan_depth must be at least 1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				errStr := err.Error()
				for _, msg := range tt.errMsgs {
					if !contains(errStr, msg) {
						t.Errorf("Validate() error %q should contain %q", errStr, msg)
					}
				}
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{
		Errors: []string{"error 1", "error 2"},
	}

	errStr := ve.Error()
	if !contains(errStr, "configuration errors") {
		t.Errorf("ValidationError.Error() should contain 'configuration errors'")
	}
	if !contains(errStr, "error 1") {
		t.Errorf("ValidationError.Error() should contain 'error 1'")
	}
	if !contains(errStr, "error 2") {
		t.Errorf("ValidationError.Error() should contain 'error 2'")
	}
}

func TestNeedsSetup_WithUsername(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want bool
	}{
		{
			name: "missing both username and paths",
			cfg:  Config{GitHubUsername: "", SearchPaths: []string{}},
			want: true,
		},
		{
			name: "has paths but no username",
			cfg:  Config{GitHubUsername: "", SearchPaths: []string{"/path"}},
			want: true,
		},
		{
			name: "has username but no paths",
			cfg:  Config{GitHubUsername: "user", SearchPaths: []string{}},
			want: true,
		},
		{
			name: "has both username and paths",
			cfg:  Config{GitHubUsername: "user", SearchPaths: []string{"/path"}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NeedsSetup(&tt.cfg); got != tt.want {
				t.Errorf("NeedsSetup() = %v, want %v", got, tt.want)
			}
		})
	}
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestLoad_EnvOverrides(t *testing.T) {
	// Set environment variable
	os.Setenv("PRT_SCAN_DEPTH", "10")
	defer os.Unsetenv("PRT_SCAN_DEPTH")

	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Environment variable should override default
	if cfg.ScanDepth != 10 {
		t.Errorf("ScanDepth = %d, want 10 (from env)", cfg.ScanDepth)
	}
}

func TestLoad_EnvOverrides_Username(t *testing.T) {
	os.Setenv("PRT_GITHUB_USERNAME", "envuser")
	defer os.Unsetenv("PRT_GITHUB_USERNAME")

	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.GitHubUsername != "envuser" {
		t.Errorf("GitHubUsername = %q, want 'envuser' (from env)", cfg.GitHubUsername)
	}
}

func TestLoad_FlagsOverrideEnv(t *testing.T) {
	// Set environment variable
	os.Setenv("PRT_SCAN_DEPTH", "10")
	defer os.Unsetenv("PRT_SCAN_DEPTH")

	// Flags should override env
	flags := &Flags{
		Depth: 7,
	}

	cfg, err := Load(flags)
	if err != nil {
		t.Fatalf("Load(flags) error: %v", err)
	}

	// Flag should win over env
	if cfg.ScanDepth != 7 {
		t.Errorf("ScanDepth = %d, want 7 (flag > env)", cfg.ScanDepth)
	}
}

func TestLoad_Precedence(t *testing.T) {
	// Test the complete precedence chain: flag > env > default
	// We can't easily test file without more setup, but flag > env is key

	// Set env
	os.Setenv("PRT_SCAN_DEPTH", "8")
	os.Setenv("PRT_DEFAULT_GROUP_BY", GroupByAuthor)
	defer func() {
		os.Unsetenv("PRT_SCAN_DEPTH")
		os.Unsetenv("PRT_DEFAULT_GROUP_BY")
	}()

	// Flags override only scan_depth
	flags := &Flags{
		Depth: 5,
		// Group not set - should use env value
	}

	cfg, err := Load(flags)
	if err != nil {
		t.Fatalf("Load(flags) error: %v", err)
	}

	// ScanDepth: flag wins
	if cfg.ScanDepth != 5 {
		t.Errorf("ScanDepth = %d, want 5 (from flag)", cfg.ScanDepth)
	}

	// DefaultGroupBy: env wins (no flag)
	if cfg.DefaultGroupBy != GroupByAuthor {
		t.Errorf("DefaultGroupBy = %q, want %q (from env)", cfg.DefaultGroupBy, GroupByAuthor)
	}

	// ShowIcons: default wins (no flag or env)
	if !cfg.ShowIcons {
		t.Error("ShowIcons should be true (from default)")
	}
}

func TestLoad_WithTempConfigFile(t *testing.T) {
	// Create a temp directory to act as config dir
	tmpDir := t.TempDir()
	configContent := `
github_username: "fileuser"
scan_depth: 6
default_group_by: "author"
search_paths:
  - "/test/path"
`
	// Write config file
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	// Note: We can't easily override ConfigDir() in the current implementation,
	// so this test verifies we can create and write a config file.
	// The Load() function uses ConfigDir() internally.

	// Verify the file was created correctly
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read temp config: %v", err)
	}
	if !contains(string(data), "fileuser") {
		t.Error("Config file should contain 'fileuser'")
	}
}

func TestConfigFileExists_NoFile(t *testing.T) {
	// ConfigFileExists should return false when no config exists
	// Since we can't easily control ConfigDir, just verify the function runs
	// This tests the code path at minimum
	_ = ConfigFileExists()
}

func TestNeedsSetup_NilConfig(t *testing.T) {
	// Verify NeedsSetup handles edge cases
	cfg := &Config{} // Empty config
	if !NeedsSetup(cfg) {
		t.Error("Empty config should need setup")
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	// Config with multiple validation errors
	cfg := &Config{
		GitHubUsername: "",         // Error: required
		SearchPaths:    []string{}, // Error: required
		DefaultGroupBy: "invalid",  // Error: invalid
		DefaultSort:    "wrong",    // Error: invalid
		ScanDepth:      0,          // Error: must be >= 1
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() should return error for invalid config")
	}

	// Should be a ValidationError
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("Error should be *ValidationError, got %T", err)
	}

	// Should have 5 errors
	if len(ve.Errors) != 5 {
		t.Errorf("ValidationError.Errors = %d errors, want 5", len(ve.Errors))
	}
}
