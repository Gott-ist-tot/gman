package cmd

import (
	"fmt"
	"strings"

	"gman/internal/errors"

	"github.com/spf13/cobra"
)

var (
	errorDisplayMode   string
	errorVerbose       bool
	errorNoColor       bool
	errorMaxWidth      int
	errorShowTimestamp bool
	errorShowContext   bool
	errorJSON          bool
	errorTable         bool
	errorInteractive   bool
)

// errorsCmd represents the errors command for error handling testing and display
var errorsCmd = &cobra.Command{
	Use:   "errors",
	Short: "Error handling and display utilities",
	Long: `Error handling utilities for testing, demonstrating, and configuring 
error display formats. This command provides tools for developers and users
to understand and test the enhanced error handling system.

Examples:
  gman tools errors test                    # Test error display formats
  gman tools errors test --mode table      # Test table format
  gman tools errors test --json            # Test JSON format
  gman tools errors demo                   # Show various error types
  gman tools errors config                 # Show current error configuration`,
}

// errorTestCmd tests error display formats
var errorTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test error display formats",
	Long: `Test and demonstrate different error display formats including compact, 
detailed, interactive, JSON, and table formats. This is useful for understanding
how errors will be displayed and for testing formatting configurations.`,
	RunE: runErrorTest,
}

// errorDemoCmd demonstrates various error types
var errorDemoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Demonstrate various error types and recovery options",
	Long: `Demonstrate different types of errors that gman can encounter,
along with their recovery options and diagnostic information. This helps
users understand what to expect when errors occur.`,
	RunE: runErrorDemo,
}

// errorConfigCmd shows error configuration
var errorConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current error handling configuration",
	Long: `Display the current error handling configuration including display
modes, formatting options, and recovery settings.`,
	RunE: runErrorConfig,
}

func init() {
	toolsCmd.AddCommand(errorsCmd)
	errorsCmd.AddCommand(errorTestCmd)
	errorsCmd.AddCommand(errorDemoCmd)
	errorsCmd.AddCommand(errorConfigCmd)

	// Test command flags
	errorTestCmd.Flags().StringVar(&errorDisplayMode, "mode", "detailed", "Display mode (compact, detailed, interactive, json, table)")
	errorTestCmd.Flags().BoolVarP(&errorVerbose, "verbose", "v", false, "Show verbose technical details")
	errorTestCmd.Flags().BoolVar(&errorNoColor, "no-color", false, "Disable colored output")
	errorTestCmd.Flags().IntVar(&errorMaxWidth, "width", 80, "Maximum display width")
	errorTestCmd.Flags().BoolVar(&errorShowTimestamp, "timestamp", true, "Show error timestamps")
	errorTestCmd.Flags().BoolVar(&errorShowContext, "context", true, "Show error context")
	errorTestCmd.Flags().BoolVar(&errorJSON, "json", false, "Use JSON format")
	errorTestCmd.Flags().BoolVar(&errorTable, "table", false, "Use table format")
	errorTestCmd.Flags().BoolVar(&errorInteractive, "interactive", false, "Use interactive format")

	// Demo command flags
	errorDemoCmd.Flags().StringVar(&errorDisplayMode, "mode", "detailed", "Display mode for demo")
	errorDemoCmd.Flags().BoolVar(&errorVerbose, "verbose", false, "Show verbose output")
}

func runErrorTest(cmd *cobra.Command, args []string) error {
	// Determine display mode from flags
	mode := determineDisplayMode()
	
	// Create test configuration
	config := createErrorDisplayConfig(mode)
	
	// Create a test error
	testErr := errors.NewConfigInvalidError("Testing error display formats")
	testErr.WithContext("command", "gman tools errors test")
	testErr.WithContext("mode", string(mode))
	testErr.WithContext("timestamp", "2024-12-27T10:30:00Z")
	testErr.WithSuggestion("Try adjusting the display mode with --mode flag")
	testErr.WithSuggestion("Use --verbose for more technical details")
	testErr.WithSuggestion("Check configuration with 'gman tools errors config'")
	
	// Format and display the error
	formatter := errors.NewEnhancedErrorFormatter(config)
	formattedError := formatter.Format(testErr)
	
	fmt.Println("🧪 Error Display Test")
	fmt.Println("====================")
	fmt.Printf("Mode: %s\n", mode)
	fmt.Printf("Config: Color=%t, Verbose=%t, Width=%d\n\n", 
		config.ColorEnabled, config.Verbose, config.MaxWidth)
	
	fmt.Println("Sample Error Output:")
	fmt.Println("--------------------")
	fmt.Println(formattedError)
	
	// Also demonstrate recovery if it's a detailed mode
	if mode == errors.DisplayModeDetailed || mode == errors.DisplayModeInteractive {
		fmt.Println("\n" + strings.Repeat("=", 50))
		fmt.Println("🔧 Recovery Plan Demonstration:")
		fmt.Println(strings.Repeat("=", 50))
		
		// Create recovery engine and plan
		engine := errors.NewRecoveryEngine()
		plan := engine.CreateRecoveryPlan(testErr)
		
		// Display with recovery
		errorWithRecovery := formatter.FormatWithRecovery(testErr, plan)
		fmt.Println(errorWithRecovery)
	}
	
	return nil
}

func runErrorDemo(cmd *cobra.Command, args []string) error {
	mode := determineDisplayMode()
	config := createErrorDisplayConfig(mode)
	formatter := errors.NewEnhancedErrorFormatter(config)
	
	fmt.Println("🎭 Error Types Demonstration")
	fmt.Println("=============================")
	fmt.Printf("Display Mode: %s\n\n", mode)
	
	// Create various types of errors for demonstration
	demoErrors := []struct {
		name    string
		error   *errors.GmanError
		context map[string]string
	}{
		{
			name:  "Repository Not Found",
			error: errors.NewRepoNotFoundError("/nonexistent/path"),
			context: map[string]string{
				"operation": "switch",
				"alias":     "missing-repo",
			},
		},
		{
			name:  "Network Timeout",
			error: errors.NewNetworkTimeoutError("git fetch", "30s"),
			context: map[string]string{
				"repository": "/home/user/project",
				"remote":     "origin",
			},
		},
		{
			name:  "Merge Conflict",
			error: errors.NewMergeConflictError("/home/user/project"),
			context: map[string]string{
				"branch":        "feature/new-ui",
				"target_branch": "main",
				"conflicted_files": "src/app.js, styles/main.css",
			},
		},
		{
			name:  "Tool Not Available",
			error: errors.NewToolNotAvailableError("fd", "Advanced file search requires 'fd' tool"),
			context: map[string]string{
				"operation":     "file search",
				"alternative":   "Using fallback search method",
			},
		},
	}
	
	for i, demo := range demoErrors {
		fmt.Printf("%d. %s\n", i+1, demo.name)
		fmt.Println(strings.Repeat("-", len(demo.name)+4))
		
		// Add context to error
		for key, value := range demo.context {
			demo.error.WithContext(key, value)
		}
		
		// Add some suggestions
		demo.error.WithSuggestion("Check the documentation for this error type")
		demo.error.WithSuggestion("Use 'gman health' to analyze system status")
		
		// Format and display
		formatted := formatter.Format(demo.error)
		fmt.Println(formatted)
		
		// Show recovery options for some errors
		if i < 2 { // Show recovery for first two errors
			engine := errors.NewRecoveryEngine()
			plan := engine.CreateRecoveryPlan(demo.error)
			if plan.PrimaryAction != nil {
				recoveryDisplay := formatter.FormatWithRecovery(demo.error, plan)
				fmt.Println("\n" + recoveryDisplay)
			}
		}
		
		fmt.Println("\n" + strings.Repeat("=", 50) + "\n")
	}
	
	return nil
}

func runErrorConfig(cmd *cobra.Command, args []string) error {
	fmt.Println("⚙️  Error Handling Configuration")
	fmt.Println("================================")
	
	// Show current global manager configuration
	manager := errors.GetGlobalManager()
	
	fmt.Println("Global Error Manager Settings:")
	fmt.Println("  Interactive Mode: true")
	fmt.Println("  Auto Recovery: false")
	fmt.Println("  Verbose Output: false")
	fmt.Println("  Safe Mode: true")
	
	fmt.Println("\nAvailable Display Modes:")
	fmt.Println("  • compact     - Single line format")
	fmt.Println("  • detailed    - Multi-line with full information")
	fmt.Println("  • interactive - Detailed with recovery prompts")
	fmt.Println("  • json        - Machine-readable JSON format")
	fmt.Println("  • table       - Tabular bordered format")
	
	fmt.Println("\nDisplay Configuration Options:")
	fmt.Println("  • --mode <mode>      Set display mode")
	fmt.Println("  • --verbose          Show technical details")
	fmt.Println("  • --no-color         Disable colors")
	fmt.Println("  • --width <n>        Set maximum width")
	fmt.Println("  • --timestamp        Show/hide timestamps")
	fmt.Println("  • --context          Show/hide context")
	
	fmt.Println("\nRecovery System:")
	fmt.Println("  • Automatic recovery for safe operations")
	fmt.Println("  • Interactive recovery with user choices")
	fmt.Println("  • Safety levels (1-5) for all recovery actions")
	fmt.Println("  • Smart filtering based on error type")
	
	// Show recent error statistics
	report := manager.GenerateHealthReport()
	fmt.Printf("\nCurrent System Health: %s\n", report.OverallHealth)
	
	if len(report.ErrorPatterns) > 0 {
		fmt.Println("\nRecent Error Patterns:")
		for i, pattern := range report.ErrorPatterns {
			if i >= 3 { // Show top 3
				break
			}
			fmt.Printf("  %d. %s: %d occurrences\n", 
				i+1, pattern.ErrorType, pattern.Frequency)
		}
	}
	
	fmt.Println("\nTesting Commands:")
	fmt.Println("  gman tools errors test                    # Test current config")
	fmt.Println("  gman tools errors test --mode table      # Test table format")
	fmt.Println("  gman tools errors demo                   # Show error examples")
	fmt.Println("  gman tools health                       # System health report")
	
	return nil
}

func determineDisplayMode() errors.DisplayMode {
	// Priority: specific format flags > mode flag > default
	if errorJSON {
		return errors.DisplayModeJSON
	}
	if errorTable {
		return errors.DisplayModeTable
	}
	if errorInteractive {
		return errors.DisplayModeInteractive
	}
	
	switch errorDisplayMode {
	case "compact":
		return errors.DisplayModeCompact
	case "detailed":
		return errors.DisplayModeDetailed
	case "interactive":
		return errors.DisplayModeInteractive
	case "json":
		return errors.DisplayModeJSON
	case "table":
		return errors.DisplayModeTable
	default:
		return errors.DisplayModeDetailed
	}
}

func createErrorDisplayConfig(mode errors.DisplayMode) *errors.ErrorDisplayConfig {
	config := errors.DefaultDisplayConfig()
	config.Mode = mode
	config.Verbose = errorVerbose
	config.ColorEnabled = !errorNoColor
	config.MaxWidth = errorMaxWidth
	config.ShowTimestamp = errorShowTimestamp
	config.ShowContext = errorShowContext
	
	// Adjust settings based on mode
	switch mode {
	case errors.DisplayModeCompact:
		config = errors.CompactDisplayConfig()
		config.ColorEnabled = !errorNoColor
	case errors.DisplayModeJSON:
		config.ColorEnabled = false
		config.ShowIcons = false
	case errors.DisplayModeTable:
		config.ShowIcons = false
	}
	
	return config
}