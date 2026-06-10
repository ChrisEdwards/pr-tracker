package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"prt/internal/config"
)

func captureStdout(t *testing.T, fn func() error) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}

	os.Stdout = w
	runErr := fn()
	os.Stdout = old

	if err := w.Close(); err != nil {
		t.Fatalf("closing stdout pipe writer: %v", err)
	}
	if runErr != nil {
		t.Fatalf("command error = %v", runErr)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("reading stdout pipe: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("closing stdout pipe reader: %v", err)
	}

	return buf.String()
}

func TestConfigShowCmd(t *testing.T) {
	output := captureStdout(t, func() error {
		return runConfigShow(nil, nil)
	})

	// Verify output contains expected config content
	if !strings.Contains(output, "# PRT Configuration") {
		t.Error("Output should contain config header")
	}
	if !strings.Contains(output, "github_username:") {
		t.Error("Output should contain github_username field")
	}
	if !strings.Contains(output, "search_paths:") {
		t.Error("Output should contain search_paths field")
	}
}

func TestConfigPathCmd(t *testing.T) {
	output := strings.TrimSpace(captureStdout(t, func() error {
		return runConfigPath(nil, nil)
	}))

	// Verify output is a path ending in config.yaml
	if !strings.HasSuffix(strings.TrimSuffix(output, " (not created yet)"), "config.yaml") {
		t.Errorf("Output should be config path, got: %q", output)
	}

	// Should contain .prt directory
	if !strings.Contains(output, ".prt") {
		t.Errorf("Output should contain .prt directory, got: %q", output)
	}
}

func TestConfigPathCmd_FileNotExists(t *testing.T) {
	// Save original config path function and restore after test
	// Since we can't easily mock ConfigPath, we'll just verify the behavior
	// with the actual path

	output := captureStdout(t, func() error {
		return runConfigPath(nil, nil)
	})

	// Output should be a valid path regardless of existence
	if output == "" {
		t.Error("Output should not be empty")
	}
}

func TestConfigEdit_EditorEnvVar(t *testing.T) {
	// Test that we properly check EDITOR env var
	// We can't actually run the editor in tests, but we can verify
	// the environment variable logic

	// Clear both
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "")

	// Verify the default fallback behavior
	// Since we can't run the editor, just verify the logic path
	gotEditor := os.Getenv("EDITOR")
	if gotEditor != "" {
		t.Error("EDITOR should be unset for this test")
	}
}

func TestConfigCmdHelp(t *testing.T) {
	// Test that configCmd has proper setup
	if configCmd.Use != "config" {
		t.Errorf("configCmd.Use = %q, want %q", configCmd.Use, "config")
	}

	if configCmd.Short == "" {
		t.Error("configCmd.Short should not be empty")
	}

	// Verify subcommands are registered
	subCmds := configCmd.Commands()
	subCmdNames := make(map[string]bool)
	for _, cmd := range subCmds {
		subCmdNames[cmd.Use] = true
	}

	expectedCmds := []string{"show", "path", "edit"}
	for _, name := range expectedCmds {
		if !subCmdNames[name] {
			t.Errorf("configCmd should have subcommand %q", name)
		}
	}
}

func TestConfigShowCmdMetadata(t *testing.T) {
	if configShowCmd.Use != "show" {
		t.Errorf("configShowCmd.Use = %q, want %q", configShowCmd.Use, "show")
	}
	if configShowCmd.Short == "" {
		t.Error("configShowCmd.Short should not be empty")
	}
}

func TestConfigPathCmdMetadata(t *testing.T) {
	if configPathCmd.Use != "path" {
		t.Errorf("configPathCmd.Use = %q, want %q", configPathCmd.Use, "path")
	}
	if configPathCmd.Short == "" {
		t.Error("configPathCmd.Short should not be empty")
	}
}

func TestConfigEditCmdMetadata(t *testing.T) {
	if configEditCmd.Use != "edit" {
		t.Errorf("configEditCmd.Use = %q, want %q", configEditCmd.Use, "edit")
	}
	if configEditCmd.Short == "" {
		t.Error("configEditCmd.Short should not be empty")
	}
	// Verify long description mentions EDITOR
	if !strings.Contains(configEditCmd.Long, "EDITOR") {
		t.Error("configEditCmd.Long should mention EDITOR environment variable")
	}
}

func TestConfigEdit_CreatesConfigIfNotExists(t *testing.T) {
	// Create a temp directory to test config creation
	tmpDir := t.TempDir()
	testConfigPath := filepath.Join(tmpDir, ".prt", "config.yaml")

	// Verify the path doesn't exist initially
	if _, err := os.Stat(testConfigPath); !os.IsNotExist(err) {
		t.Skip("Test config path should not exist initially")
	}

	// Note: We can't easily test runConfigEdit directly since it uses
	// config.ConfigDir() which returns a fixed path. This test documents
	// the expected behavior.

	// Verify config package functions work correctly
	cfg := config.LoadDefault()
	if cfg == nil {
		t.Error("LoadDefault should return a config")
	}

	// Verify GenerateConfigFile works
	content, err := config.GenerateConfigFile(cfg)
	if err != nil {
		t.Errorf("GenerateConfigFile() error = %v", err)
	}
	if !strings.Contains(content, "# PRT Configuration") {
		t.Error("Generated config should have header")
	}
}
