package github

import (
	"errors"
	"os/exec"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient()
	if c == nil {
		t.Error("NewClient should return non-nil client")
	}
}

func TestCheck_GHNotFound(t *testing.T) {
	c := &client{
		execLookPath: func(file string) (string, error) {
			return "", errors.New("executable not found")
		},
		execCommand: exec.Command, // Won't be called
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
	}

	c.Check()

	if capturedName != "gh" {
		t.Errorf("expected command 'gh', got %q", capturedName)
	}

	if len(capturedArgs) != 2 || capturedArgs[0] != "auth" || capturedArgs[1] != "status" {
		t.Errorf("expected args [auth status], got %v", capturedArgs)
	}
}
