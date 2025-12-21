package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// ValidationError holds multiple configuration validation errors.
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("configuration errors:\n  - %s", strings.Join(e.Errors, "\n  - "))
}

// Validate checks the configuration and returns an error if invalid.
// Error messages are designed to be actionable and helpful.
func (c *Config) Validate() error {
	var errs []string

	// Username required
	if c.GitHubUsername == "" {
		errs = append(errs, "github_username is required (set in config or via gh CLI auto-detect)")
	}

	// At least one search path required
	if len(c.SearchPaths) == 0 {
		errs = append(errs, "at least one search_path is required")
	}

	// Validate search paths exist
	for _, path := range c.SearchPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			errs = append(errs, fmt.Sprintf("search path does not exist: %s", path))
		}
	}

	// Valid group_by value
	if !IsValidGroupBy(c.DefaultGroupBy) {
		errs = append(errs, fmt.Sprintf("invalid default_group_by: %q (must be %q or %q)", c.DefaultGroupBy, GroupByProject, GroupByAuthor))
	}

	// Valid sort value
	if !IsValidSort(c.DefaultSort) {
		errs = append(errs, fmt.Sprintf("invalid default_sort: %q (must be %q or %q)", c.DefaultSort, SortOldest, SortNewest))
	}

	// Scan depth must be positive
	if c.ScanDepth < 1 {
		errs = append(errs, "scan_depth must be at least 1")
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}

	return nil
}

// Flags holds CLI flag values that can override config.
type Flags struct {
	Path    string // Override search_paths with a single path
	Filter  string // Filter repos by pattern
	Group   string // Override default_group_by
	Depth   int    // Override scan_depth
	JSON    bool   // Output in JSON format
	NoColor bool   // Disable colored output
}

// Load loads configuration with the following precedence (highest to lowest):
// 1. CLI flags
// 2. Environment variables (PRT_* prefix)
// 3. Config file (~/.prt/config.yaml)
// 4. Hardcoded defaults
func Load(flags *Flags) (*Config, error) {
	v := viper.New()

	// 1. Set defaults from DefaultConfig
	v.SetDefault("github_username", DefaultConfig.GitHubUsername)
	v.SetDefault("team_members", DefaultConfig.TeamMembers)
	v.SetDefault("search_paths", DefaultConfig.SearchPaths)
	v.SetDefault("include_repos", DefaultConfig.IncludeRepos)
	v.SetDefault("scan_depth", DefaultConfig.ScanDepth)
	v.SetDefault("bots", DefaultConfig.Bots)
	v.SetDefault("default_group_by", DefaultConfig.DefaultGroupBy)
	v.SetDefault("default_sort", DefaultConfig.DefaultSort)
	v.SetDefault("show_branch_name", DefaultConfig.ShowBranchName)
	v.SetDefault("show_icons", DefaultConfig.ShowIcons)
	v.SetDefault("show_other_prs", DefaultConfig.ShowOtherPRs)
	v.SetDefault("max_pr_age_days", DefaultConfig.MaxPRAgeDays)

	// 2. Load config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(ConfigDir())

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
		// Config file not found - this is OK, will use defaults
		// The wizard will be triggered later if required fields are missing
	}

	// 3. Environment variables with PRT_ prefix
	v.SetEnvPrefix("PRT")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 4. CLI flag overrides (highest precedence)
	if flags != nil {
		if flags.Path != "" {
			v.Set("search_paths", []string{flags.Path})
		}
		if flags.Depth > 0 {
			v.Set("scan_depth", flags.Depth)
		}
		if flags.Group != "" {
			v.Set("default_group_by", flags.Group)
		}
		if flags.Filter != "" {
			v.Set("include_repos", []string{flags.Filter})
		}
	}

	// 5. Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	// 6. Expand ~ in paths
	cfg.SearchPaths = ExpandPaths(cfg.SearchPaths)

	return &cfg, nil
}

// LoadDefault returns the default configuration without reading any files.
func LoadDefault() *Config {
	cfg := DefaultConfig
	return &cfg
}

// NeedsSetup returns true if the config is missing required fields
// that should trigger the first-run wizard.
func NeedsSetup(cfg *Config) bool {
	// Config needs setup if no search paths or no GitHub username
	return len(cfg.SearchPaths) == 0 || cfg.GitHubUsername == ""
}

// ConfigFileExists returns true if a config file exists at the default path.
func ConfigFileExists() bool {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(ConfigDir())

	err := v.ReadInConfig()
	return err == nil
}
