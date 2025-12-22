package cli

import (
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
