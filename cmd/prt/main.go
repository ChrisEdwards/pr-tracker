package main

import (
	"fmt"
	"os"

	"prt/internal/cli"
)

var version = "dev" // Set by ldflags at build time

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Fatal error: %v\n", r)
			os.Exit(1)
		}
	}()

	if err := cli.Execute(version); err != nil {
		os.Exit(1)
	}
}
