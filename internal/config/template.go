package config

import (
	"bytes"
	"os"
	"text/template"
)

// configTemplate is a well-commented YAML config template for new users.
// Using text/template instead of yaml.Marshal preserves comments.
const configTemplate = `# PRT Configuration
# https://github.com/ChrisEdwards/prt

# Your GitHub username (required)
# Used to identify PRs you authored and review requests for you
# Auto-detected if left empty (via ` + "`gh api user`" + `)
github_username: "{{.GitHubUsername}}"

# Team members (GitHub usernames)
# PRs from these users are highlighted as "Team PRs"
team_members:
{{- range .TeamMembers}}
  - "{{.}}"
{{- else}}
  # - "teammate1"
  # - "teammate2"
{{- end}}

# Directories to search for Git repositories
# Supports absolute paths and ~ for home directory
search_paths:
{{- range .SearchPaths}}
  - "{{.}}"
{{- else}}
  # - "~/code/work"
  # - "~/projects"
{{- end}}

# Repository name patterns to include (glob syntax)
# Leave empty to include all discovered repositories
# Examples: "myorg-*", "*-api", "frontend"
include_repos:
{{- range .IncludeRepos}}
  - "{{.}}"
{{- else}}
  # - "myorg-*"
{{- end}}

# Maximum directory depth when searching for repositories
# Default: 3
scan_depth: {{.ScanDepth}}

# Known bot accounts (PRs from these are de-prioritized)
# Pre-populated with common bots; add your org's bots here
bots:
{{- range .Bots}}
  - "{{.}}"
{{- end}}

# Default grouping: "project" or "author"
default_group_by: "{{.DefaultGroupBy}}"

# Default sort order: "oldest" or "newest" (by creation date)
default_sort: "{{.DefaultSort}}"

# Show branch names in PR output
show_branch_name: {{.ShowBranchName}}

# Show icons (requires a Nerd Font or emoji support)
show_icons: {{.ShowIcons}}

# Show "Other PRs" section (external contributors, bots, etc.)
# Default: false (hidden to reduce noise)
show_other_prs: {{.ShowOtherPRs}}

# Hide PRs older than this many days (0 = no limit)
# Useful for filtering out stale/long-running PRs
max_pr_age_days: {{.MaxPRAgeDays}}
`

// GenerateConfigFile generates a well-commented YAML config file from the given config.
func GenerateConfigFile(cfg *Config) (string, error) {
	tmpl, err := template.New("config").Parse(configTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// SaveConfig saves the config to the default config file location.
// Creates the config directory if it doesn't exist.
func SaveConfig(cfg *Config) error {
	content, err := GenerateConfigFile(cfg)
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(ConfigDir(), 0755); err != nil {
		return err
	}

	return os.WriteFile(ConfigPath(), []byte(content), 0644)
}
