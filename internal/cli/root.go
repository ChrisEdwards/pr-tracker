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

// FastScanThreshold is the minimum duration before showing progress.
// If scanning completes faster than this, no progress is shown.
const FastScanThreshold = 500 * time.Millisecond

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
	flagMaxAge  int
	flagJSON    bool
	flagNoColor bool
)

func init() {
	rootCmd.Flags().StringVarP(&flagPath, "path", "p", "", "Search path (overrides config)")
	rootCmd.Flags().StringVarP(&flagFilter, "filter", "f", "", "Filter repos by name pattern (glob)")
	rootCmd.Flags().StringVarP(&flagGroup, "group", "g", "", "Group by: project, author")
	rootCmd.Flags().IntVarP(&flagDepth, "depth", "d", 0, "Scan depth (0 uses config default)")
	rootCmd.Flags().IntVar(&flagMaxAge, "max-age", 0, "Hide PRs older than N days (0 uses config default)")
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

	// Determine output settings
	isTTY := display.IsTTY(os.Stdout)
	noColor := flagNoColor || os.Getenv("NO_COLOR") != ""
	useASCII := noColor // Use ASCII if colors are disabled

	// Apply color settings
	if noColor {
		display.DisableColors()
	}

	// 1. Load config with flag overrides
	flags := &config.Flags{
		Path:   flagPath,
		Filter: flagFilter,
		Group:  flagGroup,
		Depth:  flagDepth,
		MaxAge: flagMaxAge,
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

	// 5. Show discovery spinner while scanning
	// Only show spinner for TTY and non-JSON output
	var spinner *display.Spinner
	showProgress := isTTY && !flagJSON
	if showProgress {
		spinner = display.NewSpinner(os.Stdout)
		spinner.SetASCII(useASCII)
		spinner.Start("Discovering repositories...")
	}

	// 6. Run gh CLI check and repo scanning in parallel
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
		// Update spinner count as repos are found
		if spinner != nil {
			spinner.UpdateCount(len(r))
		}
	}()

	wg.Wait()

	// Stop spinner
	if spinner != nil {
		spinner.Stop()
	}

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

	// 7. Fetch PRs with progress display
	var progress *display.ProgressDisplay
	if showProgress && len(repos) > 0 {
		progress = display.NewProgressDisplay(len(repos),
			display.WithWriter(os.Stdout),
			display.WithTTY(isTTY),
			display.WithASCII(useASCII),
		)
	}

	var progressCallback func(done, total int, repo *models.Repository)
	if progress != nil {
		progressCallback = progress.ProgressCallback()
	}

	github.FetchAllPRs(repos, ghClient, progressCallback)

	// Clear progress display if used
	if progress != nil {
		progress.Clear()
	}

	// 8. Categorize
	cat := categorizer.NewCategorizer()
	result := cat.Categorize(repos, cfg, cfg.GitHubUsername)
	result.ScanDuration = time.Since(startTime)

	// 9. Render output
	output, err := display.Render(result, display.RenderOptions{
		ShowIcons:    cfg.ShowIcons,
		ShowBranches: cfg.ShowBranchName,
		ShowOtherPRs: cfg.ShowOtherPRs,
		NoColor:      noColor,
		JSON:         flagJSON,
	})
	if err != nil {
		return fmt.Errorf("render error: %w", err)
	}

	fmt.Print(output)
	return nil
}
