package batch

import (
	"github.com/spf13/cobra"
)

// RegisterBatchCommands registers all batch commands with the root command
func RegisterBatchCommands(rootCmd *cobra.Command) {
	// Add individual batch commands to root
	rootCmd.AddCommand(NewCommitCmd())
	rootCmd.AddCommand(NewPushCmd())
	rootCmd.AddCommand(NewPullCmd())
	rootCmd.AddCommand(NewStashCmd())
}

// Init is called to initialize the batch package
func Init(rootCmd *cobra.Command) {
	RegisterBatchCommands(rootCmd)
}
