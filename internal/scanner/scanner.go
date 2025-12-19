package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"prt/internal/config"
	"prt/internal/models"
)

// inspectConcurrency is the number of concurrent git remote inspections.
const inspectConcurrency = 10

// Scanner discovers Git repositories with GitHub remotes.
type Scanner interface {
	// Scan searches for repositories in the configured paths.
	Scan(cfg *config.Config) ([]*models.Repository, error)
}

// scanner is the default implementation of Scanner.
type scanner struct {
	maxDepth int
	filter   *RepoFilter
}

// NewScanner creates a new Scanner with the given depth limit and include patterns.
// The maxDepth controls how deep to search into subdirectories.
// The includePatterns are glob patterns to filter repository names (empty = include all).
func NewScanner(maxDepth int, includePatterns []string) (Scanner, error) {
	filter, err := NewRepoFilter(includePatterns)
	if err != nil {
		return nil, err
	}
	return &scanner{
		maxDepth: maxDepth,
		filter:   filter,
	}, nil
}

// Scan walks the configured search paths and returns all discovered repositories.
// It respects the maxDepth limit and filters results by the include patterns.
// Repository inspection (git remote calls) is parallelized for better performance.
func (s *scanner) Scan(cfg *config.Config) ([]*models.Repository, error) {
	// Phase 1: Collect all .git directory paths (fast filesystem walk)
	var repoPaths []string
	seen := make(map[string]bool) // Prevent duplicates

	for _, searchPath := range cfg.SearchPaths {
		// Normalize the search path
		searchPath = filepath.Clean(searchPath)

		// Check if search path exists
		if _, err := os.Stat(searchPath); os.IsNotExist(err) {
			continue // Skip non-existent paths
		}

		err := filepath.WalkDir(searchPath, func(path string, d fs.DirEntry, err error) error {
			// Handle access errors gracefully - skip inaccessible entries
			if err != nil {
				return nil
			}

			// Skip non-directories
			if !d.IsDir() {
				return nil
			}

			// Skip symlinks to avoid infinite loops
			if d.Type()&fs.ModeSymlink != 0 {
				return filepath.SkipDir
			}

			// Check if this is a .git directory - check this BEFORE depth
			// because we want to find .git even if it's one level beyond maxDepth
			// (the repo itself would be at maxDepth, but .git is inside it)
			if d.Name() == ".git" {
				repoPath := filepath.Dir(path)

				// Skip if we've already seen this repo
				if seen[repoPath] {
					return filepath.SkipDir
				}
				seen[repoPath] = true

				// Collect path for parallel inspection
				repoPaths = append(repoPaths, repoPath)

				// Don't descend into .git directory
				return filepath.SkipDir
			}

			// Check depth relative to search path for non-.git directories
			depth := countDepth(searchPath, path)
			if depth > s.maxDepth {
				return filepath.SkipDir
			}

			// Skip hidden directories (except we need to find .git)
			if strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
				return filepath.SkipDir
			}

			// Skip common non-repository directories for performance
			switch d.Name() {
			case "node_modules", "vendor", ".cache", "__pycache__", "venv", ".venv":
				return filepath.SkipDir
			}

			return nil
		})

		if err != nil {
			// Continue to next search path on error
			continue
		}
	}

	// Phase 2: Inspect repositories in parallel (slow git remote calls)
	if len(repoPaths) == 0 {
		return nil, nil
	}

	return s.inspectReposParallel(repoPaths), nil
}

// inspectReposParallel inspects multiple repositories concurrently.
// It filters results by the configured patterns and returns valid GitHub repos.
func (s *scanner) inspectReposParallel(paths []string) []*models.Repository {
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		repos []*models.Repository
		sem   = make(chan struct{}, inspectConcurrency)
	)

	for _, path := range paths {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			repo, err := InspectRepo(p)
			if err != nil {
				// Not a valid GitHub repo - skip silently
				return
			}

			// Apply filter
			if s.filter.Matches(repo.Name) {
				mu.Lock()
				repos = append(repos, repo)
				mu.Unlock()
			}
		}(path)
	}

	wg.Wait()
	return repos
}

// countDepth returns the depth of path relative to base.
// For example, with base=/Users/jdoe/code:
//   - /Users/jdoe/code -> 0
//   - /Users/jdoe/code/work -> 1
//   - /Users/jdoe/code/work/project -> 2
func countDepth(base, path string) int {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return 0
	}
	if rel == "." {
		return 0
	}
	return strings.Count(rel, string(os.PathSeparator)) + 1
}

// ScanWithDefaults creates a scanner with config values and performs the scan.
// This is a convenience function for common use cases.
func ScanWithDefaults(cfg *config.Config) ([]*models.Repository, error) {
	scanner, err := NewScanner(cfg.ScanDepth, cfg.IncludeRepos)
	if err != nil {
		return nil, err
	}
	return scanner.Scan(cfg)
}
