package cli

import (
	"fmt"
	"os"
	"os/exec"

	"prt/internal/config"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and manage PRT configuration",
	Long: `View and manage PRT configuration.

Subcommands:
  show    Display the current configuration
  path    Show the path to the config file
  edit    Open the config file in your editor`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long:  "Display the current PRT configuration in YAML format.",
	RunE:  runConfigShow,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show config file path",
	Long:  "Show the path to the PRT configuration file.",
	RunE:  runConfigPath,
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open config in editor",
	Long: `Open the PRT configuration file in your editor.

Uses the EDITOR environment variable, falling back to vi.`,
	RunE: runConfigEdit,
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configEditCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	// Load config without validation (show even incomplete configs)
	cfg, err := config.Load(nil)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Generate formatted YAML output
	output, err := config.GenerateConfigFile(cfg)
	if err != nil {
		return fmt.Errorf("failed to format config: %w", err)
	}

	fmt.Print(output)
	return nil
}

func runConfigPath(cmd *cobra.Command, args []string) error {
	path := config.ConfigPath()

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("%s (not created yet)\n", path)
	} else {
		fmt.Println(path)
	}

	return nil
}

func runConfigEdit(cmd *cobra.Command, args []string) error {
	path := config.ConfigPath()

	// Ensure config directory exists
	if err := os.MkdirAll(config.ConfigDir(), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create default config if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := config.LoadDefault()
		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
		fmt.Printf("Created new config file at %s\n", path)
	}

	// Get editor from environment
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}

	// Open editor
	editorCmd := exec.Command(editor, path)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	return nil
}
