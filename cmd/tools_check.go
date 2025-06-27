package cmd

import (
	"fmt"
	"os"
	"strings"

	"gman/internal/external"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var toolsCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check availability of external tools and dependencies",
	Long: `Check the availability of external tools required by gman and provide 
installation instructions for missing tools. This command helps diagnose
tool availability issues and provides platform-specific installation guidance.`,
	Run: runToolsCheck,
}

var (
	checkVerbose    bool
	checkFix        bool
	checkInstallCmd bool
)

func init() {
	toolsCmd.AddCommand(toolsCheckCmd)
	
	toolsCheckCmd.Flags().BoolVarP(&checkVerbose, "verbose", "v", false, "Show detailed version information")
	toolsCheckCmd.Flags().BoolVar(&checkFix, "fix", false, "Show installation commands for missing tools")
	toolsCheckCmd.Flags().BoolVar(&checkInstallCmd, "install", false, "Show installation commands for all tools")
}

func runToolsCheck(cmd *cobra.Command, args []string) {
	// Define all tools that gman can use
	allTools := []*external.Tool{
		external.FD,
		external.RipGrep,
		external.FZF,
	}

	fmt.Println("🔍 Running comprehensive system diagnostics...")
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
			if checkVerbose {
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
			if checkVerbose {
				fmt.Printf("     %s\n", color.CyanString(info.Tool.Description))
				fmt.Printf("     %s: %s\n", color.YellowString("Fallback"), info.Alternative)
			}
		}
		fmt.Println()

		// Show installation instructions if requested or if any tools are missing
		if checkFix || len(missing) > 0 {
			showEnhancedInstallationInstructions(missing)
		}
	}

	// Show suggestions
	if len(diagnostics.Suggestions) > 0 {
		fmt.Println("💡 Recommendations:")
		for _, suggestion := range diagnostics.Suggestions {
			fmt.Printf("   %s\n", suggestion)
		}
		fmt.Println()
	}

	// Show installation commands for all tools if requested
	if checkInstallCmd {
		showAllInstallationInstructions(allTools)
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

func showMissingToolsImpact(missing []*external.Tool) {
	fmt.Println("🔧 Impact of missing tools:")
	
	for _, tool := range missing {
		switch tool.Name {
		case "fd":
			fmt.Printf("   • Without %s: File search will be slower and less efficient\n", 
				color.YellowString("fd"))
			fmt.Printf("     - %s will fall back to basic file listing\n", 
				color.CyanString("gman tools find file"))
		
		case "ripgrep":
			fmt.Printf("   • Without %s: Content search will not be available\n", 
				color.YellowString("rg"))
			fmt.Printf("     - %s command will be disabled\n", 
				color.CyanString("gman tools find content"))
		
		case "fzf":
			fmt.Printf("   • Without %s: Interactive selection will be less user-friendly\n", 
				color.YellowString("fzf"))
			fmt.Printf("     - Search results will use basic numbered selection\n")
		}
	}
	fmt.Println()
}

func showInstallationInstructions(missing []*external.Tool) {
	fmt.Println("📦 Installation instructions:")
	
	for _, tool := range missing {
		fmt.Printf("   %s:\n", color.CyanString(tool.Name))
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