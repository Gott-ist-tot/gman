package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gman/internal/config"
	"gman/internal/display"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [path] [alias]",
	Short: "Add a repository to gman configuration",
	Long: `Add a Git repository to gman configuration.

Usage:
  gman add                    # Add current directory with auto-generated alias
  gman add <path>             # Add specified path with auto-generated alias  
  gman add <path> <alias>     # Add path with custom alias
  gman add . <alias>          # Add current directory with custom alias

The path must be a valid Git repository (contain .git directory).`,
	Args: cobra.RangeArgs(0, 2),
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	var path, alias string

	// Parse arguments
	switch len(args) {
	case 0:
		// Use current directory
		var err error
		path, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		alias = generateAlias(path)
	case 1:
		// Use specified path, generate alias
		path = args[0]
		if path == "." {
			var err error
			path, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
		}
		alias = generateAlias(path)
	case 2:
		// Use specified path and alias
		path = args[0]
		alias = args[1]
		if path == "." {
			var err error
			path, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
		}
	}

	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if alias already exists
	cfg := configMgr.GetConfig()
	if _, exists := cfg.Repositories[alias]; exists {
		return fmt.Errorf("alias '%s' already exists", alias)
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
