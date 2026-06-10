package config

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strconv"
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

# Known Bot Authors
# Pre-populated with common bots; add your org's automation accounts here
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

# Show "Other PRs" section (external contributors, Bot Authors, etc.)
# Default: false (hidden to reduce noise)
show_other_prs: {{.ShowOtherPRs}}

# Hide PRs older than this many days (0 = no limit)
# Useful for filtering out stale/long-running PRs
max_pr_age_days: {{.MaxPRAgeDays}}

# Custom Views are named filter bundles selectable with --view=<name>.
# Views only support filters in v1; display, group, sort, and color options
# are configured globally or via CLI flags.
views:
{{- range $name, $view := .Views}}
  {{$name}}:
{{- if $view.Description}}
    description: {{yamlString $view.Description}}
{{- end}}
    filters:
{{- range $entry := viewFilterEntries $view.Filters}}
      {{$entry.Key}}: {{yamlValue $entry.Value}}
{{- end}}
{{- else}}
  # alice-review:
  #   description: "Alice's ready PRs that still need approval"
  #   filters:
  #     author: "@alice"
  #     draft: false
  #     bot: false
  #     review_decision: "not-approved"
{{- end}}
`

// GenerateConfigFile generates a well-commented YAML config file from the given config.
func GenerateConfigFile(cfg *Config) (string, error) {
	tmpl, err := template.New("config").Funcs(template.FuncMap{
		"viewFilterEntries": viewFilterEntries,
		"yamlString":        yamlString,
		"yamlValue":         yamlValue,
	}).Parse(configTemplate)
	if err != nil {
		return "", err
	}

	templateConfig := *cfg
	templateConfig.Views = RegisteredViews(cfg)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, &templateConfig); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func yamlString(value string) string {
	return strconv.Quote(value)
}

func yamlValue(value interface{}) string {
	switch typed := value.(type) {
	case string:
		if isPlainYAMLViewScalar(typed) {
			return typed
		}
		return yamlString(typed)
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fmt.Sprint(typed)
	}
}

type viewFilterEntry struct {
	Key   string
	Value interface{}
}

func viewFilterEntries(filters map[string]interface{}) []viewFilterEntry {
	entries := make([]viewFilterEntry, 0, len(filters))
	for key, value := range filters {
		entries = append(entries, viewFilterEntry{Key: key, Value: value})
	}

	order := map[string]int{
		"author":          0,
		"draft":           1,
		"bot":             2,
		"review_decision": 3,
	}
	sort.Slice(entries, func(i, j int) bool {
		leftRank, leftKnown := order[entries[i].Key]
		rightRank, rightKnown := order[entries[j].Key]
		switch {
		case leftKnown && rightKnown:
			return leftRank < rightRank
		case leftKnown:
			return true
		case rightKnown:
			return false
		default:
			return entries[i].Key < entries[j].Key
		}
	})

	return entries
}

func isPlainYAMLViewScalar(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
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
