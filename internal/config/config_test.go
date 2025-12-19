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
			cfg:  Config{SearchPaths: []string{"/some/path"}},
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
