package github

import (
	"errors"
	"os/exec"
	"testing"
	"time"
)

func TestGHNotFoundError(t *testing.T) {
	err := &GHNotFoundError{}
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}

	customErr := &GHNotFoundError{Message: "custom message"}
	if customErr.Error() != "custom message" {
		t.Errorf("expected custom message, got %q", customErr.Error())
	}
}

func TestGHAuthError(t *testing.T) {
	err := &GHAuthError{}
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}

	customErr := &GHAuthError{Message: "auth failed"}
	if customErr.Error() != "auth failed" {
		t.Errorf("expected 'auth failed', got %q", customErr.Error())
	}
}

func TestRepoScanError(t *testing.T) {
	cause := errors.New("underlying error")
	err := &RepoScanError{
		RepoPath: "/path/to/repo",
		RepoName: "myrepo",
		Cause:    cause,
	}

	if err.Error() != "failed to scan myrepo: underlying error" {
		t.Errorf("unexpected error message: %q", err.Error())
	}

	if err.Unwrap() != cause {
		t.Error("Unwrap should return the cause")
	}

	// Test with empty RepoName - should use RepoPath
	err2 := &RepoScanError{
		RepoPath: "/path/to/repo",
		Cause:    cause,
	}
	if err2.Error() != "failed to scan /path/to/repo: underlying error" {
		t.Errorf("unexpected error message: %q", err2.Error())
	}
}

func TestNetworkError(t *testing.T) {
	cause := errors.New("connection refused")
	err := &NetworkError{
		Cause:   cause,
		Retries: 3,
	}

	expected := "network error after 3 retries: connection refused"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}

	if err.Unwrap() != cause {
		t.Error("Unwrap should return the cause")
	}
}

func TestRateLimitError(t *testing.T) {
	// Without reset time
	err := &RateLimitError{}
	if err.Error() != "GitHub API rate limit reached. Please wait and retry." {
		t.Errorf("unexpected error message: %q", err.Error())
	}

	// With reset time
	resetTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	errWithTime := &RateLimitError{ResetTime: resetTime}
	if errWithTime.Error() == "" {
		t.Error("expected non-empty error message with reset time")
	}
}

func TestRepoNotFoundError(t *testing.T) {
	err := &RepoNotFoundError{RepoPath: "owner/repo"}
	expected := "repository not found or no access: owner/repo"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestClassifyError_Nil(t *testing.T) {
	result := ClassifyError(nil, "/path")
	if result != nil {
		t.Error("ClassifyError(nil) should return nil")
	}
}

func TestClassifyError_RateLimit(t *testing.T) {
	err := errors.New("API rate limit exceeded")
	result := ClassifyError(err, "/path")

	if _, ok := result.(*RateLimitError); !ok {
		t.Errorf("expected RateLimitError, got %T", result)
	}
}

func TestClassifyError_NotFound(t *testing.T) {
	err := errors.New("repository not found")
	result := ClassifyError(err, "/path/to/repo")

	if repoErr, ok := result.(*RepoNotFoundError); !ok {
		t.Errorf("expected RepoNotFoundError, got %T", result)
	} else if repoErr.RepoPath != "/path/to/repo" {
		t.Errorf("expected RepoPath '/path/to/repo', got %q", repoErr.RepoPath)
	}
}

func TestClassifyError_Auth(t *testing.T) {
	tests := []string{"auth error", "401 unauthorized", "403 forbidden", "not logged in"}
	for _, errMsg := range tests {
		err := errors.New(errMsg)
		result := ClassifyError(err, "/path")

		if _, ok := result.(*GHAuthError); !ok {
			t.Errorf("expected GHAuthError for %q, got %T", errMsg, result)
		}
	}
}

func TestClassifyError_Network(t *testing.T) {
	tests := []string{"network error", "connection refused", "timeout", "dial tcp"}
	for _, errMsg := range tests {
		err := errors.New(errMsg)
		result := ClassifyError(err, "/path")

		if _, ok := result.(*NetworkError); !ok {
			t.Errorf("expected NetworkError for %q, got %T", errMsg, result)
		}
	}
}

func TestClassifyError_Default(t *testing.T) {
	err := errors.New("some unknown error")
	result := ClassifyError(err, "/path/to/repo")

	repoErr, ok := result.(*RepoScanError)
	if !ok {
		t.Errorf("expected RepoScanError for unknown error, got %T", result)
	}
	if repoErr.RepoPath != "/path/to/repo" {
		t.Errorf("expected RepoPath '/path/to/repo', got %q", repoErr.RepoPath)
	}
}

func TestClassifyError_ExitError(t *testing.T) {
	// Create an exec.ExitError with stderr containing rate limit message
	// This is a bit tricky to test properly without actually running a command
	// For now, we just verify the function handles regular errors well
	exitErr := &exec.ExitError{}
	result := ClassifyError(exitErr, "/path")

	// Should fall through to default RepoScanError since ExitError without Stderr
	if _, ok := result.(*RepoScanError); !ok {
		t.Errorf("expected RepoScanError for empty ExitError, got %T", result)
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		needles  []string
		expected bool
	}{
		{"rate limit exceeded", "", []string{"rate limit"}, true},
		{"", "RATE LIMIT", []string{"rate limit"}, true}, // case insensitive
		{"foo", "bar", []string{"baz"}, false},
		{"", "", []string{"test"}, false},
		{"network timeout", "foo", []string{"timeout", "dial"}, true},
	}

	for _, tt := range tests {
		result := containsAny(tt.s1, tt.s2, tt.needles...)
		if result != tt.expected {
			t.Errorf("containsAny(%q, %q, %v) = %v, want %v",
				tt.s1, tt.s2, tt.needles, result, tt.expected)
		}
	}
}
