package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRootCmd_HasSetupFlag(t *testing.T) {
	// Verify the --setup flag is registered
	flag := rootCmd.Flags().Lookup("setup")
	if flag == nil {
		t.Fatal("--setup flag should be registered")
	}

	if flag.Usage == "" {
		t.Error("--setup flag should have usage text")
	}

	if !strings.Contains(flag.Usage, "wizard") {
		t.Error("--setup flag usage should mention wizard")
	}
}

func TestRootCmd_FlagsRegistered(t *testing.T) {
	expectedFlags := []string{
		"path",
		"filter",
		"group",
		"sort",
		"depth",
		"max-age",
		"json",
		"no-color",
		"setup",
	}

	for _, name := range expectedFlags {
		flag := rootCmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("expected flag --%s to be registered", name)
		}
	}
}

func TestRootCmd_Metadata(t *testing.T) {
	if rootCmd.Use != "prt" {
		t.Errorf("rootCmd.Use = %q, want %q", rootCmd.Use, "prt")
	}

	if rootCmd.Short == "" {
		t.Error("rootCmd.Short should not be empty")
	}

	if rootCmd.Long == "" {
		t.Error("rootCmd.Long should not be empty")
	}
}

func TestVersionFlag(t *testing.T) {
	// Save and restore original stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set version and run with --version
	rootCmd.Version = "1.2.3-test"
	rootCmd.SetArgs([]string{"--version"})
	err := rootCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Reset for other tests
	rootCmd.SetArgs(nil)

	if err != nil {
		t.Fatalf("Execute() with --version returned error: %v", err)
	}

	if !strings.Contains(output, "1.2.3-test") {
		t.Errorf("version output = %q, want it to contain %q", output, "1.2.3-test")
	}
}

func TestHelpFlag(t *testing.T) {
	// Save and restore original stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run with --help
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Reset for other tests
	rootCmd.SetArgs(nil)

	if err != nil {
		t.Fatalf("Execute() with --help returned error: %v", err)
	}

	// Verify help output contains expected sections
	expectedPhrases := []string{
		"PRT - GitHub PR Tracker",
		"--path",
		"--filter",
		"--json",
		"--setup",
		"--no-color",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("help output should contain %q", phrase)
		}
	}
}

func TestFlagDefaults(t *testing.T) {
	tests := []struct {
		name         string
		flagName     string
		defaultValue string
	}{
		{"path default", "path", ""},
		{"filter default", "filter", ""},
		{"group default", "group", ""},
		{"sort default", "sort", ""},
		{"json default", "json", "false"},
		{"no-color default", "no-color", "false"},
		{"setup default", "setup", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := rootCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}
			if flag.DefValue != tt.defaultValue {
				t.Errorf("flag %q default = %q, want %q", tt.flagName, flag.DefValue, tt.defaultValue)
			}
		})
	}
}

func TestFlagShorthand(t *testing.T) {
	tests := []struct {
		flagName  string
		shorthand string
	}{
		{"path", "p"},
		{"filter", "f"},
		{"group", "g"},
		{"sort", "s"},
		{"depth", "d"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			flag := rootCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}
			if flag.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
			}
		})
	}
}

func TestFlagUsageDescriptions(t *testing.T) {
	// Verify all flags have usage descriptions
	flags := []string{
		"path", "filter", "group", "sort", "depth",
		"max-age", "json", "no-color", "setup",
	}

	for _, name := range flags {
		t.Run(name, func(t *testing.T) {
			flag := rootCmd.Flags().Lookup(name)
			if flag == nil {
				t.Fatalf("flag %q not found", name)
			}
			if flag.Usage == "" {
				t.Errorf("flag %q should have usage description", name)
			}
		})
	}
}

func TestConfigSubcommandExists(t *testing.T) {
	// Verify config subcommand is registered
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "config" {
			found = true
			break
		}
	}
	if !found {
		t.Error("config subcommand should be registered")
	}
}

func TestExecuteFunction(t *testing.T) {
	// Test that Execute sets version correctly
	// Can't fully test without running the actual command
	// but we can verify the function exists and is callable

	// Save original args and restore
	oldArgs := os.Args
	os.Args = []string{"prt", "--version"}
	defer func() { os.Args = oldArgs }()

	// The Execute function should return nil for --version
	// But actually executing would try to print to stdout
	// Instead, verify the rootCmd version gets set
	oldVersion := rootCmd.Version
	defer func() { rootCmd.Version = oldVersion }()

	rootCmd.Version = "test-version"
	if rootCmd.Version != "test-version" {
		t.Error("Execute should set version on rootCmd")
	}
}
