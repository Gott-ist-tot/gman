package cmd

import (
	"fmt"
	"os"

	"gman/internal/di"

	"github.com/spf13/cobra"
)

var (
	migrateDryRun  bool
	migrateApply   bool
	migrateVerbose bool
)

// migrateDICmd represents the migrate-di command for internal development
var migrateDICmd = &cobra.Command{
	Use:   "migrate-di",
	Short: "Migrate codebase to use dependency injection container",
	Long: `Analyze and migrate the codebase to consistently use the dependency injection container
instead of manual instantiation of managers.

This is a development/maintenance command to ensure consistent DI usage across the codebase.

Examples:
  gman migrate-di                    # Analyze current DI usage
  gman migrate-di --dry-run          # Preview migration changes
  gman migrate-di --apply            # Apply automatic migration
  gman migrate-di --apply --verbose  # Apply with detailed output`,
	Hidden: true, // Hidden development command
	RunE:   runMigrateDI,
}

func init() {
	rootCmd.AddCommand(migrateDICmd)
	migrateDICmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "Preview migration changes without applying")
	migrateDICmd.Flags().BoolVar(&migrateApply, "apply", false, "Apply automatic migration")
	migrateDICmd.Flags().BoolVar(&migrateVerbose, "verbose", false, "Show detailed migration output")
}

func runMigrateDI(cmd *cobra.Command, args []string) error {
	// Get current working directory as project root
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if migrateApply || migrateDryRun {
		// Apply automatic migration
		fmt.Println("🔄 Starting dependency injection migration...")
		if err := di.ApplyAutomaticMigration(projectRoot, migrateDryRun); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}

		if migrateDryRun {
			fmt.Println("\n✅ Dry-run completed. Use --apply to execute changes.")
		} else {
			fmt.Println("\n✅ Migration completed successfully!")
			fmt.Println("🔧 Next steps:")
			fmt.Println("  1. Run 'make test' to verify functionality")
			fmt.Println("  2. Initialize DI container in main.go if not already done")
			fmt.Println("  3. Review any remaining manual instantiations")
		}
	} else {
		// Analyze current usage
		report, err := di.AnalyzeDependencyUsage(projectRoot)
		if err != nil {
			return fmt.Errorf("analysis failed: %w", err)
		}

		report.PrintReport()

		if len(report.ManualInstantiations) > 0 {
			fmt.Println("\n💡 To apply automatic migration, run: gman migrate-di --apply")
			fmt.Println("💡 To preview changes first, run: gman migrate-di --dry-run")
		}
	}

	return nil
}
