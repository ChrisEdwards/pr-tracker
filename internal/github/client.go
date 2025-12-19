package github

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"prt/internal/models"
)

// prListJSONFields are the fields we request from gh pr list.
const prListJSONFields = "number,title,url,author,state,isDraft,createdAt,baseRefName,headRefName,statusCheckRollup,reviewRequests,assignees,reviews"

// Client provides methods for interacting with GitHub via the gh CLI.
type Client interface {
	// Check verifies gh CLI is installed and authenticated.
	Check() error
	// GetCurrentUser returns the authenticated GitHub username.
	GetCurrentUser() (string, error)
	// CheckAndGetUser verifies gh CLI and returns the current user in parallel.
	// This is faster than calling Check() then GetCurrentUser() sequentially.
	CheckAndGetUser() (string, error)
	// ListPRs fetches open PRs for a repository.
	ListPRs(repoPath string) ([]*models.PR, error)
}

// client is the default implementation of Client.
type client struct {
	// execLookPath allows mocking exec.LookPath for testing
	execLookPath func(file string) (string, error)
	// execCommand allows mocking exec.Command for testing
	execCommand func(name string, arg ...string) *exec.Cmd
	// retryer handles retry logic for transient failures
	retryer *Retryer
}

// NewClient creates a new GitHub client with default retry config.
func NewClient() Client {
	return NewClientWithConfig(DefaultRetryConfig)
}

// NewClientWithConfig creates a new GitHub client with custom retry config.
func NewClientWithConfig(retryConfig RetryConfig) Client {
	return &client{
		execLookPath: exec.LookPath,
		execCommand:  exec.Command,
		retryer:      NewRetryer(retryConfig),
	}
}

// Check verifies that the gh CLI is installed and authenticated.
// Returns GHNotFoundError if gh is not installed.
// Returns GHAuthError if gh is not authenticated.
func (c *client) Check() error {
	// 1. Check gh exists
	_, err := c.execLookPath("gh")
	if err != nil {
		return &GHNotFoundError{
			Message: `GitHub CLI (gh) not found.

Please install it:
  brew install gh        # macOS
  sudo apt install gh    # Debian/Ubuntu
  winget install gh      # Windows

Then authenticate:
  gh auth login`,
		}
	}

	// 2. Check authentication
	cmd := c.execCommand("gh", "auth", "status")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if err := cmd.Run(); err != nil {
		return &GHAuthError{
			Message: `GitHub CLI is not authenticated.

Please run:
  gh auth login`,
		}
	}

	return nil
}

// GetCurrentUser returns the authenticated GitHub username by querying the API.
func (c *client) GetCurrentUser() (string, error) {
	cmd := c.execCommand("gh", "api", "user", "--jq", ".login")

	out, err := cmd.Output()
	if err != nil {
		// Try to get more info from stderr
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("failed to get current user: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	username := strings.TrimSpace(string(out))
	if username == "" {
		return "", fmt.Errorf("empty username returned from GitHub API")
	}

	return username, nil
}

// CheckAndGetUser verifies gh CLI is installed/authenticated and returns
// the current username in parallel. This is faster than sequential Check()
// then GetCurrentUser() calls since both gh commands run concurrently.
func (c *client) CheckAndGetUser() (string, error) {
	// First, check gh exists (must be done first, can't parallelize)
	_, err := c.execLookPath("gh")
	if err != nil {
		return "", &GHNotFoundError{
			Message: `GitHub CLI (gh) not found.

Please install it:
  brew install gh        # macOS
  sudo apt install gh    # Debian/Ubuntu
  winget install gh      # Windows

Then authenticate:
  gh auth login`,
		}
	}

	// Run auth check and user fetch in parallel
	var wg sync.WaitGroup
	var authErr, userErr error
	var username string

	wg.Add(2)

	// Auth check goroutine
	go func() {
		defer wg.Done()
		cmd := c.execCommand("gh", "auth", "status")
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		if err := cmd.Run(); err != nil {
			authErr = &GHAuthError{
				Message: `GitHub CLI is not authenticated.

Please run:
  gh auth login`,
			}
		}
	}()

	// User fetch goroutine
	go func() {
		defer wg.Done()
		cmd := c.execCommand("gh", "api", "user", "--jq", ".login")
		out, err := cmd.Output()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				userErr = fmt.Errorf("failed to get current user: %s", strings.TrimSpace(string(exitErr.Stderr)))
			} else {
				userErr = fmt.Errorf("failed to get current user: %w", err)
			}
			return
		}
		username = strings.TrimSpace(string(out))
		if username == "" {
			userErr = fmt.Errorf("empty username returned from GitHub API")
		}
	}()

	wg.Wait()

	// Auth errors take priority since they indicate fundamental issues
	if authErr != nil {
		return "", authErr
	}
	if userErr != nil {
		return "", userErr
	}

	return username, nil
}

// ListPRs fetches open pull requests for the repository at repoPath.
// Uses retry logic for transient network failures.
// Returns empty slice if no PRs exist.
func (c *client) ListPRs(repoPath string) ([]*models.PR, error) {
	var result []*models.PR

	err := c.retryer.Do(func() error {
		cmd := c.execCommand("gh", "pr", "list",
			"--json", prListJSONFields,
			"--state", "open",
		)
		cmd.Dir = repoPath

		out, err := cmd.Output()
		if err != nil {
			// Classify the error for proper retry handling
			return ClassifyError(err, repoPath)
		}

		// Empty output or empty array means no PRs
		outStr := strings.TrimSpace(string(out))
		if outStr == "" || outStr == "[]" {
			result = []*models.PR{}
			return nil
		}

		prs, err := ParsePRList(out)
		if err != nil {
			// Parse errors are not retriable
			return &RepoScanError{
				RepoPath: repoPath,
				Cause:    err,
			}
		}

		result = prs
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
