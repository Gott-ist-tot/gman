package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"gman/internal/di"
	"gman/internal/display"
	"gman/internal/errors"

	"github.com/spf13/cobra"
)

var addPath string

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add <alias>",
	Short: "Add a repository to gman configuration",
	Long: `Add a Git repository to gman configuration.

Usage:
  gman repo add <alias>                    # Add current directory with specified alias
  gman repo add <alias> --path <path>      # Add specified path with alias

The path must be a valid Git repository (contain .git directory).
If no path is specified, the current directory will be used.`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

func init() {
	// Command is now available via: gman repo add
	// Removed direct rootCmd registration to avoid duplication
	addCmd.Flags().StringVar(&addPath, "path", "", "Path to the Git repository (default: current directory)")
}

func runAdd(cmd *cobra.Command, args []string) error {
	var path string

	// Get alias from argument
	alias := args[0]

	// Determine path
	if addPath != "" {
		path = addPath
		if path == "." {
			var err error
			path, err = os.Getwd()
			if err != nil {
				return errors.Wrap(err, errors.ErrTypeInternal, 
					"failed to get current directory").
					WithSuggestion("Ensure you have read permissions for the current directory")
			}
		}
	} else {
		// Use current directory if no path specified
		var err error
		path, err = os.Getwd()
		if err != nil {
			return errors.Wrap(err, errors.ErrTypeInternal, 
				"failed to get current directory").
				WithSuggestion("Ensure you have read permissions for the current directory")
		}
	}

	// Configuration is already loaded by root command's PersistentPreRunE
	configMgr := di.ConfigManager()

	// Check if alias already exists
	cfg := configMgr.GetConfig()
	if _, exists := cfg.Repositories[alias]; exists {
		return errors.NewRepoAlreadyExistsError(alias)
	}

	// Add repository
	if err := configMgr.AddRepository(alias, path); err != nil {
		return err
	}

	// Get absolute path for display
	absPath, _ := filepath.Abs(path)
	display.PrintSuccess(fmt.Sprintf("Added repository: %s -> %s", alias, absPath))

	return nil
}

func generateAlias(path string) string {
	// Use the base directory name as alias
	return filepath.Base(path)
}
