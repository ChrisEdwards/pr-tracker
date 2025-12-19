package cli

import (
	"fmt"

	"prt/internal/config"
)

// runWizard runs the interactive setup wizard.
// This is a stub that will be replaced with full implementation in prt-0fv.
func runWizard(cfg *config.Config) error {
	fmt.Println("Welcome to PRT (PR Tracker)! 🚀")
	fmt.Println()
	fmt.Println("PRT needs to be configured before first use.")
	fmt.Println()
	fmt.Printf("Please create a config file at: %s\n", config.ConfigPath())
	fmt.Println()
	fmt.Println("Example configuration:")
	fmt.Println()
	fmt.Println("  github_username: your-username")
	fmt.Println("  search_paths:")
	fmt.Println("    - ~/code")
	fmt.Println("    - ~/projects")
	fmt.Println()
	fmt.Println("Run `prt` again after creating the config file.")
	return nil
}
