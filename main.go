package main

import (
	"os"

	"github.com/SantiagoBobrik/claude-pulse/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
