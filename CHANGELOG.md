# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2025-12-22

### Added

- `--max-age` flag to filter out old PRs from display
- `max_pr_age_days` config option for persistent filtering of stale PRs

## [0.3.0] - 2025-12-21

### Added

- Progress indicator when scanning repos

### Changed

- Show single updating repo line instead of scrolling list during scan
- Removed "Repos with no open PRs" section for cleaner output

### Fixed

- Tree continuation lines on PR detail rows
- Tree continuation for parent-child connections in stacks
- Preserve TreeStyle for blocked PRs
- Spacing before PR number in stack display

### Performance

- Parallelize startup operations for faster loading

## [0.2.0] - 2025-12-19

### Changed

- Hide "Other PRs" section by default to reduce noise

### Added

- `show_other_prs` config option to restore the section if desired

## [0.1.0] - 2025-12-19

### Added

- Multi-repo scanning to discover Git repos in configured directories
- Smart PR categorization: My PRs, Needs Attention, Team, Other
- Stack detection for visualizing dependent PR chains
- Bot filtering to auto-deprioritize dependabot, renovate, etc.
- Sorting by oldest or newest first
- JSON output for scripting with jq
- Interactive setup wizard on first run

[Unreleased]: https://github.com/ChrisEdwards/pr-tracker/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/ChrisEdwards/pr-tracker/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/ChrisEdwards/pr-tracker/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/ChrisEdwards/pr-tracker/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/ChrisEdwards/pr-tracker/releases/tag/v0.1.0
