// Package github provides integration with the gh CLI for PR operations.
package github

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// GHNotFoundError indicates gh CLI is not installed.
type GHNotFoundError struct {
	Message string
}

func (e *GHNotFoundError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "gh CLI not found. Please install: https://cli.github.com/"
}

// GHAuthError indicates gh is not authenticated.
type GHAuthError struct {
	Message string
}

func (e *GHAuthError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "gh CLI not authenticated. Run: gh auth login"
}

// RepoScanError indicates a repository-specific failure.
type RepoScanError struct {
	RepoPath string
	RepoName string
	Cause    error
}

func (e *RepoScanError) Error() string {
	name := e.RepoName
	if name == "" {
		name = e.RepoPath
	}
	return fmt.Sprintf("failed to scan %s: %v", name, e.Cause)
}

func (e *RepoScanError) Unwrap() error {
	return e.Cause
}

// NetworkError indicates a network-related failure.
type NetworkError struct {
	Cause   error
	Retries int
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error after %d retries: %v", e.Retries, e.Cause)
}

func (e *NetworkError) Unwrap() error {
	return e.Cause
}

// RateLimitError indicates GitHub API rate limit was hit.
type RateLimitError struct {
	ResetTime time.Time
}

func (e *RateLimitError) Error() string {
	if !e.ResetTime.IsZero() {
		return fmt.Sprintf("GitHub API rate limit reached. Resets at %s", e.ResetTime.Format(time.RFC822))
	}
	return "GitHub API rate limit reached. Please wait and retry."
}

// RepoNotFoundError indicates the repo doesn't exist or user lacks access.
type RepoNotFoundError struct {
	RepoPath string
}

func (e *RepoNotFoundError) Error() string {
	return fmt.Sprintf("repository not found or no access: %s", e.RepoPath)
}

// ClassifyError examines an error from gh CLI execution and returns
// a more specific error type based on the error message/stderr.
func ClassifyError(err error, repoPath string) error {
	if err == nil {
		return nil
	}

	// Get stderr if available from exec.ExitError
	stderr := ""
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr = string(exitErr.Stderr)
	}
	errStr := err.Error()

	// Check for rate limit
	if containsAny(stderr, errStr, "rate limit", "API rate limit") {
		return &RateLimitError{}
	}

	// Check for not found / resolution errors
	if containsAny(stderr, errStr, "not found", "Could not resolve", "404") {
		return &RepoNotFoundError{RepoPath: repoPath}
	}

	// Check for auth errors
	if containsAny(stderr, errStr, "auth", "401", "403", "not logged in") {
		msg := stderr
		if msg == "" {
			msg = errStr
		}
		return &GHAuthError{Message: msg}
	}

	// Check for network errors
	if containsAny(stderr, errStr, "network", "connection", "timeout", "dial") {
		return &NetworkError{Cause: err, Retries: 0}
	}

	// Default to generic repo scan error
	return &RepoScanError{
		RepoPath: repoPath,
		Cause:    err,
	}
}

// containsAny checks if any of the needles are found in s1 or s2 (case-insensitive).
func containsAny(s1, s2 string, needles ...string) bool {
	s1Lower := strings.ToLower(s1)
	s2Lower := strings.ToLower(s2)
	for _, needle := range needles {
		needleLower := strings.ToLower(needle)
		if strings.Contains(s1Lower, needleLower) || strings.Contains(s2Lower, needleLower) {
			return true
		}
	}
	return false
}
