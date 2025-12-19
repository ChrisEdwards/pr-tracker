// Package cli provides the command-line interface for prt.
package cli

import (
	"fmt"
)

// Execute runs the CLI with the given version string.
// This is a stub that will be replaced with actual CLI implementation.
func Execute(version string) error {
	fmt.Printf("prt version %s\n", version)
	return nil
}
