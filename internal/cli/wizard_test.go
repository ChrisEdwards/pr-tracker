package cli

import (
	"bufio"
	"strings"
	"testing"
)

func TestExpandPath_TildeExpansion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantTild bool // true if result should not start with ~
	}{
		{"tilde only", "~", true},
		{"tilde slash", "~/code", true},
		{"tilde projects", "~/projects/myrepo", true},
		{"absolute path", "/home/user/code", false},
		{"relative path", "./code", false},
		{"no tilde", "code", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			if tt.wantTild && strings.HasPrefix(result, "~") {
				t.Errorf("expandPath(%q) = %q, should not start with ~", tt.input, result)
			}
			if !tt.wantTild && strings.HasPrefix(tt.input, "/") && result != tt.input {
				t.Errorf("expandPath(%q) = %q, absolute paths should not change", tt.input, result)
			}
		})
	}
}

func TestExpandPath_HomeDirectory(t *testing.T) {
	// Test that ~ expansion produces a real path
	result := expandPath("~")
	if result == "~" {
		t.Error("expandPath(~) should expand to home directory")
	}
	if !strings.HasPrefix(result, "/") {
		t.Errorf("expandPath(~) = %q, should be absolute path", result)
	}
}

func TestExpandPath_PreservesSubpath(t *testing.T) {
	result := expandPath("~/code/projects")
	if !strings.HasSuffix(result, "/code/projects") {
		t.Errorf("expandPath(~/code/projects) = %q, should preserve subpath", result)
	}
}

func TestPromptSearchPaths_ParsesCommas(t *testing.T) {
	// Create mock input
	input := "~/code, ~/projects, /tmp/test\n"
	reader := bufio.NewReader(strings.NewReader(input))

	paths, err := promptSearchPaths(reader)
	if err != nil {
		t.Fatalf("promptSearchPaths() error = %v", err)
	}

	if len(paths) != 3 {
		t.Errorf("expected 3 paths, got %d: %v", len(paths), paths)
	}

	expected := []string{"~/code", "~/projects", "/tmp/test"}
	for i, want := range expected {
		if i >= len(paths) {
			t.Errorf("missing path at index %d", i)
			continue
		}
		if paths[i] != want {
			t.Errorf("paths[%d] = %q, want %q", i, paths[i], want)
		}
	}
}

func TestPromptSearchPaths_DefaultsOnEmpty(t *testing.T) {
	// Create mock input with empty line
	input := "\n"
	reader := bufio.NewReader(strings.NewReader(input))

	paths, err := promptSearchPaths(reader)
	if err != nil {
		t.Fatalf("promptSearchPaths() error = %v", err)
	}

	// Should return defaults
	if len(paths) != 2 {
		t.Errorf("expected 2 default paths, got %d: %v", len(paths), paths)
	}
	if paths[0] != "~/code" {
		t.Errorf("paths[0] = %q, want ~/code", paths[0])
	}
	if paths[1] != "~/projects" {
		t.Errorf("paths[1] = %q, want ~/projects", paths[1])
	}
}

func TestPromptSearchPaths_TrimsPaths(t *testing.T) {
	input := "  ~/code  ,  ~/projects  \n"
	reader := bufio.NewReader(strings.NewReader(input))

	paths, err := promptSearchPaths(reader)
	if err != nil {
		t.Fatalf("promptSearchPaths() error = %v", err)
	}

	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d: %v", len(paths), paths)
	}
	if paths[0] != "~/code" {
		t.Errorf("paths[0] = %q, want ~/code (trimmed)", paths[0])
	}
	if paths[1] != "~/projects" {
		t.Errorf("paths[1] = %q, want ~/projects (trimmed)", paths[1])
	}
}

func TestPromptSearchPaths_SkipsEmptyParts(t *testing.T) {
	input := "~/code,,,~/projects,,\n"
	reader := bufio.NewReader(strings.NewReader(input))

	paths, err := promptSearchPaths(reader)
	if err != nil {
		t.Fatalf("promptSearchPaths() error = %v", err)
	}

	if len(paths) != 2 {
		t.Errorf("expected 2 paths (empty parts skipped), got %d: %v", len(paths), paths)
	}
}

func TestPromptTeamMembers_ParsesCommas(t *testing.T) {
	input := "alice, bob, charlie\n"
	reader := bufio.NewReader(strings.NewReader(input))

	members, err := promptTeamMembers(reader)
	if err != nil {
		t.Fatalf("promptTeamMembers() error = %v", err)
	}

	expected := []string{"alice", "bob", "charlie"}
	if len(members) != len(expected) {
		t.Fatalf("expected %d members, got %d: %v", len(expected), len(members), members)
	}
	for i, want := range expected {
		if members[i] != want {
			t.Errorf("members[%d] = %q, want %q", i, members[i], want)
		}
	}
}

func TestPromptTeamMembers_SkipsOnEmpty(t *testing.T) {
	input := "\n"
	reader := bufio.NewReader(strings.NewReader(input))

	members, err := promptTeamMembers(reader)
	if err != nil {
		t.Fatalf("promptTeamMembers() error = %v", err)
	}

	if members != nil {
		t.Errorf("expected nil members on empty input, got %v", members)
	}
}

func TestPromptTeamMembers_StripsAtPrefix(t *testing.T) {
	input := "@alice, @bob, charlie\n"
	reader := bufio.NewReader(strings.NewReader(input))

	members, err := promptTeamMembers(reader)
	if err != nil {
		t.Fatalf("promptTeamMembers() error = %v", err)
	}

	expected := []string{"alice", "bob", "charlie"}
	if len(members) != len(expected) {
		t.Fatalf("expected %d members, got %d: %v", len(expected), len(members), members)
	}
	for i, want := range expected {
		if members[i] != want {
			t.Errorf("members[%d] = %q, want %q (@ stripped)", i, members[i], want)
		}
	}
}

func TestPromptTeamMembers_TrimsWhitespace(t *testing.T) {
	input := "  alice  ,  bob  \n"
	reader := bufio.NewReader(strings.NewReader(input))

	members, err := promptTeamMembers(reader)
	if err != nil {
		t.Fatalf("promptTeamMembers() error = %v", err)
	}

	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d: %v", len(members), members)
	}
	if members[0] != "alice" {
		t.Errorf("members[0] = %q, want alice (trimmed)", members[0])
	}
	if members[1] != "bob" {
		t.Errorf("members[1] = %q, want bob (trimmed)", members[1])
	}
}

func TestPromptUsername_DirectInput(t *testing.T) {
	// User provides username directly
	input := "testuser\n"
	reader := bufio.NewReader(strings.NewReader(input))

	username, err := promptUsername(reader)
	if err != nil {
		t.Fatalf("promptUsername() error = %v", err)
	}

	if username != "testuser" {
		t.Errorf("username = %q, want %q", username, "testuser")
	}
}
