package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"gman/internal/di"
	"gman/internal/tui"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Launch the interactive TUI dashboard",
	Long: `Launch the interactive Terminal User Interface (TUI) dashboard for gman.

The dashboard provides a unified interface for managing multiple Git repositories
with four main panels:

1. Repository Panel - Browse and select repositories
2. Status Panel - View detailed repository status
3. Search Panel - Search files and commits across repositories
4. Preview Panel - Preview files and commit content

Key Features:
• Real-time repository status monitoring
• Integrated search with fzf (Phase 5.1)
• Keyboard-driven navigation
• Repository grouping and filtering
• File and commit previews

Navigation:
• Tab/Shift+Tab: Navigate between panels
• 1-4: Jump directly to specific panels
• Arrow keys: Navigate within panels
• ?: Toggle help
• q: Quit

The dashboard integrates with all existing gman functionality while providing
a modern, intuitive interface for repository management.`,
	Aliases: []string{"dash", "tui", "ui"},
	RunE:    runDashboard,
}

var dashboardFlags struct {
	theme string
	force bool
	debug bool
}

func init() {
	// Command is now available via: gman tools dashboard
	// Removed direct rootCmd registration to avoid duplication

	// Add flags
	dashboardCmd.Flags().StringVar(&dashboardFlags.theme, "theme", "dark", "Color theme (dark, light)")
	dashboardCmd.Flags().BoolVar(&dashboardFlags.force, "force", false, "Force TUI mode even if terminal detection fails")
	dashboardCmd.Flags().BoolVar(&dashboardFlags.debug, "debug", false, "Show terminal detection debug information")
}

func runDashboard(cmd *cobra.Command, args []string) error {
	// Initialize configuration manager
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if repositories are configured
	repos := configMgr.GetConfig().Repositories
	if repos == nil {
		repos = make(map[string]string)
	}

	if len(repos) == 0 {
		fmt.Println("No repositories configured.")
		fmt.Println("Add repositories with: gman repo add <alias> <path>")
		fmt.Println("Or run: gman repo list to see available commands")
		return nil
	}

	// Validate theme
	if dashboardFlags.theme != "dark" && dashboardFlags.theme != "light" {
		return fmt.Errorf("invalid theme: %s (available: dark, light)", dashboardFlags.theme)
	}

	// Set up terminal for TUI (only check if not forced)
	if !dashboardFlags.force {
		if err := setupTerminal(dashboardFlags.force, dashboardFlags.debug); err != nil {
			return fmt.Errorf("failed to setup terminal: %w", err)
		}
	} else if dashboardFlags.debug {
		fmt.Println("⚠️  Force flag enabled - bypassing all terminal checks")
	}
	defer restoreTerminal()

	// Run the TUI dashboard
	if err := tui.Run(configMgr); err != nil {
		return fmt.Errorf("dashboard error: %w", err)
	}

	return nil
}

// TerminalDiagnostics contains terminal detection information
type TerminalDiagnostics struct {
	StdoutIsTTY    bool
	TermEnv        string
	TermSupported  bool
	TTYPath        string
	CanOpenTTY     bool
	DetectedIssues []string
	Suggestions    []string
}

// setupTerminal prepares the terminal for TUI mode
func setupTerminal(force, debug bool) error {
	diag := checkTerminalCapabilities()

	if debug {
		printTerminalDiagnostics(diag)
	}

	// If force flag is used, skip checks
	if force {
		if debug {
			fmt.Println("⚠️  Force flag enabled - bypassing terminal checks")
		}
		return nil
	}

	// Enhanced terminal suitability check
	if !isTerminalSuitable(diag) {
		return buildTerminalError(diag)
	}

	// Additional terminal setup could go here
	// For now, Bubble Tea handles most of the setup

	return nil
}

// restoreTerminal restores the terminal after TUI mode
func restoreTerminal() {
	// Bubble Tea handles most of the cleanup
	// Additional cleanup could go here if needed
}

// checkTerminalCapabilities performs comprehensive terminal detection
func checkTerminalCapabilities() TerminalDiagnostics {
	diag := TerminalDiagnostics{
		DetectedIssues: make([]string, 0),
		Suggestions:    make([]string, 0),
	}

	// Check if stdout is a TTY
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		diag.StdoutIsTTY = true
	} else {
		diag.DetectedIssues = append(diag.DetectedIssues, "stdout is not connected to a TTY")
		diag.Suggestions = append(diag.Suggestions, "try running in a real terminal or use --force flag")
	}

	// Check TERM environment variable
	diag.TermEnv = os.Getenv("TERM")
	if diag.TermEnv == "" {
		diag.DetectedIssues = append(diag.DetectedIssues, "TERM environment variable is not set")
		diag.Suggestions = append(diag.Suggestions, "set TERM environment variable (e.g., export TERM=xterm-256color)")
	} else if diag.TermEnv == "dumb" {
		diag.DetectedIssues = append(diag.DetectedIssues, "TERM is set to 'dumb' which doesn't support TUI")
		diag.Suggestions = append(diag.Suggestions, "use a terminal that supports cursor movement and colors")
	} else {
		diag.TermSupported = true
	}

	// Try to open /dev/tty as fallback
	if !diag.StdoutIsTTY {
		if tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
			diag.CanOpenTTY = true
			diag.TTYPath = "/dev/tty"
			tty.Close()
			diag.Suggestions = append(diag.Suggestions, "terminal device found at /dev/tty - TUI might work with --force")
		} else {
			diag.DetectedIssues = append(diag.DetectedIssues, "cannot access terminal device (/dev/tty)")
		}
	}

	return diag
}

// isTerminalSuitable checks if the terminal supports TUI mode
func isTerminalSuitable(diag TerminalDiagnostics) bool {
	// Primary check: stdout is TTY and TERM is set properly
	if diag.StdoutIsTTY && diag.TermSupported {
		return true
	}

	// Fallback: if we can access TTY device and TERM is supported
	if diag.CanOpenTTY && diag.TermSupported {
		return true
	}

	// Enhanced fallback: if TERM is supported, allow even without TTY access
	// This handles cases like containers, CI environments, or certain SSH sessions
	if diag.TermSupported && diag.TermEnv != "dumb" {
		return true
	}

	// Minimal fallback: if we have any reasonable terminal environment
	// Allow TUI even with basic terminal detection
	if diag.TermEnv != "" && diag.TermEnv != "dumb" {
		return true
	}

	return false
}

// buildTerminalError creates a detailed error message with suggestions
func buildTerminalError(diag TerminalDiagnostics) error {
	var msg strings.Builder

	msg.WriteString("Terminal does not support TUI mode\n\n")

	if len(diag.DetectedIssues) > 0 {
		msg.WriteString("Detected issues:\n")
		for _, issue := range diag.DetectedIssues {
			msg.WriteString(fmt.Sprintf("  • %s\n", issue))
		}
		msg.WriteString("\n")
	}

	if len(diag.Suggestions) > 0 {
		msg.WriteString("Environment-specific solutions:\n")
		for _, suggestion := range diag.Suggestions {
			msg.WriteString(fmt.Sprintf("  • %s\n", suggestion))
		}
		msg.WriteString("\n")
	}

	// Add common environment-specific guidance
	msg.WriteString("Common solutions by environment:\n")
	msg.WriteString("  • SSH: Use 'ssh -t user@host' to allocate a proper TTY\n")
	msg.WriteString("  • VS Code: Open a real terminal instead of the integrated terminal\n")
	msg.WriteString("  • tmux/screen: Make sure session has proper TTY allocation\n")
	msg.WriteString("  • Docker: Run with 'docker run -it' to enable interactive mode\n")
	msg.WriteString("  • CI/CD: TUI mode is not recommended in automated environments\n\n")

	msg.WriteString("Quick options:\n")
	msg.WriteString("  • Use --force to attempt TUI mode anyway\n")
	msg.WriteString("  • Use --debug to see detailed diagnostic information\n")
	msg.WriteString("  • Continue using CLI commands: gman repo list, gman work status, etc.\n")

	return fmt.Errorf("%s", msg.String())
}

// printTerminalDiagnostics prints detailed diagnostic information
func printTerminalDiagnostics(diag TerminalDiagnostics) {
	fmt.Println("🔍 Terminal Diagnostics:")
	fmt.Printf("  Stdout is TTY: %v\n", diag.StdoutIsTTY)
	fmt.Printf("  TERM environment: %q\n", diag.TermEnv)
	fmt.Printf("  TERM supported: %v\n", diag.TermSupported)
	if diag.TTYPath != "" {
		fmt.Printf("  TTY device: %s (accessible: %v)\n", diag.TTYPath, diag.CanOpenTTY)
	}

	if len(diag.DetectedIssues) > 0 {
		fmt.Println("  Issues found:")
		for _, issue := range diag.DetectedIssues {
			fmt.Printf("    • %s\n", issue)
		}
	}

	if len(diag.Suggestions) > 0 {
		fmt.Println("  Suggestions:")
		for _, suggestion := range diag.Suggestions {
			fmt.Printf("    • %s\n", suggestion)
		}
	}
	fmt.Println()
}
