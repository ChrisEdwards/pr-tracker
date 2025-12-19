# PRT (PR Tracker) - Implementation Specification

> **Version**: 1.0
> **Status**: Ready for Implementation
> **Last Updated**: 2024-12-19

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Project Structure](#2-project-structure)
3. [Dependencies](#3-dependencies)
4. [Data Models](#4-data-models)
5. [Configuration System](#5-configuration-system)
6. [Core Components](#6-core-components)
7. [Implementation Phases](#7-implementation-phases)
8. [CLI Interface](#8-cli-interface)
9. [Display & Styling](#9-display--styling)
10. [Error Handling](#10-error-handling)
11. [Testing Strategy](#11-testing-strategy)
12. [Build & Distribution](#12-build--distribution)
    - [12.1 Makefile](#121-makefile)
    - [12.2 Release Artifacts](#122-release-artifacts)
    - [12.3 Version Embedding](#123-version-embedding)
    - [12.4 Homebrew Installation](#124-homebrew-installation)
    - [12.5 Installation Methods Summary](#125-installation-methods-summary)

---

## 1. Executive Summary

**PRT** is a Go CLI tool that aggregates GitHub Pull Request status across multiple local repositories. It solves "PR Fatigue" by:

- Showing PRs you authored separately from PRs awaiting your review
- Detecting and visualizing stacked PR relationships
- Prioritizing PRs by your role (reviewer > team member > other)
- Providing clickable links for quick browser access
- Supporting team workflows with configurable team member lists

### Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | Go | Fast compilation, single binary, excellent CLI libraries |
| GitHub Integration | `gh` CLI wrapper | Handles auth, widely installed, maintained by GitHub |
| Config Format | YAML | Human-readable, supports comments, familiar to developers |
| No Caching | Always fresh | Simplicity; PRs change frequently |
| Progress Display | Streaming | Immediate feedback on multi-repo scans |

---

## 2. Project Structure

```
prt/
├── cmd/
│   └── prt/
│       └── main.go                 # Application entry point
│
├── internal/
│   ├── cli/
│   │   ├── root.go                 # Root cobra command
│   │   ├── flags.go                # Flag definitions & validation
│   │   └── wizard.go               # Interactive first-run setup
│   │
│   ├── config/
│   │   ├── config.go               # Load, save, validate config
│   │   ├── defaults.go             # Default values & known bots
│   │   └── types.go                # Config struct definitions
│   │
│   ├── github/
│   │   ├── client.go               # gh CLI wrapper & orchestration
│   │   ├── parser.go               # Parse gh JSON output
│   │   ├── retry.go                # Exponential backoff retry
│   │   └── types.go                # GitHub-specific types
│   │
│   ├── scanner/
│   │   ├── scanner.go              # Repository discovery
│   │   ├── filter.go               # Glob pattern matching
│   │   └── git.go                  # Git repo detection & remote parsing
│   │
│   ├── stacks/
│   │   ├── detector.go             # Stacked PR detection algorithm
│   │   └── builder.go              # Build stack tree structures
│   │
│   ├── categorizer/
│   │   ├── categorizer.go          # PR categorization logic
│   │   └── sorter.go               # Sorting within categories
│   │
│   ├── display/
│   │   ├── renderer.go             # Main output orchestration
│   │   ├── styles.go               # Lipgloss style definitions
│   │   ├── sections.go             # Section rendering (My PRs, etc.)
│   │   ├── tree.go                 # Stack tree rendering
│   │   ├── progress.go             # Progress bar component
│   │   └── json.go                 # JSON output formatter
│   │
│   └── models/
│       ├── pr.go                   # PR domain model
│       ├── repo.go                 # Repository domain model
│       ├── stack.go                # Stack/tree domain model
│       └── result.go               # Aggregated results model
│
├── go.mod
├── go.sum
├── Makefile                        # Build, test, release targets
├── README.md                       # User documentation
└── .goreleaser.yaml                # Release automation (optional)
```

### Package Responsibility Matrix

| Package | Responsibility | Dependencies |
|---------|---------------|--------------|
| `cmd/prt` | Entry point, panic recovery | `cli` |
| `cli` | Command parsing, flag handling, wizard | `config`, `scanner`, `github`, `display` |
| `config` | Configuration I/O and validation | `models` |
| `github` | All GitHub CLI interactions | `models` |
| `scanner` | Filesystem traversal, repo detection | `models` |
| `stacks` | PR relationship detection | `models` |
| `categorizer` | PR grouping and sorting | `models`, `config` |
| `display` | Terminal output, progress, JSON | `models`, `stacks` |
| `models` | Domain types, no business logic | (none) |

---

## 3. Dependencies

### Required Dependencies

```go
// go.mod
module github.com/ChrisEdwards/prt

go 1.21

require (
    github.com/spf13/cobra v1.8.0           // CLI framework
    github.com/spf13/viper v1.18.0          // Configuration management
    github.com/charmbracelet/lipgloss v0.9.1 // Terminal styling
    github.com/charmbracelet/bubbles v0.18.0 // Progress bar, spinners
    github.com/charmbracelet/bubbletea v0.25.0 // TUI framework (for progress)
    gopkg.in/yaml.v3 v3.0.1                 // YAML parsing
    github.com/gobwas/glob v0.2.3           // Glob pattern matching
)
```

### External Requirements

| Requirement | Version | Check Method |
|-------------|---------|--------------|
| Go | 1.21+ | Build-time |
| `gh` CLI | 2.0+ | `gh --version` at runtime |
| Git | Any | Implicit via `gh` |
| macOS/Linux | - | Primary targets |

---

## 4. Data Models

### 4.1 PR Model (`internal/models/pr.go`)

```go
package models

import "time"

// PR represents a GitHub Pull Request with all relevant metadata
type PR struct {
    // Identity
    Number int    `json:"number"`
    Title  string `json:"title"`
    URL    string `json:"url"`

    // Authorship
    Author string `json:"author"`

    // State
    State   PRState `json:"state"`
    IsDraft bool    `json:"isDraft"`

    // Branches
    BaseBranch string `json:"baseBranch"` // Target (e.g., "main")
    HeadBranch string `json:"headBranch"` // Source (e.g., "feature-x")

    // Timestamps
    CreatedAt time.Time `json:"createdAt"`

    // CI Status
    CIStatus CIStatus `json:"ciStatus"`

    // Review Information
    ReviewRequests []string     `json:"reviewRequests"` // Usernames requested
    Assignees      []string     `json:"assignees"`
    Reviews        []Review     `json:"reviews"`

    // Computed fields (set during categorization)
    IsReviewRequestedFromMe bool `json:"isReviewRequestedFromMe"`
    IsAssignedToMe          bool `json:"isAssignedToMe"`
    MyReviewStatus          ReviewState `json:"myReviewStatus"` // NONE, APPROVED, CHANGES_REQUESTED, etc.

    // Repository context (set during aggregation)
    RepoName  string `json:"repoName"`
    RepoPath  string `json:"repoPath"`
}

type PRState string

const (
    PRStateOpen   PRState = "OPEN"
    PRStateDraft  PRState = "DRAFT"   // Derived from IsDraft
    PRStateMerged PRState = "MERGED"
    PRStateClosed PRState = "CLOSED"
)

type CIStatus string

const (
    CIStatusPassing CIStatus = "passing"
    CIStatusFailing CIStatus = "failing"
    CIStatusPending CIStatus = "pending"
    CIStatusNone    CIStatus = "none"
)

type Review struct {
    Author    string      `json:"author"`
    State     ReviewState `json:"state"`
    Submitted time.Time   `json:"submittedAt"`
}

type ReviewState string

const (
    ReviewStateNone             ReviewState = "NONE"
    ReviewStateApproved         ReviewState = "APPROVED"
    ReviewStateChangesRequested ReviewState = "CHANGES_REQUESTED"
    ReviewStateCommented        ReviewState = "COMMENTED"
    ReviewStatePending          ReviewState = "PENDING"
    ReviewStateDismissed        ReviewState = "DISMISSED"
)

// Helper methods

func (p *PR) Age() time.Duration {
    return time.Since(p.CreatedAt)
}

func (p *PR) AgeString() string {
    age := p.Age()
    switch {
    case age < time.Hour:
        return fmt.Sprintf("%dm ago", int(age.Minutes()))
    case age < 24*time.Hour:
        return fmt.Sprintf("%dh ago", int(age.Hours()))
    default:
        return fmt.Sprintf("%dd ago", int(age.Hours()/24))
    }
}

func (p *PR) EffectiveState() PRState {
    if p.IsDraft {
        return PRStateDraft
    }
    return p.State
}
```

### 4.2 Repository Model (`internal/models/repo.go`)

```go
package models

// Repository represents a local Git repository linked to GitHub
type Repository struct {
    Name       string `json:"name"`       // e.g., "prt"
    Path       string `json:"path"`       // e.g., "/Users/jdoe/code/prt"
    RemoteURL  string `json:"remoteUrl"`  // e.g., "git@github.com:org/prt.git"
    Owner      string `json:"owner"`      // e.g., "org"
    PRs        []*PR  `json:"prs"`

    // Scan metadata
    ScanError  error  `json:"-"`          // Non-nil if scan failed
    ScanStatus ScanStatus `json:"scanStatus"`
}

type ScanStatus string

const (
    ScanStatusSuccess   ScanStatus = "success"
    ScanStatusNoPRs     ScanStatus = "no_prs"
    ScanStatusError     ScanStatus = "error"
    ScanStatusSkipped   ScanStatus = "skipped"
)

func (r *Repository) FullName() string {
    return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

func (r *Repository) HasPRs() bool {
    return len(r.PRs) > 0
}
```

### 4.3 Stack Model (`internal/models/stack.go`)

```go
package models

// StackNode represents a PR in a stack tree
type StackNode struct {
    PR       *PR          `json:"pr"`
    Parent   *StackNode   `json:"-"`          // Avoid circular JSON
    Children []*StackNode `json:"children"`

    // Stack metadata
    Depth    int  `json:"depth"`              // 0 = root
    IsOrphan bool `json:"isOrphan"`           // Parent was merged
}

// Stack represents a complete PR stack (may have multiple roots)
type Stack struct {
    Roots []*StackNode `json:"roots"`

    // Flattened view for easy iteration
    AllNodes []*StackNode `json:"allNodes"`
}

// IsBlocked returns true if this node has an unmerged parent
func (n *StackNode) IsBlocked() bool {
    return n.Parent != nil && !n.IsOrphan
}

// GetRoot walks up to find the root of this stack
func (n *StackNode) GetRoot() *StackNode {
    current := n
    for current.Parent != nil {
        current = current.Parent
    }
    return current
}
```

### 4.4 Result Model (`internal/models/result.go`)

```go
package models

// ScanResult contains the complete output of a prt scan
type ScanResult struct {
    // Categorized PRs
    MyPRs            []*PR `json:"myPRs"`
    NeedsMyAttention []*PR `json:"needsMyAttention"`
    TeamPRs          []*PR `json:"teamPRs"`
    OtherPRs         []*PR `json:"otherPRs"`

    // Repository information
    ReposWithPRs    []*Repository `json:"reposWithPRs"`
    ReposWithoutPRs []*Repository `json:"reposWithoutPRs"`
    ReposWithErrors []*Repository `json:"reposWithErrors"`

    // Stack information (keyed by repo name)
    Stacks map[string]*Stack `json:"stacks"`

    // Metadata
    TotalReposScanned int       `json:"totalReposScanned"`
    TotalPRsFound     int       `json:"totalPRsFound"`
    ScanDuration      time.Duration `json:"scanDuration"`
    Username          string    `json:"username"`
}
```

---

## 5. Configuration System

### 5.1 Config Schema (`internal/config/types.go`)

```go
package config

// Config represents the complete prt configuration
type Config struct {
    // Identity
    GitHubUsername string `yaml:"github_username" mapstructure:"github_username"`

    // Team
    TeamMembers []string `yaml:"team_members" mapstructure:"team_members"`

    // Repository Discovery
    SearchPaths  []string `yaml:"search_paths" mapstructure:"search_paths"`
    IncludeRepos []string `yaml:"include_repos" mapstructure:"include_repos"`
    ScanDepth    int      `yaml:"scan_depth" mapstructure:"scan_depth"`

    // Known Bots
    Bots []string `yaml:"bots" mapstructure:"bots"`

    // Display
    DefaultGroupBy string `yaml:"default_group_by" mapstructure:"default_group_by"`
    DefaultSort    string `yaml:"default_sort" mapstructure:"default_sort"`
    ShowBranchName bool   `yaml:"show_branch_name" mapstructure:"show_branch_name"`
    ShowIcons      bool   `yaml:"show_icons" mapstructure:"show_icons"`
}

// GroupBy options
const (
    GroupByProject = "project"
    GroupByAuthor  = "author"
)

// Sort options
const (
    SortOldest = "oldest"
    SortNewest = "newest"
)
```

### 5.2 Defaults (`internal/config/defaults.go`)

```go
package config

var DefaultConfig = Config{
    GitHubUsername: "", // Must be set or auto-detected
    TeamMembers:    []string{},
    SearchPaths:    []string{},
    IncludeRepos:   []string{}, // Empty = match all
    ScanDepth:      3,
    Bots: []string{
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
    },
    DefaultGroupBy: GroupByProject,
    DefaultSort:    SortOldest,
    ShowBranchName: true,
    ShowIcons:      true,
}

// ConfigDir returns the prt config directory path
func ConfigDir() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".prt")
}

// ConfigPath returns the full config file path
func ConfigPath() string {
    return filepath.Join(ConfigDir(), "config.yaml")
}
```

### 5.3 Config File Template

Generated on first run with comments:

```yaml
# PRT Configuration
# https://github.com/ChrisEdwards/prt

# Your GitHub username (required)
# Used to identify PRs you authored and review requests for you
# Auto-detected if left empty (via `gh api user`)
github_username: ""

# Team members (GitHub usernames)
# PRs from these users are highlighted as "Team PRs"
team_members:
  # - "teammate1"
  # - "teammate2"

# Directories to search for Git repositories
# Supports absolute paths and ~ for home directory
search_paths:
  # - "~/code/work"
  # - "~/projects"

# Repository name patterns to include (glob syntax)
# Leave empty to include all discovered repositories
# Examples: "myorg-*", "*-api", "frontend"
include_repos:
  # - "myorg-*"

# Maximum directory depth when searching for repositories
# Default: 3
scan_depth: 3

# Known bot accounts (PRs from these are de-prioritized)
# Pre-populated with common bots; add your org's bots here
bots:
  - "dependabot[bot]"
  - "dependabot"
  - "renovate[bot]"
  - "renovate"
  - "github-actions[bot]"
  - "codecov[bot]"
  - "codecov"
  - "semantic-release-bot"
  - "greenkeeper[bot]"
  - "snyk-bot"
  - "imgbot[bot]"
  - "allcontributors[bot]"
  - "mergify[bot]"
  - "kodiakhq[bot]"
  - "stale[bot]"

# Default grouping: "project" or "author"
default_group_by: "project"

# Default sort order: "oldest" or "newest" (by creation date)
default_sort: "oldest"

# Show branch names in PR output
show_branch_name: true

# Show icons (requires a Nerd Font or emoji support)
show_icons: true
```

### 5.4 Config Loading Logic

```go
// Precedence: CLI Flags > Environment > Config File > Defaults
func Load(flags *Flags) (*Config, error) {
    v := viper.New()

    // 1. Set defaults
    v.SetDefault("scan_depth", DefaultConfig.ScanDepth)
    v.SetDefault("default_group_by", DefaultConfig.DefaultGroupBy)
    // ... etc

    // 2. Load config file if exists
    v.SetConfigName("config")
    v.SetConfigType("yaml")
    v.AddConfigPath(ConfigDir())

    if err := v.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, fmt.Errorf("error reading config: %w", err)
        }
        // Config not found - will be created by wizard
    }

    // 3. Environment variables (PRT_GITHUB_USERNAME, etc.)
    v.SetEnvPrefix("PRT")
    v.AutomaticEnv()

    // 4. CLI flag overrides
    if flags.Path != "" {
        v.Set("search_paths", []string{flags.Path})
    }
    if flags.Depth > 0 {
        v.Set("scan_depth", flags.Depth)
    }
    // ... etc

    // 5. Unmarshal
    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
```

---

## 6. Core Components

### 6.1 GitHub Client (`internal/github/client.go`)

#### gh CLI JSON Fields

```bash
gh pr list --json number,title,url,author,state,isDraft,createdAt,baseRefName,headRefName,statusCheckRollup,reviewRequests,assignees,reviews,labels
```

#### Client Interface

```go
package github

type Client interface {
    // Check checks if gh CLI is available and authenticated
    Check() error

    // GetCurrentUser returns the authenticated GitHub username
    GetCurrentUser() (string, error)

    // ListPRs fetches all open/draft PRs for a repository
    ListPRs(repoPath string) ([]*models.PR, error)
}

type client struct {
    retryConfig RetryConfig
}

type RetryConfig struct {
    MaxAttempts int           // Default: 3
    InitialWait time.Duration // Default: 1s
    MaxWait     time.Duration // Default: 10s
}

func NewClient(cfg RetryConfig) Client {
    return &client{retryConfig: cfg}
}
```

#### Implementation Details

```go
func (c *client) Check() error {
    // Check gh exists
    _, err := exec.LookPath("gh")
    if err != nil {
        return &GHNotFoundError{
            Message: "GitHub CLI (gh) not found. Please install it:\n\n  brew install gh\n\nThen authenticate:\n\n  gh auth login",
        }
    }

    // Check authentication
    cmd := exec.Command("gh", "auth", "status")
    if err := cmd.Run(); err != nil {
        return &GHAuthError{
            Message: "GitHub CLI is not authenticated. Please run:\n\n  gh auth login",
        }
    }

    return nil
}

func (c *client) GetCurrentUser() (string, error) {
    cmd := exec.Command("gh", "api", "user", "--jq", ".login")
    out, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to get current user: %w", err)
    }
    return strings.TrimSpace(string(out)), nil
}

func (c *client) ListPRs(repoPath string) ([]*models.PR, error) {
    // Execute with retry
    return c.withRetry(func() ([]*models.PR, error) {
        cmd := exec.Command("gh", "pr", "list",
            "--json", "number,title,url,author,state,isDraft,createdAt,baseRefName,headRefName,statusCheckRollup,reviewRequests,assignees,reviews",
            "--state", "open",
        )
        cmd.Dir = repoPath

        out, err := cmd.Output()
        if err != nil {
            return nil, c.classifyError(err, repoPath)
        }

        return ParsePRList(out)
    })
}
```

### 6.2 Repository Scanner (`internal/scanner/scanner.go`)

```go
package scanner

type Scanner interface {
    // Scan discovers Git repositories in configured paths
    Scan(cfg *config.Config) ([]*models.Repository, error)
}

type scanner struct {
    maxDepth    int
    includeGlob []glob.Glob
}

func NewScanner(maxDepth int, includePatterns []string) (Scanner, error) {
    globs := make([]glob.Glob, 0, len(includePatterns))
    for _, pattern := range includePatterns {
        g, err := glob.Compile(pattern)
        if err != nil {
            return nil, fmt.Errorf("invalid pattern %q: %w", pattern, err)
        }
        globs = append(globs, g)
    }
    return &scanner{maxDepth: maxDepth, includeGlob: globs}, nil
}

func (s *scanner) Scan(cfg *config.Config) ([]*models.Repository, error) {
    var repos []*models.Repository
    seen := make(map[string]bool)

    for _, searchPath := range cfg.SearchPaths {
        expanded := expandPath(searchPath)

        err := filepath.WalkDir(expanded, func(path string, d fs.DirEntry, err error) error {
            if err != nil {
                return nil // Skip inaccessible directories
            }

            // Check depth
            depth := countDepth(expanded, path)
            if depth > s.maxDepth {
                if d.IsDir() {
                    return filepath.SkipDir
                }
                return nil
            }

            // Skip symlinks
            if d.Type()&fs.ModeSymlink != 0 {
                return filepath.SkipDir
            }

            // Look for .git directories
            if d.IsDir() && d.Name() == ".git" {
                repoPath := filepath.Dir(path)
                if seen[repoPath] {
                    return filepath.SkipDir
                }
                seen[repoPath] = true

                repo, err := s.inspectRepo(repoPath)
                if err != nil {
                    // Log warning, continue
                    return filepath.SkipDir
                }

                if s.matchesIncludePatterns(repo.Name) {
                    repos = append(repos, repo)
                }

                return filepath.SkipDir
            }

            return nil
        })

        if err != nil {
            return nil, err
        }
    }

    return repos, nil
}

func (s *scanner) matchesIncludePatterns(name string) bool {
    if len(s.includeGlob) == 0 {
        return true // No patterns = match all
    }
    for _, g := range s.includeGlob {
        if g.Match(name) {
            return true
        }
    }
    return false
}

func (s *scanner) inspectRepo(path string) (*models.Repository, error) {
    // Get remote URL
    cmd := exec.Command("git", "remote", "get-url", "origin")
    cmd.Dir = path
    out, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("no origin remote: %w", err)
    }

    remoteURL := strings.TrimSpace(string(out))
    owner, name := parseGitHubRemote(remoteURL)

    if owner == "" || name == "" {
        return nil, fmt.Errorf("not a GitHub repository")
    }

    return &models.Repository{
        Name:      name,
        Path:      path,
        RemoteURL: remoteURL,
        Owner:     owner,
    }, nil
}
```

### 6.3 Stack Detector (`internal/stacks/detector.go`)

```go
package stacks

type Detector interface {
    // DetectStacks analyzes PRs and builds stack trees
    DetectStacks(prs []*models.PR) *models.Stack
}

type detector struct{}

func NewDetector() Detector {
    return &detector{}
}

func (d *detector) DetectStacks(prs []*models.PR) *models.Stack {
    if len(prs) == 0 {
        return &models.Stack{}
    }

    // Build maps for efficient lookup
    // Key: branch name, Value: PR that has this as head branch
    headBranchToPR := make(map[string]*models.PR)
    for _, pr := range prs {
        headBranchToPR[pr.HeadBranch] = pr
    }

    // Create nodes for all PRs
    nodes := make(map[int]*models.StackNode)
    for _, pr := range prs {
        nodes[pr.Number] = &models.StackNode{
            PR:       pr,
            Children: []*models.StackNode{},
        }
    }

    // Build parent-child relationships
    // If PR_A targets branch X, and PR_B's head IS branch X, then PR_B is parent of PR_A
    for _, pr := range prs {
        // Does another PR have its head branch as our base branch?
        if parentPR, ok := headBranchToPR[pr.BaseBranch]; ok {
            parentNode := nodes[parentPR.Number]
            childNode := nodes[pr.Number]

            childNode.Parent = parentNode
            parentNode.Children = append(parentNode.Children, childNode)
        }
    }

    // Identify roots (PRs with no parent in our set)
    var roots []*models.StackNode
    var allNodes []*models.StackNode

    for _, node := range nodes {
        allNodes = append(allNodes, node)
        if node.Parent == nil {
            roots = append(roots, node)
        }
    }

    // Calculate depths
    for _, root := range roots {
        setDepths(root, 0)
    }

    return &models.Stack{
        Roots:    roots,
        AllNodes: allNodes,
    }
}

func setDepths(node *models.StackNode, depth int) {
    node.Depth = depth
    for _, child := range node.Children {
        setDepths(child, depth+1)
    }
}
```

### 6.4 PR Categorizer (`internal/categorizer/categorizer.go`)

```go
package categorizer

type Categorizer interface {
    Categorize(repos []*models.Repository, cfg *config.Config, username string) *models.ScanResult
}

type categorizer struct{}

func NewCategorizer() Categorizer {
    return &categorizer{}
}

func (c *categorizer) Categorize(repos []*models.Repository, cfg *config.Config, username string) *models.ScanResult {
    result := &models.ScanResult{
        MyPRs:            []*models.PR{},
        NeedsMyAttention: []*models.PR{},
        TeamPRs:          []*models.PR{},
        OtherPRs:         []*models.PR{},
        ReposWithPRs:     []*models.Repository{},
        ReposWithoutPRs:  []*models.Repository{},
        ReposWithErrors:  []*models.Repository{},
        Stacks:           make(map[string]*models.Stack),
        Username:         username,
    }

    teamSet := toSet(cfg.TeamMembers)
    botSet := toSet(cfg.Bots)

    for _, repo := range repos {
        if repo.ScanError != nil {
            result.ReposWithErrors = append(result.ReposWithErrors, repo)
            continue
        }

        if !repo.HasPRs() {
            result.ReposWithoutPRs = append(result.ReposWithoutPRs, repo)
            continue
        }

        result.ReposWithPRs = append(result.ReposWithPRs, repo)
        result.TotalPRsFound += len(repo.PRs)

        // Detect stacks for this repo
        detector := stacks.NewDetector()
        result.Stacks[repo.Name] = detector.DetectStacks(repo.PRs)

        // Categorize each PR
        for _, pr := range repo.PRs {
            // Set repo context
            pr.RepoName = repo.Name
            pr.RepoPath = repo.Path

            // Compute user-specific fields
            pr.IsReviewRequestedFromMe = contains(pr.ReviewRequests, username)
            pr.IsAssignedToMe = contains(pr.Assignees, username)
            pr.MyReviewStatus = findMyReviewStatus(pr.Reviews, username)

            // Categorize
            switch {
            case pr.Author == username:
                result.MyPRs = append(result.MyPRs, pr)

            case pr.IsReviewRequestedFromMe || pr.IsAssignedToMe:
                // Only add to "needs attention" if I haven't approved
                if pr.MyReviewStatus != models.ReviewStateApproved {
                    result.NeedsMyAttention = append(result.NeedsMyAttention, pr)
                } else {
                    // I approved, treat as team PR (if team member) or other
                    if teamSet[pr.Author] {
                        result.TeamPRs = append(result.TeamPRs, pr)
                    } else {
                        result.OtherPRs = append(result.OtherPRs, pr)
                    }
                }

            case teamSet[pr.Author]:
                result.TeamPRs = append(result.TeamPRs, pr)

            case botSet[pr.Author]:
                result.OtherPRs = append(result.OtherPRs, pr)

            default:
                result.OtherPRs = append(result.OtherPRs, pr)
            }
        }
    }

    result.TotalReposScanned = len(repos)

    // Sort each category
    sortPRs(result.MyPRs, cfg.DefaultSort)
    sortPRs(result.NeedsMyAttention, cfg.DefaultSort)
    sortPRs(result.TeamPRs, cfg.DefaultSort)
    sortPRs(result.OtherPRs, cfg.DefaultSort)

    return result
}

func findMyReviewStatus(reviews []models.Review, username string) models.ReviewState {
    // Find the most recent review from this user
    var latest *models.Review
    for i := range reviews {
        r := &reviews[i]
        if r.Author == username {
            if latest == nil || r.Submitted.After(latest.Submitted) {
                latest = r
            }
        }
    }
    if latest == nil {
        return models.ReviewStateNone
    }
    return latest.State
}
```

---

## 7. Implementation Phases

### Phase 1: Foundation (Days 1-2)

**Goal**: Basic project structure, config system, and gh integration

| Task | File(s) | Acceptance Criteria |
|------|---------|---------------------|
| Initialize Go module | `go.mod` | `go build` succeeds |
| Create directory structure | All dirs | Structure matches spec |
| Implement config types | `internal/config/types.go` | Compiles |
| Implement defaults | `internal/config/defaults.go` | All defaults defined |
| Implement config loading | `internal/config/config.go` | Loads YAML, applies precedence |
| Implement gh check | `internal/github/client.go` | Detects gh, auth status |
| Implement get user | `internal/github/client.go` | Returns GitHub username |

**Deliverable**: `prt` binary that loads config and prints username

### Phase 2: Repository Discovery (Days 2-3)

**Goal**: Find Git repos in configured paths

| Task | File(s) | Acceptance Criteria |
|------|---------|---------------------|
| Implement directory walker | `internal/scanner/scanner.go` | Respects depth, no symlinks |
| Implement .git detection | `internal/scanner/scanner.go` | Finds repos |
| Implement remote parsing | `internal/scanner/git.go` | Extracts owner/name |
| Implement glob matching | `internal/scanner/filter.go` | Filters by pattern |

**Deliverable**: `prt` lists discovered repos

### Phase 3: PR Fetching (Days 3-4)

**Goal**: Fetch PRs concurrently with progress

| Task | File(s) | Acceptance Criteria |
|------|---------|---------------------|
| Implement PR list parsing | `internal/github/parser.go` | Parses gh JSON |
| Implement retry logic | `internal/github/retry.go` | 3 retries, backoff |
| Implement concurrent fetch | `internal/github/client.go` | Goroutines + channels |
| Implement progress bar | `internal/display/progress.go` | Shows repo count |

**Deliverable**: `prt` fetches PRs with progress bar

### Phase 4: Stack Detection (Day 4)

**Goal**: Detect and build PR stacks

| Task | File(s) | Acceptance Criteria |
|------|---------|---------------------|
| Implement stack detection | `internal/stacks/detector.go` | Finds parent-child |
| Implement tree building | `internal/stacks/builder.go` | Builds StackNode tree |
| Handle orphaned children | `internal/stacks/detector.go` | Marks orphans |

**Deliverable**: Stack trees built for each repo

### Phase 5: Categorization (Day 5)

**Goal**: Categorize PRs by user role

| Task | File(s) | Acceptance Criteria |
|------|---------|---------------------|
| Implement categorizer | `internal/categorizer/categorizer.go` | All 4 categories |
| Implement sorter | `internal/categorizer/sorter.go` | By date |
| Compute user fields | `internal/categorizer/categorizer.go` | Review status, etc. |

**Deliverable**: PRs correctly categorized

### Phase 6: Display (Days 5-7)

**Goal**: Beautiful terminal output

| Task | File(s) | Acceptance Criteria |
|------|---------|---------------------|
| Define styles | `internal/display/styles.go` | All colors defined |
| Implement section rendering | `internal/display/sections.go` | 5 sections |
| Implement tree rendering | `internal/display/tree.go` | Unicode box chars |
| Implement PR row rendering | `internal/display/renderer.go` | All fields shown |
| Implement empty states | `internal/display/sections.go` | "None" messages |
| Implement JSON output | `internal/display/json.go` | Valid JSON |

**Deliverable**: Full styled output

### Phase 7: CLI & Wizard (Days 7-8)

**Goal**: Complete CLI with all flags and setup wizard

| Task | File(s) | Acceptance Criteria |
|------|---------|---------------------|
| Implement root command | `internal/cli/root.go` | Runs full flow |
| Implement all flags | `internal/cli/flags.go` | All flags work |
| Implement wizard | `internal/cli/wizard.go` | Creates valid config |
| Wire everything | `cmd/prt/main.go` | End-to-end works |

**Deliverable**: Complete `prt` CLI

### Phase 8: Testing & Polish (Days 8-10)

**Goal**: Tests, error messages, documentation

| Task | File(s) | Acceptance Criteria |
|------|---------|---------------------|
| Unit tests | `*_test.go` | 80%+ coverage |
| Integration tests | `internal/cli/*_test.go` | Happy path tested |
| Error message review | All | Clear, actionable |
| README | `README.md` | Install & usage |

**Deliverable**: Release-ready binary

---

## 8. CLI Interface

### 8.1 Command Structure

```
prt [flags]

Flags:
  -p, --path string      Override search paths (can be specified multiple times)
  -f, --filter string    Filter repositories by name pattern (glob)
  -g, --group string     Group PRs by: project, author (default: from config)
  -d, --depth int        Maximum directory scan depth (default: from config)
      --json             Output as JSON
      --no-color         Disable colored output
  -h, --help             Show help

Examples:
  prt                           # Run with config defaults
  prt -p ~/code/work            # Scan specific path
  prt -f "myorg-*"              # Only repos matching pattern
  prt -g author                 # Group by author instead of project
  prt --json                    # Output JSON for scripting
  prt --json | jq '.myPRs'      # Filter JSON output
```

### 8.2 Interactive Setup Wizard

Triggered on first run when no config exists:

```
Welcome to PRT (PR Tracker)! 🚀

Let's set up your configuration.

? What is your GitHub username? (leave blank to auto-detect)
> jdoe

✓ Detected username: jdoe

? Where should PRT look for repositories?
  (Enter paths separated by commas, ~ supported)
> ~/code/work, ~/projects/oss

? Any repository name patterns to include? (glob syntax, blank for all)
>

? Add team members? (GitHub usernames, comma-separated)
> sarah-dev, mike-frontend, tom-backend

Configuration saved to ~/.prt/config.yaml

Run `prt` to see your PR dashboard!
```

### 8.3 Help Text

```
PRT - GitHub PR Tracker

Aggregate and visualize GitHub Pull Request status across multiple
local repositories. Highlights PRs requiring your attention and
shows stacked PR relationships.

Usage:
  prt [flags]

Flags:
  -p, --path strings     Search paths (overrides config)
  -f, --filter string    Filter repos by name pattern
  -g, --group string     Group by: project, author
  -d, --depth int        Scan depth (default 3)
      --json             JSON output
      --no-color         Disable colors
  -h, --help             Show this help

Configuration:
  Config file: ~/.prt/config.yaml
  First run will launch setup wizard.

Examples:
  prt                        Show all PRs
  prt -p ~/work              Scan specific path
  prt -f "api-*"             Only api-* repos
  prt --json | jq            Pipe to jq

More info: https://github.com/ChrisEdwards/prt
```

---

## 9. Display & Styling

### 9.1 Style Definitions (`internal/display/styles.go`)

```go
package display

import "github.com/charmbracelet/lipgloss"

var (
    // Section headers
    HeaderStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("15")). // White
        Background(lipgloss.Color("57")). // Purple
        Padding(0, 1)

    SubheaderStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("244")) // Gray

    // PR States
    DraftStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("244")). // Gray
        Italic(true)

    NeedsReviewStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("46")) // Green

    ApprovedStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("39")) // Blue

    ChangesRequestedStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("214")) // Orange

    BlockedStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("244")). // Gray
        Faint(true)

    MergedStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("93")). // Purple
        Strikethrough(true)

    // CI Status
    CIPassingStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("46")) // Green

    CIFailingStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("196")) // Red

    CIPendingStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("226")) // Yellow

    // Priority
    HighPriorityStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("46")) // Green

    TeamStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("15")) // White

    OtherStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("244")). // Gray
        Faint(true)

    // URLs
    URLStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("39")). // Blue
        Underline(true)

    // Tree
    TreeBranchStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("240")) // Dark gray

    // Empty state
    EmptyStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("244")).
        Italic(true)

    // Meta info
    MetaStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("244"))
)

// Icons (when enabled)
const (
    IconPR           = "📋"
    IconDraft        = "📝"
    IconMerged       = "🟣"
    IconApproved     = "✅"
    IconChanges      = "🔄"
    IconReview       = "👀"
    IconBlocked      = "🔒"
    IconCIPassing    = "✅"
    IconCIFailing    = "❌"
    IconCIPending    = "⏳"
    IconRepo         = "📦"
    IconNoContent    = "∅"

    // Tree characters
    TreeVertical     = "│"
    TreeBranch       = "├──"
    TreeLastBranch   = "└──"
    TreeIndent       = "    "
)
```

### 9.2 Output Format

```
PRT ══════════════════════════════════════════════════════════════════

📋 MY PRS ────────────────────────────────────────────────────────────

  contrast-sdk
  ├── #402 Feature: Transformer Logic
  │   Draft · Created 4d ago · CI ✅
  │   feature/transformer → main
  │   https://github.com/org/contrast-sdk/pull/402
  │
  └── #398 Refactor: API endpoints
      Waiting review · Created 1w ago · CI ✅ · 1 approval
      refactor/api → main
      https://github.com/org/contrast-sdk/pull/398

👀 NEEDS MY ATTENTION ────────────────────────────────────────────────

  aiml-core
  └── #156 Fix: Memory leak in training loop
      Review requested · Created 2d ago · CI ❌
      @sarah-dev · fix/memory-leak → main
      https://github.com/org/aiml-core/pull/156

👥 TEAM PRS ──────────────────────────────────────────────────────────

  contrast-sdk
  ├── #405 Fix: Unit tests for transformer (blocked by #402)
  │   Draft · Created 2h ago · CI ❌
  │   @sarah-dev · fix/transformer-tests → feature/transformer
  │   https://github.com/org/contrast-sdk/pull/405
  │
  └── #399 Docs: Update README
      Waiting review · Created 5d ago · CI ✅
      @tom-backend · docs/readme → main
      https://github.com/org/contrast-sdk/pull/399
      ✓ You approved

🤖 OTHER PRS ─────────────────────────────────────────────────────────

  contrast-sdk
  └── #410 chore(deps): bump lodash from 4.17.20 to 4.17.21
      Created 1d ago · CI ✅
      @dependabot[bot]
      https://github.com/org/contrast-sdk/pull/410

📂 REPOS WITH NO OPEN PRS ────────────────────────────────────────────

  • aiml-tools (/Users/jdoe/code/work/aiml-tools)
  • contrast-cli (/Users/jdoe/code/work/contrast-cli)

══════════════════════════════════════════════════════════════════════
Scanned 5 repos · Found 6 PRs · 2.3s
```

### 9.3 Empty State Examples

```
📋 MY PRS ────────────────────────────────────────────────────────────
  None

👀 NEEDS MY ATTENTION ────────────────────────────────────────────────
  None - you're all caught up! 🎉
```

### 9.4 Progress Bar During Fetch

```
Scanning repositories...

  ████████████░░░░░░░░ 60% (3/5 repos)

  ✓ contrast-sdk (4 PRs)
  ✓ aiml-core (1 PR)
  ✓ aiml-tools (0 PRs)
  ⠋ contrast-cli...
```

---

## 10. Error Handling

### 10.1 Error Types

```go
package github

// GHNotFoundError indicates gh CLI is not installed
type GHNotFoundError struct {
    Message string
}

func (e *GHNotFoundError) Error() string {
    return e.Message
}

// GHAuthError indicates gh is not authenticated
type GHAuthError struct {
    Message string
}

func (e *GHAuthError) Error() string {
    return e.Message
}

// RepoScanError indicates a repository-specific failure
type RepoScanError struct {
    RepoPath string
    Cause    error
}

func (e *RepoScanError) Error() string {
    return fmt.Sprintf("failed to scan %s: %v", e.RepoPath, e.Cause)
}

// NetworkError indicates a network-related failure
type NetworkError struct {
    Cause       error
    Retries     int
    LastAttempt time.Time
}

func (e *NetworkError) Error() string {
    return fmt.Sprintf("network error after %d retries: %v", e.Retries, e.Cause)
}
```

### 10.2 Error Handling Matrix

| Error | Detection | Action | User Message |
|-------|-----------|--------|--------------|
| gh not found | `exec.LookPath` | Exit immediately | "GitHub CLI (gh) not found. Install: `brew install gh`" |
| gh not authed | `gh auth status` | Exit immediately | "Please authenticate: `gh auth login`" |
| No username in config + auto-detect fails | Config validation | Exit immediately | "Could not determine GitHub username. Set in ~/.prt/config.yaml" |
| No search paths | Config validation | Exit immediately | "No search paths configured. Run `prt` to launch setup wizard." |
| No repos found | After scan | Exit with message | "No Git repositories found in configured paths." |
| Repo not GitHub | Remote parsing | Skip, log warning | (internal warning only) |
| Network timeout | gh execution | Retry 3x, backoff | "Network error. Retrying... (2/3)" |
| Network failure after retries | gh execution | Exit | "Network error persisted after 3 retries. Please check connection." |
| Rate limit | gh execution | Exit | "GitHub API rate limit reached. Wait and retry." |
| Repo auth error | gh execution | Skip, continue | Warning: "Skipped repo-name: authentication error" |

### 10.3 Retry Implementation

```go
func (c *client) withRetry(fn func() ([]*models.PR, error)) ([]*models.PR, error) {
    var lastErr error

    for attempt := 1; attempt <= c.retryConfig.MaxAttempts; attempt++ {
        result, err := fn()
        if err == nil {
            return result, nil
        }

        // Don't retry non-network errors
        if !isNetworkError(err) {
            return nil, err
        }

        lastErr = err

        if attempt < c.retryConfig.MaxAttempts {
            wait := c.calculateBackoff(attempt)
            time.Sleep(wait)
        }
    }

    return nil, &NetworkError{
        Cause:   lastErr,
        Retries: c.retryConfig.MaxAttempts,
    }
}

func (c *client) calculateBackoff(attempt int) time.Duration {
    wait := c.retryConfig.InitialWait * time.Duration(1<<(attempt-1))
    if wait > c.retryConfig.MaxWait {
        wait = c.retryConfig.MaxWait
    }
    return wait
}
```

---

## 11. Testing Strategy

### 11.1 Unit Tests

| Package | Test Focus | Mocking Strategy |
|---------|------------|------------------|
| `config` | Load, save, defaults, precedence | File system (temp files) |
| `scanner` | Directory walking, filtering, depth | File system (temp dirs) |
| `github` | Parsing, retry logic | Mock exec commands |
| `stacks` | Detection algorithm, edge cases | Pure logic (no mocks) |
| `categorizer` | All categorization rules | Pure logic (no mocks) |
| `display` | Output formatting | Snapshot testing |

### 11.2 Key Test Cases

**Config Tests:**
- Load valid YAML
- Handle missing file (trigger wizard)
- CLI flag overrides config
- Environment variable overrides
- Invalid YAML produces error

**Scanner Tests:**
- Find repos at various depths
- Respect depth limit
- Skip symlinks
- Apply glob filters
- Handle permission errors gracefully

**GitHub Tests:**
- Parse valid PR JSON
- Handle empty response
- Retry on network error
- Don't retry on auth error
- Classify error types correctly

**Stack Tests:**
- Single PR (no stack)
- Two-level stack
- Three+ level stack
- Multiple independent stacks
- Orphaned child (merged parent)
- No false positives on branch names

**Categorizer Tests:**
- My PR detection
- Review requested detection
- Assigned detection
- Team member detection
- Bot detection
- Approved vs needs attention
- Sort oldest/newest

### 11.3 Integration Tests

```go
func TestFullScanFlow(t *testing.T) {
    // Setup: temp config, mock gh responses
    // Execute: run full scan
    // Assert: output matches expected structure
}

func TestWizardFlow(t *testing.T) {
    // Setup: no config file
    // Execute: simulate wizard inputs
    // Assert: config file created correctly
}
```

### 11.4 Test Commands

```makefile
test:
	go test ./... -v

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

test-race:
	go test ./... -race
```

---

## 12. Build & Distribution

### 12.1 Makefile

```makefile
.PHONY: build test clean install release

VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS = -ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/prt ./cmd/prt

test:
	go test ./... -v

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

clean:
	rm -rf bin/
	rm -f coverage.out

install: build
	cp bin/prt /usr/local/bin/

# Build for multiple platforms
release:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/prt-darwin-amd64 ./cmd/prt
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/prt-darwin-arm64 ./cmd/prt
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/prt-linux-amd64 ./cmd/prt

lint:
	golangci-lint run

fmt:
	go fmt ./...
```

### 12.2 Release Artifacts

For each release, produce:

| Artifact | Platform | Notes |
|----------|----------|-------|
| `prt-darwin-amd64` | macOS Intel | Primary target |
| `prt-darwin-arm64` | macOS Apple Silicon | Primary target |
| `prt-linux-amd64` | Linux | Secondary target |
| `checksums.txt` | All | SHA256 sums |

### 12.3 Version Embedding

```go
// cmd/prt/main.go
package main

var version = "dev" // Set by ldflags

func main() {
    rootCmd.Version = version
    // ...
}
```

### 12.4 Homebrew Installation

#### Homebrew Tap Setup

Use the existing Homebrew tap repository: `ChrisEdwards/homebrew-tap` (https://github.com/ChrisEdwards/homebrew-tap)

```
homebrew-tap/
├── Formula/
│   └── prt.rb
└── README.md
```

#### Formula (`Formula/prt.rb`)

```ruby
class Prt < Formula
  desc "GitHub PR Tracker - Aggregate PR status across multiple repositories"
  homepage "https://github.com/ChrisEdwards/prt"
  version "1.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/ChrisEdwards/prt/releases/download/v#{version}/prt-darwin-arm64.tar.gz"
      sha256 "REPLACE_WITH_ARM64_SHA256"
    else
      url "https://github.com/ChrisEdwards/prt/releases/download/v#{version}/prt-darwin-amd64.tar.gz"
      sha256 "REPLACE_WITH_AMD64_SHA256"
    end
  end

  on_linux do
    url "https://github.com/ChrisEdwards/prt/releases/download/v#{version}/prt-linux-amd64.tar.gz"
    sha256 "REPLACE_WITH_LINUX_SHA256"
  end

  depends_on "gh" => :recommended

  def install
    bin.install "prt"
  end

  def caveats
    <<~EOS
      PRT requires the GitHub CLI (gh) to be installed and authenticated.

      If not already installed:
        brew install gh

      Then authenticate:
        gh auth login

      Run `prt` to start the setup wizard.
    EOS
  end

  test do
    assert_match "prt version", shell_output("#{bin}/prt --version")
  end
end
```

#### User Installation

```bash
# Add the tap (one-time)
brew tap ChrisEdwards/tap

# Install prt
brew install prt

# Or in one command
brew install ChrisEdwards/tap/prt
```

#### Updating the Formula

When releasing a new version:

1. **Build release artifacts**
   ```bash
   make release
   ```

2. **Create tarballs**
   ```bash
   cd bin
   tar -czf prt-darwin-amd64.tar.gz prt-darwin-amd64
   tar -czf prt-darwin-arm64.tar.gz prt-darwin-arm64
   tar -czf prt-linux-amd64.tar.gz prt-linux-amd64
   ```

3. **Generate SHA256 checksums**
   ```bash
   shasum -a 256 *.tar.gz > checksums.txt
   ```

4. **Create GitHub release** with tarballs attached

5. **Update formula** with new version and SHA256 values

#### Automating Releases with GoReleaser (Optional)

Add `.goreleaser.yaml` to automate the entire release process:

```yaml
# .goreleaser.yaml
version: 2

project_name: prt

before:
  hooks:
    - go mod tidy

builds:
  - main: ./cmd/prt
    binary: prt
    env:
      - CGO_ENABLED=0
    goos:
      - darwin
      - linux
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}-{{ .Os }}-{{ .Arch }}"

checksum:
  name_template: "checksums.txt"

release:
  github:
    owner: ChrisEdwards
    name: prt

brews:
  - name: prt
    repository:
      owner: ChrisEdwards
      name: homebrew-tap
    folder: Formula
    homepage: "https://github.com/ChrisEdwards/prt"
    description: "GitHub PR Tracker - Aggregate PR status across multiple repositories"
    license: "MIT"
    dependencies:
      - name: gh
        type: recommended
    caveats: |
      PRT requires the GitHub CLI (gh) to be installed and authenticated.

      If not already installed:
        brew install gh

      Then authenticate:
        gh auth login

      Run `prt` to start the setup wizard.
    test: |
      assert_match "prt version", shell_output("#{bin}/prt --version")
```

Then release with:
```bash
git tag v1.0.0
git push origin v1.0.0
goreleaser release --clean
```

GoReleaser will:
- Build binaries for all platforms
- Create tarballs
- Generate checksums
- Create GitHub release
- Automatically update the Homebrew formula

### 12.5 Installation Methods Summary

| Method | Command | Best For |
|--------|---------|----------|
| **Homebrew** | `brew install ChrisEdwards/tap/prt` | Team members on macOS |
| **Go Install** | `go install github.com/ChrisEdwards/prt@latest` | Developers with Go |
| **Binary** | Download from releases | CI/CD, non-Homebrew users |

---

## Appendix A: gh CLI JSON Schema

Response from `gh pr list --json ...`:

```json
[
  {
    "number": 402,
    "title": "Feature: Transformer Logic",
    "url": "https://github.com/org/repo/pull/402",
    "author": {
      "login": "jdoe"
    },
    "state": "OPEN",
    "isDraft": true,
    "createdAt": "2024-12-15T10:30:00Z",
    "baseRefName": "main",
    "headRefName": "feature/transformer",
    "statusCheckRollup": [
      {
        "context": "ci/build",
        "state": "SUCCESS"
      },
      {
        "context": "ci/test",
        "state": "PENDING"
      }
    ],
    "reviewRequests": [
      {
        "login": "reviewer1"
      }
    ],
    "assignees": [
      {
        "login": "assignee1"
      }
    ],
    "reviews": [
      {
        "author": {
          "login": "reviewer1"
        },
        "state": "APPROVED",
        "submittedAt": "2024-12-16T14:00:00Z"
      }
    ]
  }
]
```

---

## Appendix B: Glossary

| Term | Definition |
|------|------------|
| **Stacked PR** | A PR whose base branch is another PR's head branch |
| **Parent PR** | In a stack, the PR that must merge first |
| **Child PR** | In a stack, a PR dependent on its parent |
| **Orphaned child** | A child PR whose parent has been merged |
| **Review requested** | GitHub explicitly asked you to review |
| **Assigned** | You're listed as an assignee on the PR |

---

## Appendix C: Future Considerations (Out of Scope for V1)

- Watch mode (auto-refresh)
- Desktop notifications
- Caching with TTL
- GitLab/Bitbucket support
- Browser extension companion
- Slack integration
- Custom themes

---

*End of Implementation Specification*
