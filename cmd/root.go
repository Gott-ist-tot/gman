package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gman",
	Short: "Smart multi-repository Git assistant and navigator",
	Long: `gman is your intelligent assistant for navigating and monitoring multiple Git repositories.
🧭 Designed as a "smart副駕駛" that provides powerful observability and safe navigation tools.

🚀 QUICK START:
  gman tools setup                    # Interactive setup wizard for new users
  gman work status                    # Monitor status across all repositories
  gman switch                         # Smart repository navigation with recent history
  gman tools dashboard                # Launch interactive TUI dashboard

📁 COMMAND GROUPS (with shortcuts):
  gman repo    (r)                    # Repository management (add, remove, list, groups)
  gman work    (w)                    # Safe Git workflows (status monitoring, sync)
  gman tools   (t)                    # Advanced tools (search, dashboard, health, setup)

🔍 INTELLIGENT NAVIGATION:
  gman switch                         # Smart switching with recent history priority
  gman switch --recent               # Show only recently accessed repositories
  gman work status --verbose        # Detailed multi-repository status overview

🛠️ SMART DISCOVERY & SEARCH:
  gman tools find file "config.yaml" # Lightning-fast file search across repositories
  gman tools find content "TODO"     # Powerful content search with ripgrep
  gman tools find commit "fix bug"   # Enhanced commit search across repositories
  gman tools setup discover ~/Projects # Auto-discover and configure Git repositories

📋 TASK MANAGEMENT & EXTERNAL TOOL INTEGRATION:
  gman tools task create feature-auth  # Create task-oriented file collections
  gman tools task add auth src/*.go    # Add files to task collections
  gman tools task list-files auth | xargs aider  # External tool integration (aider, linters, etc.)

🏥 SYSTEM HEALTH & DIAGNOSTICS:
  gman tools health                   # Comprehensive system health check
  gman tools health --fix            # Show installation commands for missing tools

🎯 PHILOSOPHY: gman focuses on safe, read-focused operations and intelligent navigation,
leaving precise write operations to you and native Git commands for maximum safety.

💡 TIP: Use 'gman <group> --help' to see all commands in each group.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/gman/config.yml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".config/gman" (without extension).
		configDir := home + "/.config/gman"
		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		// Silently handle config file not found - it will be created when needed
	}
}
