package cmd

import (
	"fmt"

	"gman/internal/errors"

	"github.com/spf13/cobra"
)

var (
	healthVerbose bool
	healthJSON    bool
)

// healthCmd represents the health command
var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Display system health and error analysis",
	Long: `Display a comprehensive health report including error patterns, 
system metrics, and recommendations for improving system stability.

The health command analyzes recent error patterns and provides:
- Overall system health status
- Common error patterns and their frequency
- Preventive tips to avoid future issues
- Recommended actions for system improvement

Examples:
  gman tools health                 # Show basic health report
  gman tools health --verbose      # Show detailed health information
  gman tools health --json         # Output health report in JSON format`,
	RunE: runHealth,
}

func init() {
	toolsCmd.AddCommand(healthCmd)
	
	healthCmd.Flags().BoolVarP(&healthVerbose, "verbose", "v", false, "Show detailed health information")
	healthCmd.Flags().BoolVar(&healthJSON, "json", false, "Output health report in JSON format")
}

func runHealth(cmd *cobra.Command, args []string) error {
	// Generate health report
	manager := errors.GetGlobalManager()
	report := manager.GenerateHealthReport()

	if healthJSON {
		// Output JSON format (could be enhanced with proper JSON marshaling)
		fmt.Printf(`{
  "generated_at": "%s",
  "overall_health": "%s",
  "total_errors_24h": "%s",
  "resolution_rate": "%s",
  "most_common_error": "%s"
}`, 
			report.GeneratedAt.Format("2006-01-02T15:04:05Z07:00"),
			report.OverallHealth,
			report.SystemMetrics["total_errors_24h"],
			report.SystemMetrics["resolution_rate"],
			report.SystemMetrics["most_common_error"])
		return nil
	}

	// Display formatted health report
	healthOutput := errors.FormatHealthReport(report)
	fmt.Print(healthOutput)

	// Show additional details in verbose mode
	if healthVerbose {
		fmt.Println("\n📋 Detailed Analysis:")
		
		if len(report.ErrorPatterns) > 0 {
			fmt.Println("\nError Pattern Details:")
			for i, pattern := range report.ErrorPatterns {
				if i >= 5 { // Limit to top 5 in verbose mode
					break
				}
				
				fmt.Printf("\n%d. %s (occurred %d times)\n", 
					i+1, pattern.ErrorType, pattern.Frequency)
				fmt.Printf("   Last occurrence: %s\n", 
					pattern.LastOccurrence.Format("2006-01-02 15:04:05"))
				
				if len(pattern.Contexts) > 0 {
					fmt.Printf("   Contexts: %v\n", pattern.Contexts)
				}
				
				if len(pattern.Suggestions) > 0 {
					fmt.Println("   Suggestions:")
					for _, suggestion := range pattern.Suggestions {
						fmt.Printf("     • %s\n", suggestion)
					}
				}
			}
		}

		// Show all preventive tips in verbose mode
		if len(report.PreventiveTips) > 3 {
			fmt.Println("\n💡 Additional Preventive Tips:")
			for i, tip := range report.PreventiveTips[3:] {
				fmt.Printf("  %d. %s\n", i+4, tip)
			}
		}

		// Show error history summary
		history := manager.GetErrorHistory()
		if len(history) > 0 {
			fmt.Printf("\n📊 Recent Error Summary (%d total events):\n", len(history))
			
			// Group by type for summary
			typeCounts := make(map[errors.ErrorType]int)
			for _, event := range history {
				typeCounts[event.Error.Type]++
			}
			
			for errorType, count := range typeCounts {
				fmt.Printf("   %s: %d occurrences\n", errorType, count)
			}
		}
	}

	return nil
}