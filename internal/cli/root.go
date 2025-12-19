// Package cli provides the command-line interface for prt.
package cli

import (
	"fmt"
	"os"
	"time"

	"prt/internal/categorizer"
	"prt/internal/config"
	"prt/internal/display"
	"prt/internal/github"
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

	// 4. Check gh CLI and get client
	ghClient := github.NewClient()
	if err := ghClient.Check(); err != nil {
		return err
	}

	// 5. Auto-detect username if needed
	if cfg.GitHubUsername == "" {
		user, err := ghClient.GetCurrentUser()
		if err != nil {
			return fmt.Errorf("cannot determine GitHub username: %w", err)
		}
		cfg.GitHubUsername = user
	}

	// 6. Scan for repositories
	scnr, err := scanner.NewScanner(cfg.ScanDepth, cfg.IncludeRepos)
	if err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	repos, err := scnr.Scan(cfg)
	if err != nil {
		return fmt.Errorf("scan error: %w", err)
	}

	if len(repos) == 0 {
		fmt.Println("No Git repositories found in configured paths.")
		return nil
	}

	// 7. Fetch PRs with progress
	// Progress callback is nil for now - will be implemented with progress display
	github.FetchAllPRs(repos, ghClient, nil)

	// 8. Categorize
	cat := categorizer.NewCategorizer()
	result := cat.Categorize(repos, cfg, cfg.GitHubUsername)
	result.ScanDuration = time.Since(startTime)

	// 9. Render output
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
