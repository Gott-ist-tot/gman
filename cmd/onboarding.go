package cmd

import (
	"fmt"
	
	"github.com/spf13/cobra"
	"gman/internal/config"
	"gman/pkg/types"
)

// onboardingCmd represents commands for new user experience
var onboardingCmd = &cobra.Command{
	Use:   "onboarding",
	Short: "New user onboarding utilities",
	Long: `Onboarding utilities to help new users get started with gman.
	
These commands help with initial setup and learning gman basics.`,
	Hidden: true, // Hide from main help but accessible
}

// welcomeCmd shows welcome message for new users
var welcomeCmd = &cobra.Command{
	Use:   "welcome",
	Short: "Show welcome message and quick start guide",
	Long: `Display welcome message and essential commands for new gman users.

This is automatically shown on first run and provides a quick overview
of key gman features and commands.`,
	RunE: runWelcome,
}

// checkFirstRunCmd checks if this is the first run and offers setup
var checkFirstRunCmd = &cobra.Command{
	Use:   "check-first-run",
	Short: "Check if this is first run and offer setup",
	Long:  `Internal command to check first run status and offer setup wizard.`,
	RunE:  runCheckFirstRun,
	Hidden: true,
}

func init() {
	rootCmd.AddCommand(onboardingCmd)
	onboardingCmd.AddCommand(welcomeCmd)
	onboardingCmd.AddCommand(checkFirstRunCmd)
}

func runWelcome(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Welcome to gman - Git Repository Manager!")
	fmt.Println()
	fmt.Println("gman helps you manage multiple Git repositories efficiently.")
	fmt.Println()
	
	// Check if user has any repositories configured
	configMgr := config.NewManager()
	if err := configMgr.Load(); err == nil {
		cfg := configMgr.GetConfig()
		if len(cfg.Repositories) > 0 {
			fmt.Printf("✅ You have %d repositories configured.\n", len(cfg.Repositories))
			fmt.Println()
			showQuickCommands(cfg)
		} else {
			fmt.Println("❓ You don't have any repositories configured yet.")
			fmt.Println()
			showSetupInstructions()
		}
	} else {
		fmt.Println("❓ No configuration found.")
		fmt.Println()
		showSetupInstructions()
	}
	
	fmt.Println("📚 Learn More:")
	fmt.Println("  gman --help                  # Show all available commands")
	fmt.Println("  gman <command> --help        # Get help for specific commands")
	fmt.Println("  gman dashboard               # Launch interactive TUI")
	fmt.Println()
	fmt.Println("🔗 Documentation: https://github.com/your-org/gman")
	
	return nil
}

func runCheckFirstRun(cmd *cobra.Command, args []string) error {
	configMgr := config.NewManager()
	
	// Try to load existing config
	if err := configMgr.Load(); err != nil {
		// Config doesn't exist - this is likely a first run
		return offerSetup()
	}
	
	cfg := configMgr.GetConfig()
	if len(cfg.Repositories) == 0 {
		// Config exists but no repositories - offer setup
		return offerSetup()
	}
	
	// User has repositories configured, just show a brief welcome
	fmt.Println("Welcome back to gman! Use 'gman --help' for commands.")
	return nil
}

func offerSetup() error {
	fmt.Println("🚀 Welcome to gman!")
	fmt.Println()
	fmt.Println("It looks like this is your first time using gman.")
	fmt.Println("Would you like to run the setup wizard to get started? (Y/n)")
	
	var response string
	fmt.Scanln(&response)
	
	if response == "" || response == "y" || response == "Y" || response == "yes" {
		fmt.Println()
		fmt.Println("Starting setup wizard...")
		fmt.Println()
		
		// Run setup command
		setupCmd := &cobra.Command{}
		return runSetup(setupCmd, []string{})
	}
	
	fmt.Println()
	fmt.Println("Setup skipped. You can run 'gman setup' anytime to configure gman.")
	showManualSetupInstructions()
	
	return nil
}

func showSetupInstructions() {
	fmt.Println("🔧 Getting Started:")
	fmt.Println("  gman setup                   # Run interactive setup wizard")
	fmt.Println("  gman setup discover          # Auto-discover Git repositories")
	fmt.Println("  gman add <alias> <path>      # Manually add a repository")
	fmt.Println()
}

func showQuickCommands(cfg *types.Config) {
	fmt.Println("🚀 Quick Commands:")
	fmt.Println("  gman list                    # List all repositories")
	fmt.Println("  gman status                  # Show repository status")
	fmt.Println("  gman switch                  # Interactive repository switching")
	fmt.Println("  gman sync                    # Sync all repositories")
	
	// Show example with first repository
	if len(cfg.Repositories) > 0 {
		var firstAlias string
		for alias := range cfg.Repositories {
			firstAlias = alias
			break
		}
		fmt.Printf("  gman switch %s               # Switch to specific repository\n", firstAlias)
	}
	
	fmt.Println()
}

func showManualSetupInstructions() {
	fmt.Println("📝 Manual Setup:")
	fmt.Println("  gman add <alias> <path>      # Add a repository")
	fmt.Println("  gman list                    # Verify your repositories")
	fmt.Println("  gman status                  # Check repository status")
	fmt.Println()
	fmt.Println("💡 Tip: Use 'gman setup discover' to automatically find repositories.")
	fmt.Println()
}

// IsFirstRun checks if this appears to be a first run
func IsFirstRun() bool {
	configMgr := config.NewManager()
	
	// Try to load config
	if err := configMgr.Load(); err != nil {
		return true // Config doesn't exist
	}
	
	cfg := configMgr.GetConfig()
	return len(cfg.Repositories) == 0 // Config exists but empty
}

// ShowFirstRunWelcome shows welcome message on first run
func ShowFirstRunWelcome() error {
	return runWelcome(nil, nil)
}