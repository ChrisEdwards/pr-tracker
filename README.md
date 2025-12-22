# PRT - GitHub PR Tracker

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)

> Aggregate GitHub PR status across multiple local repositories

PRT solves "PR Fatigue" - the cognitive burden of tracking PRs across many repos. With a single command, see:

- **My PRs** - PRs you authored, waiting for review
- **Needs My Attention** - PRs requesting your review or assigned to you
- **Team PRs** - PRs from your configured team members
- **Stacked PRs** - Visual tree of dependent PR chains

![PRT Preview](assets/prt-preview.png)

## Features

- **Multi-repo scanning** - Discover Git repos in configured directories
- **Smart categorization** - PRs sorted by your relationship to them
- **Stack detection** - Visualize dependent PR chains (stacked PRs)
- **Bot filtering** - Auto-deprioritize dependabot, renovate, etc.
- **Beautiful output** - Styled terminal UI with icons
- **JSON output** - Pipe to `jq` for scripting
- **Zero config** - Setup wizard on first run

## Installation

### Homebrew (macOS)

```bash
brew tap ChrisEdwards/tap
brew install prt
```

### Binary Download

Download pre-built binaries from [Releases](https://github.com/ChrisEdwards/pr-tracker/releases).

Available platforms:
- macOS (Intel): `prt_<version>_darwin_amd64.tar.gz`
- macOS (Apple Silicon): `prt_<version>_darwin_arm64.tar.gz`
- Linux (x64): `prt_<version>_linux_amd64.tar.gz`
- Linux (arm64): `prt_<version>_linux_arm64.tar.gz`
- Windows (x64): `prt_<version>_windows_amd64.zip`

### Build from Source

```bash
git clone https://github.com/ChrisEdwards/prt.git
cd prt
make build
./bin/prt
```

## Quick Start

1. **Install the GitHub CLI** (if not already installed):
   ```bash
   brew install gh
   gh auth login
   ```

2. **Run PRT** - the setup wizard launches automatically:
   ```bash
   prt
   ```

3. **Configure your paths and team** in the wizard

4. **Run PRT again** to see your PR dashboard!

## Usage

```bash
# Show PR dashboard (default)
prt

# Scan a specific path
prt -p ~/code/work

# Filter repos by pattern
prt -f "api-*"

# Show newest PRs first
prt -s newest

# Output as JSON for scripting
prt --json | jq '.needs_my_attention | length'

# Disable colors (for piping)
prt --no-color > prs.txt
```

## Command Line Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--path` | `-p` | Override search paths from config |
| `--filter` | `-f` | Filter repos by name pattern (glob) |
| `--group` | `-g` | Group by: `project` or `author` |
| `--sort` | `-s` | Sort by: `oldest` or `newest` |
| `--depth` | `-d` | Scan depth (default: 3) |
| `--max-age` | | Hide PRs older than N days (0 = no limit) |
| `--json` | | Output as JSON |
| `--no-color` | | Disable colored output |
| `--version` | `-v` | Show version |
| `--help` | `-h` | Show help |

## Configuration

Configuration file: `~/.prt/config.yaml`

```yaml
# Your GitHub username (auto-detected if empty)
github_username: "jdoe"

# Team members - their PRs are highlighted
team_members:
  - "alice"
  - "bob"
  - "charlie"

# Directories to scan for Git repositories
search_paths:
  - "~/code/work"
  - "~/projects/oss"

# Only include repos matching these patterns (empty = all)
include_repos:
  - "myorg-*"
  - "frontend"

# Max directory depth when scanning (default: 3)
scan_depth: 3

# Known bot accounts (pre-populated, add your own)
bots:
  - "dependabot[bot]"
  - "renovate[bot]"
  - "github-actions[bot]"

# Display options
default_group_by: "project"  # project | author
default_sort: "oldest"       # oldest | newest
show_branch_name: true
show_icons: true
show_other_prs: false        # Show "Other PRs" section

# Filtering options
max_pr_age_days: 0           # Hide PRs older than N days (0 = no limit)
```

### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `github_username` | (auto-detect) | Your GitHub username |
| `team_members` | `[]` | GitHub usernames to highlight |
| `search_paths` | `[]` | Directories to scan |
| `include_repos` | `[]` | Glob patterns to filter repos |
| `scan_depth` | `3` | Max directory depth |
| `bots` | (see defaults) | Known bot accounts |
| `default_group_by` | `project` | Group PRs by project or author |
| `default_sort` | `oldest` | Sort by oldest or newest first |
| `show_branch_name` | `true` | Show branch names |
| `show_icons` | `true` | Show emoji icons |
| `show_other_prs` | `false` | Show "Other PRs" section |
| `max_pr_age_days` | `0` | Hide PRs older than N days (0 = no limit) |

### Environment Variables

All configuration options can be set via environment variables with the `PRT_` prefix. Environment variables override config file values but are overridden by CLI flags.

| Variable | Config Equivalent | Example |
|----------|-------------------|---------|
| `PRT_GITHUB_USERNAME` | `github_username` | `export PRT_GITHUB_USERNAME=jdoe` |
| `PRT_SCAN_DEPTH` | `scan_depth` | `export PRT_SCAN_DEPTH=5` |
| `PRT_DEFAULT_GROUP_BY` | `default_group_by` | `export PRT_DEFAULT_GROUP_BY=author` |
| `PRT_DEFAULT_SORT` | `default_sort` | `export PRT_DEFAULT_SORT=newest` |
| `PRT_SHOW_BRANCH_NAME` | `show_branch_name` | `export PRT_SHOW_BRANCH_NAME=false` |
| `PRT_SHOW_ICONS` | `show_icons` | `export PRT_SHOW_ICONS=false` |
| `PRT_SHOW_OTHER_PRS` | `show_other_prs` | `export PRT_SHOW_OTHER_PRS=true` |
| `PRT_MAX_PR_AGE_DAYS` | `max_pr_age_days` | `export PRT_MAX_PR_AGE_DAYS=30` |

**Configuration precedence** (highest to lowest):
1. CLI flags (`--sort newest`)
2. Environment variables (`PRT_DEFAULT_SORT=newest`)
3. Config file (`~/.prt/config.yaml`)
4. Built-in defaults

## Output Categories

### My PRs
PRs you authored. These are your "outgoing" PRs waiting for review.

### Needs My Attention
PRs where:
- You're requested as a reviewer
- You're assigned to the PR
- You haven't approved yet

### Team PRs
PRs from users in your `team_members` list.

### Other PRs
Everything else, including:
- External contributors
- Bots (dependabot, renovate, etc.)

## Stacked PRs

PRT detects "stacked PRs" - chains of dependent PRs. When a PR targets another PR's branch (instead of main), it's visualized as a tree:

```
├── #402 Feature: Auth
│   └── #405 Tests for Auth (blocked)
```

Child PRs are marked as "blocked" until their parent merges.

## JSON Output

Use `--json` for scripting:

```bash
# Count PRs needing your attention
prt --json | jq '.needs_my_attention | length'

# List all PR URLs
prt --json | jq '.my_prs[].url'

# Get scan metadata
prt --json | jq '{repos: .total_repos_scanned, prs: .total_prs_found, user: .username}'

# Export to file
prt --json > ~/pr-snapshot.json
```

### JSON Schema

Top-level structure:

| Field | Type | Description |
|-------|------|-------------|
| `my_prs` | `PR[]` | PRs you authored |
| `needs_my_attention` | `PR[]` | PRs requesting your review or assigned to you |
| `team_prs` | `PR[]` | PRs from your configured team members |
| `other_prs` | `PR[]` | All other PRs |
| `repos_with_prs` | `Repository[]` | Repositories with open PRs |
| `repos_without_prs` | `Repository[]` | Repositories with no open PRs |
| `repos_with_errors` | `Repository[]` | Repositories that failed to scan |
| `stacks` | `object` | Map of repo name to Stack (stacked PRs) |
| `total_repos_scanned` | `int` | Number of repositories scanned |
| `total_prs_found` | `int` | Total PR count |
| `scan_duration_ns` | `int` | Scan time in nanoseconds |
| `username` | `string` | Your GitHub username |

PR object:

| Field | Type | Description |
|-------|------|-------------|
| `number` | `int` | PR number |
| `title` | `string` | PR title |
| `url` | `string` | GitHub URL |
| `author` | `string` | Author's GitHub username |
| `state` | `string` | `OPEN`, `DRAFT`, `MERGED`, or `CLOSED` |
| `is_draft` | `bool` | Whether PR is a draft |
| `base_branch` | `string` | Target branch (e.g., `main`) |
| `head_branch` | `string` | Source branch |
| `created_at` | `string` | ISO 8601 timestamp |
| `ci_status` | `string` | `passing`, `failing`, `pending`, or `none` |
| `review_requests` | `string[]` | Usernames requested to review |
| `assignees` | `string[]` | Assigned usernames |
| `reviews` | `Review[]` | Code reviews |
| `repo_name` | `string` | Repository name |
| `repo_owner` | `string` | Repository owner |

Repository object:

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Repository name |
| `path` | `string` | Local filesystem path |
| `remote_url` | `string` | Git remote URL |
| `owner` | `string` | GitHub owner/org |
| `prs` | `PR[]` | PRs in this repository |
| `scan_status` | `string` | `success`, `no_prs`, `error`, or `skipped` |

## Requirements

- **GitHub CLI (`gh`)** - Must be installed and authenticated
- **macOS, Linux, or Windows** - Pre-built binaries available for all platforms
- **Git repositories** - With GitHub remotes

## Troubleshooting

### "gh: command not found"
Install the GitHub CLI:
```bash
brew install gh  # macOS
# or see https://cli.github.com/
```

### "gh is not authenticated"
Authenticate with GitHub:
```bash
gh auth login
```

### No repositories found
Check that:
1. Your `search_paths` are correct in `~/.prt/config.yaml`
2. The directories contain Git repos with GitHub remotes
3. Your `scan_depth` is deep enough

### PRs not showing
- PRs must be **open** (not merged/closed)
- Repo must have a GitHub remote (not GitLab, Bitbucket, etc.)
- Check `gh pr list` works in the repo directory

## Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Submit a pull request

## Releasing

Releases are automated via GitHub Actions and [goreleaser](https://goreleaser.com/).

### Setup (one-time)

1. Create a Personal Access Token at https://github.com/settings/tokens with `repo` scope
2. Add it as a repository secret named `HOMEBREW_TAP_TOKEN` in Settings > Secrets > Actions

### Creating a Release

```bash
git tag v1.0.0
git push origin v1.0.0
```

This triggers the release workflow which:
- Builds binaries for all platforms (macOS, Linux, Windows)
- Creates a GitHub release with changelog
- Updates the Homebrew formula in [homebrew-tap](https://github.com/ChrisEdwards/homebrew-tap)

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [GitHub CLI](https://cli.github.com/) - GitHub API access
