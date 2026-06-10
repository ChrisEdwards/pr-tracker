package scanner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"prt/internal/config"
)

func mustRun(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	if err := cmd.Run(); err != nil {
		t.Fatalf("running %s %v: %v", cmd.Path, cmd.Args[1:], err)
	}
}

func createGitRepo(t *testing.T, path, remote string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("creating repo dir: %v", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = path
	mustRun(t, cmd)

	if remote != "" {
		cmd = exec.Command("git", "remote", "add", "origin", remote)
		cmd.Dir = path
		mustRun(t, cmd)
	}
}

func TestNewScanner(t *testing.T) {
	t.Run("valid patterns", func(t *testing.T) {
		s, err := NewScanner(3, []string{"myorg-*", "*-api"})
		if err != nil {
			t.Fatalf("NewScanner() error = %v", err)
		}
		if s == nil {
			t.Error("NewScanner() returned nil scanner")
		}
	})

	t.Run("empty patterns", func(t *testing.T) {
		s, err := NewScanner(3, []string{})
		if err != nil {
			t.Fatalf("NewScanner() error = %v", err)
		}
		if s == nil {
			t.Error("NewScanner() returned nil scanner")
		}
	})

	t.Run("invalid pattern", func(t *testing.T) {
		_, err := NewScanner(3, []string{"["})
		if err == nil {
			t.Error("NewScanner() expected error for invalid pattern")
		}
	})
}

func TestCountDepth(t *testing.T) {
	tests := []struct {
		name string
		base string
		path string
		want int
	}{
		{
			name: "same directory",
			base: "/Users/jdoe/code",
			path: "/Users/jdoe/code",
			want: 0,
		},
		{
			name: "one level deep",
			base: "/Users/jdoe/code",
			path: "/Users/jdoe/code/work",
			want: 1,
		},
		{
			name: "two levels deep",
			base: "/Users/jdoe/code",
			path: "/Users/jdoe/code/work/project",
			want: 2,
		},
		{
			name: "three levels deep",
			base: "/Users/jdoe/code",
			path: "/Users/jdoe/code/work/project/sub",
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countDepth(tt.base, tt.path)
			if got != tt.want {
				t.Errorf("countDepth(%q, %q) = %d, want %d", tt.base, tt.path, got, tt.want)
			}
		})
	}
}

func TestScanner_Scan(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	t.Run("finds git repo", func(t *testing.T) {
		// Create temp directory structure
		tmpDir := t.TempDir()

		// Create a git repo with GitHub remote
		repoPath := filepath.Join(tmpDir, "myrepo")
		createGitRepo(t, repoPath, "git@github.com:testorg/myrepo.git")

		// Create scanner and scan
		s, err := NewScanner(3, nil)
		if err != nil {
			t.Fatalf("NewScanner() error = %v", err)
		}

		cfg := &config.Config{
			SearchPaths: []string{tmpDir},
		}

		repos, err := s.Scan(cfg)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(repos) != 1 {
			t.Fatalf("Scan() found %d repos, want 1", len(repos))
		}

		if repos[0].Name != "myrepo" {
			t.Errorf("repo.Name = %q, want %q", repos[0].Name, "myrepo")
		}
		if repos[0].Owner != "testorg" {
			t.Errorf("repo.Owner = %q, want %q", repos[0].Owner, "testorg")
		}
	})

	t.Run("respects depth limit", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create repo at depth 2: tmpDir/level1/myrepo
		deepPath := filepath.Join(tmpDir, "level1", "myrepo")
		createGitRepo(t, deepPath, "git@github.com:org/deeprepo.git")

		// With depth 1, should not find the repo (it's at depth 2)
		s1, _ := NewScanner(1, nil)
		repos1, _ := s1.Scan(&config.Config{SearchPaths: []string{tmpDir}})
		if len(repos1) != 0 {
			t.Errorf("Scan(depth=1) found %d repos, want 0", len(repos1))
		}

		// With depth 2, should find the repo
		s2, _ := NewScanner(2, nil)
		repos2, _ := s2.Scan(&config.Config{SearchPaths: []string{tmpDir}})
		if len(repos2) != 1 {
			t.Errorf("Scan(depth=2) found %d repos, want 1", len(repos2))
		}
	})

	t.Run("applies filter patterns", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create two repos
		for _, name := range []string{"myorg-api", "other-service"} {
			repoPath := filepath.Join(tmpDir, name)
			createGitRepo(t, repoPath, "git@github.com:org/"+name+".git")
		}

		// Filter for myorg-* only
		s, _ := NewScanner(3, []string{"myorg-*"})
		repos, _ := s.Scan(&config.Config{SearchPaths: []string{tmpDir}})

		if len(repos) != 1 {
			t.Fatalf("Scan() found %d repos, want 1", len(repos))
		}
		if repos[0].Name != "myorg-api" {
			t.Errorf("repo.Name = %q, want myorg-api", repos[0].Name)
		}
	})

	t.Run("skips non-GitHub repos", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a GitLab repo
		repoPath := filepath.Join(tmpDir, "gitlab-repo")
		createGitRepo(t, repoPath, "git@gitlab.com:org/repo.git")

		s, _ := NewScanner(3, nil)
		repos, _ := s.Scan(&config.Config{SearchPaths: []string{tmpDir}})

		if len(repos) != 0 {
			t.Errorf("Scan() found %d repos, want 0 (should skip non-GitHub)", len(repos))
		}
	})

	t.Run("handles non-existent search path", func(t *testing.T) {
		s, _ := NewScanner(3, nil)
		repos, err := s.Scan(&config.Config{
			SearchPaths: []string{"/nonexistent/path"},
		})

		if err != nil {
			t.Errorf("Scan() error = %v, want nil (should skip gracefully)", err)
		}
		if len(repos) != 0 {
			t.Errorf("Scan() found %d repos, want 0", len(repos))
		}
	})

	t.Run("no duplicates", func(t *testing.T) {
		tmpDir := t.TempDir()

		repoPath := filepath.Join(tmpDir, "repo")
		createGitRepo(t, repoPath, "git@github.com:org/repo.git")

		// Scan with same path twice
		s, _ := NewScanner(3, nil)
		repos, _ := s.Scan(&config.Config{
			SearchPaths: []string{tmpDir, tmpDir},
		})

		if len(repos) != 1 {
			t.Errorf("Scan() found %d repos, want 1 (no duplicates)", len(repos))
		}
	})

	t.Run("multiple search paths", func(t *testing.T) {
		tmpDir1 := t.TempDir()
		tmpDir2 := t.TempDir()

		// Create repo in each
		for i, dir := range []string{tmpDir1, tmpDir2} {
			repoPath := filepath.Join(dir, "repo")
			createGitRepo(t, repoPath, fmt.Sprintf("git@github.com:org/repo%d.git", i+1))
		}

		s, _ := NewScanner(3, nil)
		repos, _ := s.Scan(&config.Config{
			SearchPaths: []string{tmpDir1, tmpDir2},
		})

		if len(repos) != 2 {
			t.Errorf("Scan() found %d repos, want 2", len(repos))
		}
	})
}

func TestScanWithDefaults(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmpDir := t.TempDir()

	// Create a repo
	repoPath := filepath.Join(tmpDir, "testrepo")
	createGitRepo(t, repoPath, "git@github.com:org/testrepo.git")

	cfg := &config.Config{
		SearchPaths:  []string{tmpDir},
		IncludeRepos: []string{},
		ScanDepth:    3,
	}

	repos, err := ScanWithDefaults(cfg)
	if err != nil {
		t.Fatalf("ScanWithDefaults() error = %v", err)
	}

	if len(repos) != 1 {
		t.Errorf("ScanWithDefaults() found %d repos, want 1", len(repos))
	}
}
