package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"prt/internal/config"
	"prt/internal/github"
)

// runWizard runs the interactive setup wizard.
func runWizard(cfg *config.Config) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Welcome to PRT (PR Tracker)! 🚀")
	fmt.Println()
	fmt.Println("Let's set up your configuration.")
	fmt.Println()

	// 1. GitHub username
	username, err := promptUsername(reader)
	if err != nil {
		return err
	}
	cfg.GitHubUsername = username

	// 2. Search paths
	paths, err := promptSearchPaths(reader)
	if err != nil {
		return err
	}
	if len(paths) > 0 {
		cfg.SearchPaths = paths
	}

	// 3. Team members
	members, err := promptTeamMembers(reader)
	if err != nil {
		return err
	}
	if len(members) > 0 {
		cfg.TeamMembers = members
	}

	// Save config
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Printf("✓ Configuration saved to %s\n", config.ConfigPath())
	fmt.Println()
	fmt.Println("Run `prt` to see your PR dashboard!")

	return nil
}

// promptUsername prompts for GitHub username with auto-detection option.
func promptUsername(reader *bufio.Reader) (string, error) {
	fmt.Print("? What is your GitHub username? (leave blank to auto-detect)\n> ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	username := strings.TrimSpace(input)

	if username == "" {
		// Try auto-detect
		fmt.Print("  Detecting GitHub username...")
		client := github.NewClient()
		detected, err := client.GetCurrentUser()
		if err != nil {
			fmt.Println(" failed")
			fmt.Println("  Could not auto-detect username. Please enter manually.")
			fmt.Print("> ")
			input, err = reader.ReadString('\n')
			if err != nil {
				return "", err
			}
			username = strings.TrimSpace(input)
			if username == "" {
				return "", fmt.Errorf("GitHub username is required")
			}
		} else {
			username = detected
			fmt.Printf(" ✓ %s\n", username)
		}
	}

	fmt.Println()
	return username, nil
}

// promptSearchPaths prompts for repository search paths.
func promptSearchPaths(reader *bufio.Reader) ([]string, error) {
	fmt.Print("? Where should PRT look for repositories?\n  (Enter paths separated by commas, ~ supported)\n> ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	input = strings.TrimSpace(input)

	if input == "" {
		fmt.Println("  Using default: ~/code, ~/projects")
		fmt.Println()
		return []string{"~/code", "~/projects"}, nil
	}

	// Parse comma-separated paths
	rawPaths := strings.Split(input, ",")
	var paths []string
	var validCount, invalidCount int

	for _, p := range rawPaths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		paths = append(paths, p)

		// Validate path exists
		expanded := expandPath(p)
		if _, err := os.Stat(expanded); os.IsNotExist(err) {
			fmt.Printf("  ⚠ Warning: Path does not exist: %s\n", p)
			invalidCount++
		} else {
			validCount++
		}
	}

	if validCount > 0 {
		fmt.Printf("  ✓ %d valid path(s) configured\n", validCount)
	}
	fmt.Println()

	return paths, nil
}

// promptTeamMembers prompts for team member usernames.
func promptTeamMembers(reader *bufio.Reader) ([]string, error) {
	fmt.Print("? Add team members? (GitHub usernames, comma-separated, blank to skip)\n> ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	input = strings.TrimSpace(input)

	if input == "" {
		fmt.Println("  Skipped team members")
		fmt.Println()
		return nil, nil
	}

	// Parse comma-separated usernames
	rawMembers := strings.Split(input, ",")
	var members []string

	for _, m := range rawMembers {
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		// Remove @ prefix if present
		m = strings.TrimPrefix(m, "@")
		members = append(members, m)
	}

	if len(members) > 0 {
		fmt.Printf("  ✓ Added %d team member(s)\n", len(members))
	}
	fmt.Println()

	return members, nil
}

// expandPath expands ~ to the user's home directory.
func expandPath(p string) string {
	if strings.HasPrefix(p, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return filepath.Join(home, p[1:])
	}
	return p
}
