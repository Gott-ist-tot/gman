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
	errorJSON          bool
	errorTable         bool
)

// errorsCmd represents the errors command for error handling testing and display
var errorsCmd = &cobra.Command{
	Use:   "errors",
	Short: "Error handling and display utilities",
	Long: `Error handling utilities for testing and demonstrating error display formats.

Examples:
  gman tools errors test              # Test error display formats
  gman tools errors demo             # Show various error types`,
}

// errorTestCmd tests error display formats
var errorTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test error display formats",
	Long: `Test and demonstrate different error display formats including compact, 
detailed, and JSON formats.`,
	RunE: runErrorTest,
}

// errorDemoCmd demonstrates various error types
var errorDemoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Demonstrate various error types",
	Long: `Demonstrate different types of errors that gman can encounter.`,
	RunE: runErrorDemo,
}

func init() {
	toolsCmd.AddCommand(errorsCmd)
	errorsCmd.AddCommand(errorTestCmd)
	errorsCmd.AddCommand(errorDemoCmd)

	// Test command flags
	errorTestCmd.Flags().StringVar(&errorDisplayMode, "mode", "detailed", "Display mode (compact, detailed, json)")
	errorTestCmd.Flags().BoolVarP(&errorVerbose, "verbose", "v", false, "Show verbose technical details")
	errorTestCmd.Flags().BoolVar(&errorNoColor, "no-color", false, "Disable colored output")
	errorTestCmd.Flags().BoolVar(&errorJSON, "json", false, "Use JSON format")

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
	testErr.WithSuggestion("Try adjusting the display mode with --mode flag")
	testErr.WithSuggestion("Use --verbose for more technical details")
	
	// Format and display the error
	formatter := errors.NewEnhancedErrorFormatter(config)
	formattedError := formatter.Format(testErr)
	
	fmt.Println("🧪 Error Display Test")
	fmt.Println("====================")
	fmt.Printf("Mode: %s\n", mode)
	fmt.Printf("Config: Color=%t, Verbose=%t\n\n", 
		config.ColorEnabled, config.Verbose)
	
	fmt.Println("Sample Error Output:")
	fmt.Println("--------------------")
	fmt.Println(formattedError)
	
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
	}{
		{
			name:  "Repository Not Found",
			error: errors.NewRepoNotFoundError("/nonexistent/path"),
		},
		{
			name:  "Network Timeout",
			error: errors.NewNetworkTimeoutError("git fetch", "30s"),
		},
		{
			name:  "Merge Conflict",
			error: errors.NewMergeConflictError("/home/user/project"),
		},
		{
			name:  "Tool Not Available",
			error: errors.NewToolNotAvailableError("fd", "brew install fd"),
		},
	}
	
	for i, demo := range demoErrors {
		fmt.Printf("%d. %s\n", i+1, demo.name)
		fmt.Println(strings.Repeat("-", len(demo.name)+4))
		
		// Add some suggestions
		demo.error.WithSuggestion("Check the documentation for this error type")
		demo.error.WithSuggestion("Use 'gman health' to analyze system status")
		
		// Format and display
		formatted := formatter.Format(demo.error)
		fmt.Println(formatted)
		
		fmt.Println("\n" + strings.Repeat("=", 50) + "\n")
	}
	
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
	
	switch errorDisplayMode {
	case "compact":
		return errors.DisplayModeCompact
	case "detailed":
		return errors.DisplayModeDetailed
	case "json":
		return errors.DisplayModeJSON
	default:
		return errors.DisplayModeDetailed
	}
}

func createErrorDisplayConfig(mode errors.DisplayMode) *errors.ErrorDisplayConfig {
	config := errors.DefaultDisplayConfig()
	config.Mode = mode
	config.Verbose = errorVerbose
	config.ColorEnabled = !errorNoColor
	
	// Adjust settings based on mode
	switch mode {
	case errors.DisplayModeCompact:
		config = errors.CompactDisplayConfig()
		config.ColorEnabled = !errorNoColor
	case errors.DisplayModeJSON:
		config.ColorEnabled = false
	}
	
	return config
}