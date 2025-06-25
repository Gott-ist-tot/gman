package cmd

import (
	"gman/cmd/batch"
)

// Initialize batch commands
func init() {
	// Register all batch commands with the root command
	batch.Init(rootCmd)
}
