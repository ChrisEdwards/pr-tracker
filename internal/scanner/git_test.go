package scanner

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseGitHubRemote(t *testing.T) {
	tests := []struct {
		name      string
		remoteURL string
		wantOwner string
		wantRepo  string
	}{
		// SSH format
		{
			name:      "SSH with .git suffix",
			remoteURL: "git@github.com:owner/repo.git",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "SSH without .git suffix",
			remoteURL: "git@github.com:owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "SSH with hyphenated names",
			remoteURL: "git@github.com:my-org/my-repo.git",
			wantOwner: "my-org",
			wantRepo:  "my-repo",
		},
		{
			name:      "SSH with underscores",
			remoteURL: "git@github.com:some_org/some_repo.git",
			wantOwner: "some_org",
			wantRepo:  "some_repo",
		},

		// HTTPS format
		{
			name:      "HTTPS with .git suffix",
			remoteURL: "https://github.com/owner/repo.git",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "HTTPS without .git suffix",
			remoteURL: "https://github.com/owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "HTTP (not HTTPS)",
			remoteURL: "http://github.com/owner/repo.git",
			wantOwner: "owner",
			wantRepo:  "repo",
		},

		// SSH URL format
		{
			name:      "SSH URL with .git suffix",
			remoteURL: "ssh://git@github.com/owner/repo.git",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "SSH URL without .git suffix",
			remoteURL: "ssh://git@github.com/owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
		},

		// Whitespace handling
		{
			name:      "URL with leading/trailing whitespace",
			remoteURL: "  git@github.com:owner/repo.git  ",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "URL with newline",
			remoteURL: "git@github.com:owner/repo.git\n",
			wantOwner: "owner",
			wantRepo:  "repo",
		},

		// Non-GitHub remotes (should return empty)
		{
			name:      "GitLab SSH",
			remoteURL: "git@gitlab.com:owner/repo.git",
			wantOwner: "",
			wantRepo:  "",
		},
		{
			name:      "GitLab HTTPS",
			remoteURL: "https://gitlab.com/owner/repo.git",
			wantOwner: "",
			wantRepo:  "",
		},
		{
			name:      "Bitbucket SSH",
			remoteURL: "git@bitbucket.org:owner/repo.git",
			wantOwner: "",
			wantRepo:  "",
		},
		{
			name:      "Empty string",
			remoteURL: "",
			wantOwner: "",
			wantRepo:  "",
		},
		{
			name:      "Random string",
			remoteURL: "not-a-url",
			wantOwner: "",
			wantRepo:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOwner, gotRepo := ParseGitHubRemote(tt.remoteURL)
			if gotOwner != tt.wantOwner {
				t.Errorf("ParseGitHubRemote() owner = %q, want %q", gotOwner, tt.wantOwner)
			}
			if gotRepo != tt.wantRepo {
				t.Errorf("ParseGitHubRemote() repo = %q, want %q", gotRepo, tt.wantRepo)
			}
		})
	}
}

func TestGetRemoteURL(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	t.Run("valid git repo", func(t *testing.T) {
		// Create a temporary directory
		tmpDir, err := os.MkdirTemp("", "git-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Initialize a git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to init git repo: %v", err)
		}

		// Set up git config for the repo
		cmd = exec.Command("git", "config", "user.email", "test@test.com")
		cmd.Dir = tmpDir
		_ = cmd.Run()
		cmd = exec.Command("git", "config", "user.name", "Test")
		cmd.Dir = tmpDir
		_ = cmd.Run()

		// Add a remote
		testURL := "git@github.com:testowner/testrepo.git"
		cmd = exec.Command("git", "remote", "add", "origin", testURL)
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Test GetRemoteURL
		got, err := GetRemoteURL(tmpDir)
		if err != nil {
			t.Errorf("GetRemoteURL() error = %v, want nil", err)
		}
		if got != testURL {
			t.Errorf("GetRemoteURL() = %q, want %q", got, testURL)
		}
	})

	t.Run("non-git directory", func(t *testing.T) {
		// Create a temporary directory (not a git repo)
		tmpDir, err := os.MkdirTemp("", "non-git-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Test GetRemoteURL - should fail
		_, err = GetRemoteURL(tmpDir)
		if err == nil {
			t.Error("GetRemoteURL() expected error for non-git directory, got nil")
		}
	})

	t.Run("git repo without origin remote", func(t *testing.T) {
		// Create a temporary directory
		tmpDir, err := os.MkdirTemp("", "git-no-origin-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Initialize a git repo (but don't add origin)
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to init git repo: %v", err)
		}

		// Test GetRemoteURL - should fail
		_, err = GetRemoteURL(tmpDir)
		if err == nil {
			t.Error("GetRemoteURL() expected error for repo without origin, got nil")
		}
	})
}

func TestInspectRepo(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	t.Run("valid GitHub repo", func(t *testing.T) {
		// Create a temporary directory
		tmpDir, err := os.MkdirTemp("", "inspect-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Initialize a git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to init git repo: %v", err)
		}

		// Add a GitHub remote
		testURL := "git@github.com:myorg/myrepo.git"
		cmd = exec.Command("git", "remote", "add", "origin", testURL)
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Test InspectRepo
		repo, err := InspectRepo(tmpDir)
		if err != nil {
			t.Fatalf("InspectRepo() error = %v, want nil", err)
		}

		if repo.Name != "myrepo" {
			t.Errorf("InspectRepo() Name = %q, want %q", repo.Name, "myrepo")
		}
		if repo.Owner != "myorg" {
			t.Errorf("InspectRepo() Owner = %q, want %q", repo.Owner, "myorg")
		}
		if repo.RemoteURL != testURL {
			t.Errorf("InspectRepo() RemoteURL = %q, want %q", repo.RemoteURL, testURL)
		}
		if repo.Path != tmpDir {
			t.Errorf("InspectRepo() Path = %q, want %q", repo.Path, tmpDir)
		}
	})

	t.Run("non-GitHub repo", func(t *testing.T) {
		// Create a temporary directory
		tmpDir, err := os.MkdirTemp("", "inspect-gitlab-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Initialize a git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to init git repo: %v", err)
		}

		// Add a GitLab remote (not GitHub)
		cmd = exec.Command("git", "remote", "add", "origin", "git@gitlab.com:owner/repo.git")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Test InspectRepo - should fail for non-GitHub
		_, err = InspectRepo(tmpDir)
		if err == nil {
			t.Error("InspectRepo() expected error for non-GitHub repo, got nil")
		}
	})

	t.Run("non-git directory", func(t *testing.T) {
		// Create a temporary directory (not a git repo)
		tmpDir, err := os.MkdirTemp("", "inspect-non-git-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Test InspectRepo - should fail
		_, err = InspectRepo(tmpDir)
		if err == nil {
			t.Error("InspectRepo() expected error for non-git directory, got nil")
		}
	})
}

func TestInspectRepo_Integration(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	// Test against the actual pr-tracker repo if we can find it
	// This is an integration test that validates against real data
	cwd, err := os.Getwd()
	if err != nil {
		t.Skip("Could not get working directory")
	}

	// Walk up to find the repo root
	repoRoot := cwd
	for {
		if _, err := os.Stat(filepath.Join(repoRoot, ".git")); err == nil {
			break
		}
		parent := filepath.Dir(repoRoot)
		if parent == repoRoot {
			t.Skip("Could not find repo root")
		}
		repoRoot = parent
	}

	repo, err := InspectRepo(repoRoot)
	if err != nil {
		// This might fail if the repo uses a different remote, which is fine
		t.Skipf("Could not inspect repo: %v", err)
	}

	// Basic sanity checks
	if repo.Name == "" {
		t.Error("InspectRepo() returned empty Name")
	}
	if repo.Owner == "" {
		t.Error("InspectRepo() returned empty Owner")
	}
	if repo.RemoteURL == "" {
		t.Error("InspectRepo() returned empty RemoteURL")
	}
	if repo.Path != repoRoot {
		t.Errorf("InspectRepo() Path = %q, want %q", repo.Path, repoRoot)
	}
}
