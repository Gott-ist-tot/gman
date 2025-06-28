package cmd

import (
	"fmt"
	"os"
	"strings"

	"gman/internal/errors"
	"gman/internal/external"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	healthVerbose    bool
	healthJSON       bool
	healthFix        bool
	healthInstallCmd bool
)

// healthCmd represents the health command
var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Comprehensive system health and diagnostics",
	Long: `Display comprehensive system health status including:
- External tool availability and versions
- Configuration validation
- Repository health checks
- Error pattern analysis
- System dependencies

Examples:
  gman tools health                    # Show comprehensive health report
  gman tools health --verbose         # Show detailed diagnostic information
  gman tools health --fix             # Show installation commands for missing tools
  gman tools health --install         # Show installation commands for all tools
  gman tools health --json            # Output health report in JSON format`,
	RunE: runHealth,
}

func init() {
	// Note: healthCmd is no longer added to toolsCmd here as it's directly added in tools.go
	
	healthCmd.Flags().BoolVarP(&healthVerbose, "verbose", "v", false, "Show detailed diagnostic information")
	healthCmd.Flags().BoolVar(&healthJSON, "json", false, "Output health report in JSON format")
	healthCmd.Flags().BoolVar(&healthFix, "fix", false, "Show installation commands for missing tools")
	healthCmd.Flags().BoolVar(&healthInstallCmd, "install", false, "Show installation commands for all tools")
}

func runHealth(cmd *cobra.Command, args []string) error {
	// Define all tools that gman can use
	allTools := []*external.Tool{
		external.FD,
		external.RipGrep,
		external.FZF,
	}

	fmt.Println("🏥 Comprehensive System Health Diagnostics")
	fmt.Println("==========================================")
	fmt.Println()

	// Run system diagnostics
	diagnostics := external.RunSystemDiagnostics(allTools...)

	// Show summary
	readiness := diagnostics.GetReadiness()
	fmt.Printf("📊 %s System Readiness: %d%% (%d/%d tools available)\n", 
		getReadinessEmoji(readiness), readiness, 
		diagnostics.Summary.AvailableTools, diagnostics.Summary.TotalTools)
	fmt.Printf("🖥️  Platform: %s\n", diagnostics.Platform)
	fmt.Println()

	// Show available tools
	var available, missing []external.DiagnosticInfo
	for _, info := range diagnostics.Tools {
		if info.Available {
			available = append(available, info)
		} else {
			missing = append(missing, info)
		}
	}

	if len(available) > 0 {
		fmt.Printf("✅ %s Available tools (%d):\n", 
			color.GreenString("AVAILABLE"), len(available))
		
		for _, info := range available {
			if healthVerbose {
				if info.Error != nil {
					fmt.Printf("   • %s: %s (version check failed)\n", 
						info.Tool.Name, color.GreenString("INSTALLED"))
					fmt.Printf("     %s: %s\n", color.YellowString("Warning"), info.Error)
				} else {
					fmt.Printf("   • %s: %s (%s)\n", 
						info.Tool.Name, color.GreenString("INSTALLED"), info.Version)
				}
				fmt.Printf("     %s\n", color.CyanString(info.Tool.Description))
			} else {
				fmt.Printf("   • %s: %s\n", info.Tool.Name, color.GreenString("INSTALLED"))
			}
		}
		fmt.Println()
	}

	// Show missing tools with enhanced information
	if len(missing) > 0 {
		fmt.Printf("❌ %s Missing tools (%d):\n", 
			color.RedString("MISSING"), len(missing))
		
		for _, info := range missing {
			status := color.RedString("NOT FOUND")
			if !info.Tool.Required {
				status += color.YellowString(" (Optional)")
			}
			
			fmt.Printf("   • %s: %s\n", info.Tool.Name, status)
			if healthVerbose {
				fmt.Printf("     %s\n", color.CyanString(info.Tool.Description))
				fmt.Printf("     %s: %s\n", color.YellowString("Fallback"), info.Alternative)
			}
		}
		fmt.Println()

		// Show installation instructions if requested or if any tools are missing
		if healthFix || len(missing) > 0 {
			showEnhancedInstallationInstructions(missing)
		}
	}

	// Show system error health
	fmt.Println("🔍 System Error Analysis:")
	manager := errors.GetGlobalManager()
	report := manager.GenerateHealthReport()
	
	fmt.Printf("   Error Patterns: %d detected\n", len(report.ErrorPatterns))
	if len(report.ErrorPatterns) > 0 && healthVerbose {
		fmt.Println("   Recent Error Patterns:")
		for i, pattern := range report.ErrorPatterns {
			if i >= 3 { // Show top 3
				break
			}
			fmt.Printf("     %d. %s: %d occurrences\n", 
				i+1, pattern.ErrorType, pattern.Frequency)
		}
	}
	fmt.Println()

	// Show suggestions
	if len(diagnostics.Suggestions) > 0 {
		fmt.Println("💡 Recommendations:")
		for _, suggestion := range diagnostics.Suggestions {
			fmt.Printf("   %s\n", suggestion)
		}
		fmt.Println()
	}

	// Show installation commands for all tools if requested
	if healthInstallCmd {
		showAllInstallationInstructions(allTools)
	}

	// JSON output option
	if healthJSON {
		fmt.Printf(`{
  "system_readiness": %d,
  "tools_available": %d,
  "tools_total": %d,
  "error_patterns": %d,
  "overall_health": "%s"
}`, readiness, diagnostics.Summary.AvailableTools, 
			diagnostics.Summary.TotalTools, len(report.ErrorPatterns), report.OverallHealth)
		return nil
	}

	// Exit with appropriate code based on critical issues
	if diagnostics.HasCriticalIssues() {
		fmt.Printf("⚠️  %s: Critical tools are missing. Install them to ensure full functionality.\n", 
			color.RedString("CRITICAL"))
		os.Exit(2)
	} else if len(missing) > 0 {
		fmt.Printf("ℹ️  %s: Optional tools are missing. gman will work but with reduced functionality.\n", 
			color.YellowString("INFO"))
		os.Exit(1)
	}

	return nil
}

func getReadinessEmoji(readiness int) string {
	switch {
	case readiness >= 100:
		return "🟢"
	case readiness >= 80:
		return "🟡"
	case readiness >= 60:
		return "🟠"
	default:
		return "🔴"
	}
}

func showEnhancedInstallationInstructions(missing []external.DiagnosticInfo) {
	fmt.Println("📦 Installation instructions:")
	
	for _, info := range missing {
		priority := "Optional"
		if info.Tool.Required {
			priority = color.RedString("Required")
		} else {
			priority = color.YellowString("Optional")
		}
		
		fmt.Printf("   %s [%s]:\n", color.CyanString(info.Tool.Name), priority)
		fmt.Printf("     %s\n", info.Tool.Description)
		fmt.Printf("     Website: %s\n", color.BlueString(info.Tool.Website))
		
		instructions := info.Tool.GetInstallInstructions()
		lines := strings.Split(instructions, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				fmt.Printf("     %s\n", line)
			}
		}
		fmt.Println()
	}
}

func showAllInstallationInstructions(tools []*external.Tool) {
	fmt.Println("📚 Complete installation guide:")
	fmt.Println()
	
	for _, tool := range tools {
		status := color.GreenString("INSTALLED")
		if !tool.IsAvailable() {
			status = color.RedString("MISSING")
		}
		
		fmt.Printf("   %s [%s]:\n", color.CyanString(tool.Name), status)
		instructions := tool.GetInstallInstructions()
		
		// Indent the instructions
		lines := strings.Split(instructions, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				fmt.Printf("     %s\n", line)
			}
		}
		fmt.Println()
	}
}