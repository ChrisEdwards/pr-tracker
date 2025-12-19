package github

import (
	"io"
	"os/exec"
)

// Client provides methods for interacting with GitHub via the gh CLI.
type Client interface {
	// Check verifies gh CLI is installed and authenticated.
	Check() error
}

// client is the default implementation of Client.
type client struct {
	// execLookPath allows mocking exec.LookPath for testing
	execLookPath func(file string) (string, error)
	// execCommand allows mocking exec.Command for testing
	execCommand func(name string, arg ...string) *exec.Cmd
}

// NewClient creates a new GitHub client.
func NewClient() Client {
	return &client{
		execLookPath: exec.LookPath,
		execCommand:  exec.Command,
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
