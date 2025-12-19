// Package scanner provides functionality for discovering and inspecting
// Git repositories with GitHub remotes.
package scanner

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"prt/internal/models"
)

var (
	// SSH format: git@github.com:owner/repo.git
	sshRegex = regexp.MustCompile(`^git@github\.com:([^/]+)/([^/]+?)(\.git)?$`)

	// HTTPS format: https://github.com/owner/repo.git
	httpsRegex = regexp.MustCompile(`^https?://github\.com/([^/]+)/([^/]+?)(\.git)?$`)

	// SSH URL format: ssh://git@github.com/owner/repo.git
	sshURLRegex = regexp.MustCompile(`^ssh://git@github\.com/([^/]+)/([^/]+?)(\.git)?$`)
)

// ParseGitHubRemote extracts the owner and repository name from a GitHub remote URL.
// Returns empty strings if the URL is not a recognized GitHub format.
//
// Supported formats:
//   - SSH: git@github.com:owner/repo.git
//   - HTTPS: https://github.com/owner/repo.git
//   - SSH URL: ssh://git@github.com/owner/repo.git
//
// The .git suffix is optional in all formats.
func ParseGitHubRemote(remoteURL string) (owner, repo string) {
	remoteURL = strings.TrimSpace(remoteURL)

	// Try SSH format
	if matches := sshRegex.FindStringSubmatch(remoteURL); len(matches) >= 3 {
		return matches[1], matches[2]
	}

	// Try HTTPS format
	if matches := httpsRegex.FindStringSubmatch(remoteURL); len(matches) >= 3 {
		return matches[1], matches[2]
	}

	// Try SSH URL format
	if matches := sshURLRegex.FindStringSubmatch(remoteURL); len(matches) >= 3 {
		return matches[1], matches[2]
	}

	// Not a GitHub remote
	return "", ""
}

// GetRemoteURL returns the URL of the "origin" remote for a Git repository.
// Returns an error if the repository has no origin remote.
func GetRemoteURL(repoPath string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoPath

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("no origin remote: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}

// InspectRepo examines a directory and returns Repository information if it's
// a Git repository with a GitHub remote. Returns an error if the directory
// is not a Git repo or doesn't have a GitHub remote.
func InspectRepo(path string) (*models.Repository, error) {
	remoteURL, err := GetRemoteURL(path)
	if err != nil {
		return nil, err
	}

	owner, name := ParseGitHubRemote(remoteURL)
	if owner == "" || name == "" {
		return nil, fmt.Errorf("not a GitHub repository: %s", remoteURL)
	}

	return &models.Repository{
		Name:      name,
		Path:      path,
		RemoteURL: remoteURL,
		Owner:     owner,
	}, nil
}
