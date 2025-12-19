// Package cli provides the command-line interface for prt.
package cli

import (
	"fmt"
	"os"
	"sync"
	"time"

	"prt/internal/categorizer"
	"prt/internal/config"
	"prt/internal/display"
	"prt/internal/github"
	"prt/internal/models"
	"prt/internal/scanner"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "prt",
		Short: "GitHub PR Tracker",
		Long: `PRT - GitHub PR Tracker

Aggregate and visualize GitHub Pull Request status across multiple
local repositories. Highlights PRs requiring your attention and
shows stacked PR relationships.`,
		RunE:          runPRT,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Flags
	flagPath    string
	flagFilter  string
	flagGroup   string
	flagDepth   int
	flagJSON    bool
	flagNoColor bool
)

func init() {
	rootCmd.Flags().StringVarP(&flagPath, "path", "p", "", "Search path (overrides config)")
	rootCmd.Flags().StringVarP(&flagFilter, "filter", "f", "", "Filter repos by name pattern (glob)")
	rootCmd.Flags().StringVarP(&flagGroup, "group", "g", "", "Group by: project, author")
	rootCmd.Flags().IntVarP(&flagDepth, "depth", "d", 0, "Scan depth (0 uses config default)")
	rootCmd.Flags().BoolVar(&flagJSON, "json", false, "Output as JSON")
	rootCmd.Flags().BoolVar(&flagNoColor, "no-color", false, "Disable colored output")
}

// Execute runs the CLI with the given version string.
func Execute(version string) error {
	rootCmd.Version = version
	return rootCmd.Execute()
}

func runPRT(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	// 1. Load config with flag overrides
	flags := &config.Flags{
		Path:   flagPath,
		Filter: flagFilter,
		Group:  flagGroup,
		Depth:  flagDepth,
	}

	cfg, err := config.Load(flags)
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	// 2. Check if setup needed
	if config.NeedsSetup(cfg) {
		return runWizard(cfg)
	}

	// 3. Validate config
	if err := cfg.Validate(); err != nil {
		return err
	}

	// 4. Create scanner early (needed for parallel scan)
	scnr, err := scanner.NewScanner(cfg.ScanDepth, cfg.IncludeRepos)
	if err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	// 5. Run gh CLI check and repo scanning in parallel
	// This saves time by scanning repos while waiting for gh API calls
	ghClient := github.NewClient()
	needsUsername := cfg.GitHubUsername == ""

	var wg sync.WaitGroup
	var ghErr error
	var scanErr error
	var repos []*models.Repository
	var username string

	wg.Add(2)

	// Goroutine A: gh CLI check + optional username fetch
	go func() {
		defer wg.Done()
		if needsUsername {
			// Combined check + user fetch (parallel internally)
			user, err := ghClient.CheckAndGetUser()
			if err != nil {
				ghErr = err
				return
			}
			username = user
		} else {
			// Just check gh CLI
			if err := ghClient.Check(); err != nil {
				ghErr = err
			}
		}
	}()

	// Goroutine B: Scan for repositories
	go func() {
		defer wg.Done()
		r, err := scnr.Scan(cfg)
		if err != nil {
			scanErr = fmt.Errorf("scan error: %w", err)
			return
		}
		repos = r
	}()

	wg.Wait()

	// Check for errors (gh errors take priority)
	if ghErr != nil {
		return ghErr
	}
	if scanErr != nil {
		return scanErr
	}

	// Apply username if it was fetched
	if needsUsername {
		cfg.GitHubUsername = username
	}

	if len(repos) == 0 {
		fmt.Println("No Git repositories found in configured paths.")
		return nil
	}

	// 6. Fetch PRs with progress
	// Progress callback is nil for now - will be implemented with progress display
	github.FetchAllPRs(repos, ghClient, nil)

	// 7. Categorize
	cat := categorizer.NewCategorizer()
	result := cat.Categorize(repos, cfg, cfg.GitHubUsername)
	result.ScanDuration = time.Since(startTime)

	// 8. Render output
	output, err := display.Render(result, display.RenderOptions{
		ShowIcons:    cfg.ShowIcons,
		ShowBranches: cfg.ShowBranchName,
		ShowOtherPRs: cfg.ShowOtherPRs,
		NoColor:      flagNoColor || os.Getenv("NO_COLOR") != "",
		JSON:         flagJSON,
	})
	if err != nil {
		return fmt.Errorf("render error: %w", err)
	}

	fmt.Print(output)
	return nil
}
