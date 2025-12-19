package github

import (
	"errors"
	"os/exec"
	"testing"
	"time"
)

// testRetryer creates a Retryer with no delays for testing.
func testRetryer() *Retryer {
	r := NewDefaultRetryer()
	r.sleep = func(d time.Duration) {} // No-op sleep for tests
	return r
}

func TestNewClient(t *testing.T) {
	c := NewClient()
	if c == nil {
		t.Error("NewClient should return non-nil client")
	}
}

func TestNewClientWithConfig(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts: 5,
		InitialWait: 2 * time.Second,
		MaxWait:     20 * time.Second,
	}
	c := NewClientWithConfig(cfg)
	if c == nil {
		t.Error("NewClientWithConfig should return non-nil client")
	}
}

func TestCheck_GHNotFound(t *testing.T) {
	c := &client{
		execLookPath: func(file string) (string, error) {
			return "", errors.New("executable not found")
		},
		execCommand: exec.Command, // Won't be called
		retryer:     testRetryer(),
	}

	err := c.Check()
	if err == nil {
		t.Fatal("expected error when gh not found")
	}

	ghErr, ok := err.(*GHNotFoundError)
	if !ok {
		t.Fatalf("expected GHNotFoundError, got %T", err)
	}

	if ghErr.Message == "" {
		t.Error("expected non-empty error message")
	}
}

func TestCheck_GHNotAuthenticated(t *testing.T) {
	c := &client{
		execLookPath: func(file string) (string, error) {
			return "/usr/bin/gh", nil // gh is found
		},
		execCommand: func(name string, arg ...string) *exec.Cmd {
			// Return a command that will fail
			return exec.Command("false")
		},
		retryer: testRetryer(),
	}

	err := c.Check()
	if err == nil {
		t.Fatal("expected error when gh not authenticated")
	}

	authErr, ok := err.(*GHAuthError)
	if !ok {
		t.Fatalf("expected GHAuthError, got %T", err)
	}

	if authErr.Message == "" {
		t.Error("expected non-empty error message")
	}
}

func TestCheck_Success(t *testing.T) {
	c := &client{
		execLookPath: func(file string) (string, error) {
			return "/usr/bin/gh", nil
		},
		execCommand: func(name string, arg ...string) *exec.Cmd {
			// Return a command that will succeed
			return exec.Command("true")
		},
		retryer: testRetryer(),
	}

	err := c.Check()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCheck_VerifiesGHFirst(t *testing.T) {
	lookPathCalled := false
	commandCalled := false

	c := &client{
		execLookPath: func(file string) (string, error) {
			lookPathCalled = true
			if file != "gh" {
				t.Errorf("expected to look for 'gh', got %q", file)
			}
			return "", errors.New("not found")
		},
		execCommand: func(name string, arg ...string) *exec.Cmd {
			commandCalled = true
			return exec.Command("true")
		},
		retryer: testRetryer(),
	}

	c.Check()

	if !lookPathCalled {
		t.Error("expected execLookPath to be called")
	}

	if commandCalled {
		t.Error("execCommand should not be called if gh not found")
	}
}

func TestCheck_AuthCommandArgs(t *testing.T) {
	var capturedName string
	var capturedArgs []string

	c := &client{
		execLookPath: func(file string) (string, error) {
			return "/usr/bin/gh", nil
		},
		execCommand: func(name string, arg ...string) *exec.Cmd {
			capturedName = name
			capturedArgs = arg
			return exec.Command("true")
		},
		retryer: testRetryer(),
	}

	c.Check()

	if capturedName != "gh" {
		t.Errorf("expected command 'gh', got %q", capturedName)
	}

	if len(capturedArgs) != 2 || capturedArgs[0] != "auth" || capturedArgs[1] != "status" {
		t.Errorf("expected args [auth status], got %v", capturedArgs)
	}
}

func TestGetCurrentUser_Success(t *testing.T) {
	c := &client{
		execLookPath: exec.LookPath,
		execCommand: func(name string, arg ...string) *exec.Cmd {
			// Return a command that echoes the username
			return exec.Command("echo", "testuser")
		},
		retryer: testRetryer(),
	}

	user, err := c.GetCurrentUser()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user != "testuser" {
		t.Errorf("expected 'testuser', got %q", user)
	}
}

func TestGetCurrentUser_TrimsWhitespace(t *testing.T) {
	c := &client{
		execLookPath: exec.LookPath,
		execCommand: func(name string, arg ...string) *exec.Cmd {
			// Return a command that echoes with extra whitespace
			return exec.Command("echo", "  testuser  ")
		},
		retryer: testRetryer(),
	}

	user, err := c.GetCurrentUser()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user != "testuser" {
		t.Errorf("expected 'testuser', got %q", user)
	}
}

func TestGetCurrentUser_EmptyResponse(t *testing.T) {
	c := &client{
		execLookPath: exec.LookPath,
		execCommand: func(name string, arg ...string) *exec.Cmd {
			// Return a command that echoes empty string
			return exec.Command("echo", "")
		},
		retryer: testRetryer(),
	}

	_, err := c.GetCurrentUser()
	if err == nil {
		t.Fatal("expected error for empty response")
	}

	if err.Error() != "empty username returned from GitHub API" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestGetCurrentUser_CommandFails(t *testing.T) {
	c := &client{
		execLookPath: exec.LookPath,
		execCommand: func(name string, arg ...string) *exec.Cmd {
			// Return a command that fails
			return exec.Command("false")
		},
		retryer: testRetryer(),
	}

	_, err := c.GetCurrentUser()
	if err == nil {
		t.Fatal("expected error when command fails")
	}
}

func TestGetCurrentUser_CommandArgs(t *testing.T) {
	var capturedName string
	var capturedArgs []string

	c := &client{
		execLookPath: exec.LookPath,
		execCommand: func(name string, arg ...string) *exec.Cmd {
			capturedName = name
			capturedArgs = arg
			return exec.Command("echo", "testuser")
		},
		retryer: testRetryer(),
	}

	c.GetCurrentUser()

	if capturedName != "gh" {
		t.Errorf("expected command 'gh', got %q", capturedName)
	}

	expectedArgs := []string{"api", "user", "--jq", ".login"}
	if len(capturedArgs) != len(expectedArgs) {
		t.Fatalf("expected %d args, got %d", len(expectedArgs), len(capturedArgs))
	}

	for i, expected := range expectedArgs {
		if capturedArgs[i] != expected {
			t.Errorf("arg[%d]: expected %q, got %q", i, expected, capturedArgs[i])
		}
	}
}

// ListPRs tests

func TestListPRs_Success(t *testing.T) {
	validJSON := `[{
		"number": 123,
		"title": "Test PR",
		"url": "https://github.com/org/repo/pull/123",
		"author": {"login": "testuser"},
		"state": "OPEN",
		"isDraft": false,
		"createdAt": "2024-12-15T10:30:00Z",
		"baseRefName": "main",
		"headRefName": "feature",
		"statusCheckRollup": [],
		"reviewRequests": [],
		"assignees": [],
		"reviews": []
	}]`

	c := &client{
		execLookPath: exec.LookPath,
		execCommand: func(name string, arg ...string) *exec.Cmd {
			return exec.Command("echo", validJSON)
		},
		retryer: testRetryer(),
	}

	// Use current directory (exists) for testing
	prs, err := c.ListPRs(".")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(prs) != 1 {
		t.Fatalf("expected 1 PR, got %d", len(prs))
	}

	if prs[0].Number != 123 {
		t.Errorf("expected PR number 123, got %d", prs[0].Number)
	}
	if prs[0].Title != "Test PR" {
		t.Errorf("expected title 'Test PR', got %q", prs[0].Title)
	}
	if prs[0].Author != "testuser" {
		t.Errorf("expected author 'testuser', got %q", prs[0].Author)
	}
}

func TestListPRs_EmptyArray(t *testing.T) {
	c := &client{
		execLookPath: exec.LookPath,
		execCommand: func(name string, arg ...string) *exec.Cmd {
			return exec.Command("echo", "[]")
		},
		retryer: testRetryer(),
	}

	prs, err := c.ListPRs(".")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(prs) != 0 {
		t.Errorf("expected 0 PRs, got %d", len(prs))
	}
}

func TestListPRs_EmptyOutput(t *testing.T) {
	c := &client{
		execLookPath: exec.LookPath,
		execCommand: func(name string, arg ...string) *exec.Cmd {
			return exec.Command("echo", "")
		},
		retryer: testRetryer(),
	}

	prs, err := c.ListPRs(".")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(prs) != 0 {
		t.Errorf("expected 0 PRs, got %d", len(prs))
	}
}

func TestListPRs_CommandArgs(t *testing.T) {
	var capturedName string
	var capturedArgs []string

	c := &client{
		execLookPath: exec.LookPath,
		execCommand: func(name string, arg ...string) *exec.Cmd {
			capturedName = name
			capturedArgs = arg
			return exec.Command("echo", "[]")
		},
		retryer: testRetryer(),
	}

	c.ListPRs(".")

	if capturedName != "gh" {
		t.Errorf("expected command 'gh', got %q", capturedName)
	}

	// Check for expected args
	foundPR := false
	foundList := false
	foundJSON := false
	foundState := false

	for i, arg := range capturedArgs {
		if arg == "pr" {
			foundPR = true
		}
		if arg == "list" {
			foundList = true
		}
		if arg == "--json" {
			foundJSON = true
		}
		if arg == "--state" && i+1 < len(capturedArgs) && capturedArgs[i+1] == "open" {
			foundState = true
		}
	}

	if !foundPR {
		t.Error("expected 'pr' in args")
	}
	if !foundList {
		t.Error("expected 'list' in args")
	}
	if !foundJSON {
		t.Error("expected '--json' in args")
	}
	if !foundState {
		t.Error("expected '--state open' in args")
	}
}

func TestListPRs_MultiplePRs(t *testing.T) {
	validJSON := `[
		{
			"number": 1,
			"title": "PR One",
			"url": "https://github.com/org/repo/pull/1",
			"author": {"login": "user1"},
			"state": "OPEN",
			"isDraft": false,
			"createdAt": "2024-12-15T10:30:00Z",
			"baseRefName": "main",
			"headRefName": "feature1",
			"statusCheckRollup": [],
			"reviewRequests": [],
			"assignees": [],
			"reviews": []
		},
		{
			"number": 2,
			"title": "PR Two",
			"url": "https://github.com/org/repo/pull/2",
			"author": {"login": "user2"},
			"state": "OPEN",
			"isDraft": true,
			"createdAt": "2024-12-16T10:30:00Z",
			"baseRefName": "main",
			"headRefName": "feature2",
			"statusCheckRollup": [],
			"reviewRequests": [],
			"assignees": [],
			"reviews": []
		}
	]`

	c := &client{
		execLookPath: exec.LookPath,
		execCommand: func(name string, arg ...string) *exec.Cmd {
			return exec.Command("echo", validJSON)
		},
		retryer: testRetryer(),
	}

	prs, err := c.ListPRs(".")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(prs) != 2 {
		t.Fatalf("expected 2 PRs, got %d", len(prs))
	}

	if prs[0].Number != 1 || prs[1].Number != 2 {
		t.Errorf("unexpected PR numbers: %d, %d", prs[0].Number, prs[1].Number)
	}

	if !prs[1].IsDraft {
		t.Error("expected second PR to be a draft")
	}
}

func TestListPRs_RetriesOnTransientError(t *testing.T) {
	calls := 0
	validJSON := `[]`

	c := &client{
		execLookPath: exec.LookPath,
		execCommand: func(name string, arg ...string) *exec.Cmd {
			calls++
			if calls < 2 {
				// First call fails with network-like error
				return exec.Command("false")
			}
			// Second call succeeds
			return exec.Command("echo", validJSON)
		},
		retryer: testRetryer(),
	}

	prs, err := c.ListPRs(".")
	if err != nil {
		t.Fatalf("expected success after retry, got %v", err)
	}

	if calls < 2 {
		t.Errorf("expected at least 2 calls (retry), got %d", calls)
	}

	if prs == nil {
		t.Error("expected non-nil PRs slice")
	}
}

// CheckAndGetUser tests

func TestCheckAndGetUser_GHNotFound(t *testing.T) {
	c := &client{
		execLookPath: func(file string) (string, error) {
			return "", errors.New("executable not found")
		},
		execCommand: exec.Command, // Won't be called
		retryer:     testRetryer(),
	}

	_, err := c.CheckAndGetUser()
	if err == nil {
		t.Fatal("expected error when gh not found")
	}

	ghErr, ok := err.(*GHNotFoundError)
	if !ok {
		t.Fatalf("expected GHNotFoundError, got %T", err)
	}

	if ghErr.Message == "" {
		t.Error("expected non-empty error message")
	}
}

func TestCheckAndGetUser_Success(t *testing.T) {
	c := &client{
		execLookPath: func(file string) (string, error) {
			return "/usr/bin/gh", nil
		},
		execCommand: func(name string, arg ...string) *exec.Cmd {
			// Both auth status and api user commands go here
			if len(arg) >= 2 && arg[0] == "auth" && arg[1] == "status" {
				return exec.Command("true")
			}
			if len(arg) >= 2 && arg[0] == "api" && arg[1] == "user" {
				return exec.Command("echo", "testuser")
			}
			return exec.Command("true")
		},
		retryer: testRetryer(),
	}

	user, err := c.CheckAndGetUser()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user != "testuser" {
		t.Errorf("expected 'testuser', got %q", user)
	}
}

func TestCheckAndGetUser_AuthFailure(t *testing.T) {
	c := &client{
		execLookPath: func(file string) (string, error) {
			return "/usr/bin/gh", nil
		},
		execCommand: func(name string, arg ...string) *exec.Cmd {
			// Auth fails, user fetch succeeds
			if len(arg) >= 2 && arg[0] == "auth" && arg[1] == "status" {
				return exec.Command("false")
			}
			if len(arg) >= 2 && arg[0] == "api" && arg[1] == "user" {
				return exec.Command("echo", "testuser")
			}
			return exec.Command("true")
		},
		retryer: testRetryer(),
	}

	_, err := c.CheckAndGetUser()
	if err == nil {
		t.Fatal("expected error when auth fails")
	}

	authErr, ok := err.(*GHAuthError)
	if !ok {
		t.Fatalf("expected GHAuthError, got %T", err)
	}

	if authErr.Message == "" {
		t.Error("expected non-empty error message")
	}
}

func TestCheckAndGetUser_UserFetchFailure(t *testing.T) {
	c := &client{
		execLookPath: func(file string) (string, error) {
			return "/usr/bin/gh", nil
		},
		execCommand: func(name string, arg ...string) *exec.Cmd {
			// Auth succeeds, user fetch fails
			if len(arg) >= 2 && arg[0] == "auth" && arg[1] == "status" {
				return exec.Command("true")
			}
			if len(arg) >= 2 && arg[0] == "api" && arg[1] == "user" {
				return exec.Command("false")
			}
			return exec.Command("true")
		},
		retryer: testRetryer(),
	}

	_, err := c.CheckAndGetUser()
	if err == nil {
		t.Fatal("expected error when user fetch fails")
	}

	// Should not be an auth error
	if _, ok := err.(*GHAuthError); ok {
		t.Error("expected non-auth error for user fetch failure")
	}
}

func TestCheckAndGetUser_EmptyUsername(t *testing.T) {
	c := &client{
		execLookPath: func(file string) (string, error) {
			return "/usr/bin/gh", nil
		},
		execCommand: func(name string, arg ...string) *exec.Cmd {
			if len(arg) >= 2 && arg[0] == "auth" && arg[1] == "status" {
				return exec.Command("true")
			}
			if len(arg) >= 2 && arg[0] == "api" && arg[1] == "user" {
				return exec.Command("echo", "")
			}
			return exec.Command("true")
		},
		retryer: testRetryer(),
	}

	_, err := c.CheckAndGetUser()
	if err == nil {
		t.Fatal("expected error for empty username")
	}
}
