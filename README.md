# PRT - GitHub PR Tracker

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)

> Aggregate GitHub PR status across multiple local repositories

PRT solves "PR Fatigue" - the cognitive burden of tracking PRs across many repos. With a single command, see:

- **My PRs** - PRs you authored, waiting for review
- **Needs My Attention** - PRs requesting your review or assigned to you
- **Team PRs** - PRs from your configured team members
- **Stacked PRs** - Visual tree of dependent PR chains

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

### Go Install

```bash
go install github.com/ChrisEdwards/prt@latest
```

### Binary Download

Download pre-built binaries from [Releases](https://github.com/ChrisEdwards/prt/releases).

Available platforms:
- macOS (Intel): `prt-darwin-amd64.tar.gz`
- macOS (Apple Silicon): `prt-darwin-arm64.tar.gz`
- Linux (x64): `prt-linux-amd64.tar.gz`

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

# Output as JSON for scripting
prt --json | jq '.needsMyAttention | length'

# Disable colors (for piping)
prt --no-color > prs.txt
```

## Command Line Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--path` | `-p` | Override search paths from config |
| `--filter` | `-f` | Filter repos by name pattern (glob) |
| `--group` | `-g` | Group by: `project` or `author` |
| `--depth` | `-d` | Scan depth (default: 3) |
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
prt --json | jq '.needsMyAttention | length'

# List all PR URLs
prt --json | jq '.myPRs[].url'

# Export to file
prt --json > ~/pr-snapshot.json
```

## Requirements

- **GitHub CLI (`gh`)** - Must be installed and authenticated
- **macOS or Linux** - Windows not currently supported
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

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [GitHub CLI](https://cli.github.com/) - GitHub API access
