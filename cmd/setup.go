package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gman/internal/di"

	"github.com/spf13/cobra"
)

// setupCmd represents the setup command for new user onboarding
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup wizard for new users",
	Long: `Setup wizard guides new users through initial gman configuration.

This command helps you:
- Discover existing Git repositories on your system
- Configure repository aliases for easy access
- Set up default preferences
- Learn basic gman commands

The wizard will create your configuration file and get you started quickly.`,
	RunE: runSetup,
}

// setupDiscoverCmd discovers Git repositories automatically
var setupDiscoverCmd = &cobra.Command{
	Use:   "discover [path]",
	Short: "Automatically discover Git repositories",
	Long: `Automatically discover Git repositories in the specified path (or current directory).

This command will:
- Scan the directory tree for Git repositories
- Suggest repository aliases based on directory names
- Allow you to select which repositories to add to gman

Examples:
  gman setup discover                    # Discover in current directory
  gman setup discover ~/Projects        # Discover in specific directory
  gman setup discover --depth 2 ~/Code  # Limit search depth`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSetupDiscover,
}

var (
	discoverDepth   int
	discoverConfirm bool
)

func init() {
	// Command is now available via: gman tools setup
	// Removed direct rootCmd registration to avoid duplication
	setupCmd.AddCommand(setupDiscoverCmd)

	// Add flags for discover command
	setupDiscoverCmd.Flags().IntVar(&discoverDepth, "depth", 3, "Maximum directory depth to search")
	setupDiscoverCmd.Flags().BoolVar(&discoverConfirm, "auto-confirm", false, "Automatically confirm all discovered repositories")
}

func runSetup(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Welcome to gman - Git Repository Manager!")
	fmt.Println()

	// Check if config already exists
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err == nil {
		cfg := configMgr.GetConfig()
		if len(cfg.Repositories) > 0 {
			fmt.Printf("⚠️  You already have %d repositories configured.\n", len(cfg.Repositories))
			fmt.Println("Do you want to continue with setup anyway? (y/N)")

			if !askConfirmation(false) {
				fmt.Println("Setup cancelled. Use 'gman list' to see your current repositories.")
				return nil
			}
		}
	}

	fmt.Println("This setup wizard will help you get started with gman.")
	fmt.Println()

	// Step 1: Repository Discovery
	if err := runSetupStep1(); err != nil {
		return fmt.Errorf("setup step 1 failed: %w", err)
	}

	// Step 2: Basic Configuration
	if err := runSetupStep2(); err != nil {
		return fmt.Errorf("setup step 2 failed: %w", err)
	}

	// Step 3: Quick Tutorial
	if err := runSetupStep3(); err != nil {
		return fmt.Errorf("setup step 3 failed: %w", err)
	}

	fmt.Println("🎉 Setup complete! You're ready to use gman.")
	fmt.Println()
	fmt.Println("Try these commands to get started:")
	fmt.Println("  gman list              # List all repositories")
	fmt.Println("  gman status            # Show repository status")
	fmt.Println("  gman switch            # Interactive repository switching")
	fmt.Println("  gman dashboard         # Launch interactive TUI")

	return nil
}

func runSetupStep1() error {
	fmt.Println("📁 Step 1: Repository Discovery")
	fmt.Println("Let's find your Git repositories...")
	fmt.Println()

	// Ask for discovery path
	fmt.Print("Enter path to search for repositories (default: current directory): ")
	reader := bufio.NewReader(os.Stdin)
	pathInput, _ := reader.ReadString('\n')
	pathInput = strings.TrimSpace(pathInput)

	if pathInput == "" {
		wd, _ := os.Getwd()
		pathInput = wd
	}

	// Expand home directory
	if strings.HasPrefix(pathInput, "~/") {
		home, _ := os.UserHomeDir()
		pathInput = filepath.Join(home, pathInput[2:])
	}

	fmt.Printf("Searching for Git repositories in: %s\n", pathInput)
	fmt.Printf("Search depth: %d levels\n", 3)
	fmt.Println()

	// Discover repositories
	repos, err := discoverRepositories(pathInput, 3)
	if err != nil {
		return fmt.Errorf("failed to discover repositories: %w", err)
	}

	if len(repos) == 0 {
		fmt.Println("No Git repositories found.")
		fmt.Println("You can add repositories manually later using 'gman add <alias> <path>'")
		return nil
	}

	fmt.Printf("Found %d Git repositories:\n", len(repos))
	fmt.Println()

	// Show discovered repositories and let user select
	selectedRepos := selectRepositories(repos)

	if len(selectedRepos) == 0 {
		fmt.Println("No repositories selected.")
		return nil
	}

	// Add selected repositories to config
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		// Create new config if doesn't exist
		if err := configMgr.CreateDefaultConfig(); err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}
	}

	cfg := configMgr.GetConfig()
	for alias, path := range selectedRepos {
		cfg.Repositories[alias] = path
	}

	if err := configMgr.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✅ Added %d repositories to gman configuration.\n", len(selectedRepos))
	fmt.Println()

	return nil
}

func runSetupStep2() error {
	fmt.Println("⚙️  Step 2: Basic Configuration")
	fmt.Println("Let's configure some preferences...")
	fmt.Println()

	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg := configMgr.GetConfig()

	// Configure sync mode
	fmt.Println("Default sync mode determines how 'gman sync' operates:")
	fmt.Println("  1. ff-only    - Fast-forward only (safe, recommended)")
	fmt.Println("  2. rebase     - Rebase local changes")
	fmt.Println("  3. autostash  - Auto-stash and rebase")
	fmt.Println()
	fmt.Print("Choose default sync mode (1-3, default: 1): ")

	choice := askChoice([]string{"ff-only", "rebase", "autostash"}, 1)
	cfg.Settings.DefaultSyncMode = choice

	// Configure parallel jobs
	fmt.Println()
	fmt.Printf("Parallel jobs for concurrent operations (current: %d): ", cfg.Settings.ParallelJobs)
	if cfg.Settings.ParallelJobs == 0 {
		cfg.Settings.ParallelJobs = 5 // default
		fmt.Printf("%d", cfg.Settings.ParallelJobs)
	}
	fmt.Println()

	// Configure display preferences
	fmt.Print("Show last commit in repository list by default? (Y/n): ")
	cfg.Settings.ShowLastCommit = askConfirmation(true)

	if err := configMgr.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("✅ Configuration saved.")
	fmt.Println()

	return nil
}

func runSetupStep3() error {
	fmt.Println("📚 Step 3: Quick Tutorial")
	fmt.Println("Here are the key commands you'll use with gman:")
	fmt.Println()

	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return err
	}

	cfg := configMgr.GetConfig()

	// Show basic commands with examples
	if len(cfg.Repositories) > 0 {
		// Pick first repo as example
		var exampleAlias string
		for alias := range cfg.Repositories {
			exampleAlias = alias
			break
		}

		fmt.Printf("📋 Repository Management:\n")
		fmt.Printf("  gman list                    # List all repositories\n")
		fmt.Printf("  gman add myrepo /path/to/repo # Add new repository\n")
		fmt.Printf("  gman remove %s               # Remove repository\n", exampleAlias)
		fmt.Println()

		fmt.Printf("🔄 Repository Operations:\n")
		fmt.Printf("  gman status                  # Show status of all repositories\n")
		fmt.Printf("  gman sync                    # Sync all repositories\n")
		fmt.Printf("  gman sync %s                 # Sync specific repository\n", exampleAlias)
		fmt.Println()

		fmt.Printf("🧭 Navigation:\n")
		fmt.Printf("  gman switch                  # Interactive repository selection\n")
		fmt.Printf("  gman switch %s               # Switch to specific repository\n", exampleAlias)
		fmt.Printf("  gman recent                  # Show recently accessed repositories\n")
		fmt.Println()

		fmt.Printf("🎯 Advanced Features:\n")
		fmt.Printf("  gman dashboard               # Launch interactive TUI\n")
		fmt.Printf("  gman group create mygroup %s # Create repository group\n", exampleAlias)
		fmt.Printf("  gman branch list             # Cross-repository branch management\n")
		fmt.Println()
	}

	fmt.Println("💡 Pro Tips:")
	fmt.Println("  - Use 'gman <command> --help' for detailed help on any command")
	fmt.Println("  - The shell wrapper enables directory switching with 'gman switch'")
	fmt.Println("  - Use the interactive dashboard for a visual interface")
	fmt.Println()

	fmt.Print("Would you like to try the interactive dashboard now? (y/N): ")
	if askConfirmation(false) {
		fmt.Println("Launching dashboard...")
		// Note: We can't actually launch it here due to command structure,
		// but we'll show the instruction
		fmt.Println("Run: gman dashboard")
	}

	return nil
}

func runSetupDiscover(cmd *cobra.Command, args []string) error {
	var searchPath string
	if len(args) > 0 {
		searchPath = args[0]
	} else {
		wd, _ := os.Getwd()
		searchPath = wd
	}

	// Expand home directory
	if strings.HasPrefix(searchPath, "~/") {
		home, _ := os.UserHomeDir()
		searchPath = filepath.Join(home, searchPath[2:])
	}

	fmt.Printf("🔍 Discovering Git repositories in: %s\n", searchPath)
	fmt.Printf("Search depth: %d levels\n", discoverDepth)
	fmt.Println()

	repos, err := discoverRepositories(searchPath, discoverDepth)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}

	if len(repos) == 0 {
		fmt.Println("No Git repositories found.")
		return nil
	}

	fmt.Printf("Found %d Git repositories:\n", len(repos))
	for alias, path := range repos {
		fmt.Printf("  %s → %s\n", alias, path)
	}
	fmt.Println()

	var selectedRepos map[string]string

	if discoverConfirm {
		selectedRepos = repos
		fmt.Printf("Auto-confirming all %d repositories.\n", len(repos))
	} else {
		selectedRepos = selectRepositories(repos)
	}

	if len(selectedRepos) == 0 {
		fmt.Println("No repositories selected.")
		return nil
	}

	// Load and update config
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		if err := configMgr.CreateDefaultConfig(); err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}
	}

	cfg := configMgr.GetConfig()
	added := 0

	for alias, path := range selectedRepos {
		if _, exists := cfg.Repositories[alias]; !exists {
			cfg.Repositories[alias] = path
			added++
		} else {
			fmt.Printf("⚠️  Repository '%s' already exists, skipping.\n", alias)
		}
	}

	if added > 0 {
		if err := configMgr.Save(); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}
		fmt.Printf("✅ Added %d new repositories to gman.\n", added)
	} else {
		fmt.Println("No new repositories were added.")
	}

	return nil
}

// discoverRepositories searches for Git repositories in the given path
func discoverRepositories(rootPath string, maxDepth int) (map[string]string, error) {
	repos := make(map[string]string)
	gitMgr := di.GitManager()

	// Convert to absolute path
	absRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// First check if the root path itself is a git repository
	if gitMgr.IsGitRepository(absRootPath) {
		alias := generateRepoAlias(absRootPath)
		repos[alias] = absRootPath
	}

	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		// Check depth limit
		relPath, _ := filepath.Rel(rootPath, path)
		depth := strings.Count(relPath, string(filepath.Separator))

		// For .git directories, we need to check their parent repository
		if info.IsDir() && info.Name() == ".git" {
			repoPath := filepath.Dir(path)
			repoRelPath, _ := filepath.Rel(rootPath, repoPath)
			repoDepth := strings.Count(repoRelPath, string(filepath.Separator))
			if repoDepth > maxDepth {
				return filepath.SkipDir
			}
		} else if depth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden directories (except .git itself)
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != ".git" {
			return filepath.SkipDir
		}

		// Check if this is a Git repository
		if info.IsDir() && info.Name() == ".git" {
			repoPath := filepath.Dir(path)
			if gitMgr.IsGitRepository(repoPath) {
				alias := generateRepoAlias(repoPath)
				// Only add if we haven't already added it (avoid duplicating root)
				if _, exists := repos[alias]; !exists {
					repos[alias] = repoPath
				}
			}
			return filepath.SkipDir // Don't recurse into .git
		}

		return nil
	})

	return repos, err
}

// generateRepoAlias creates a repository alias from its path
func generateRepoAlias(repoPath string) string {
	base := filepath.Base(repoPath)

	// Clean up common prefixes/suffixes
	base = strings.TrimSuffix(base, ".git")

	// Handle special cases
	if base == "." || base == "" {
		parent := filepath.Base(filepath.Dir(repoPath))
		if parent != "." && parent != "" {
			base = parent
		} else {
			base = "repository"
		}
	}

	return base
}

// selectRepositories allows user to select which repositories to add
func selectRepositories(repos map[string]string) map[string]string {
	selected := make(map[string]string)

	fmt.Println("Select repositories to add to gman:")
	fmt.Println("(Press Enter to confirm all, or specify numbers separated by spaces)")
	fmt.Println()

	// Create indexed list
	var repoList []struct {
		alias string
		path  string
	}

	i := 1
	for alias, path := range repos {
		fmt.Printf("[%d] %s → %s\n", i, alias, path)
		repoList = append(repoList, struct {
			alias string
			path  string
		}{alias, path})
		i++
	}

	fmt.Println()
	fmt.Print("Selection (default: all): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		// Select all
		return repos
	}

	// Parse selection
	parts := strings.Fields(input)
	for _, part := range parts {
		if part == "all" {
			return repos
		}

		var idx int
		if _, err := fmt.Sscanf(part, "%d", &idx); err == nil {
			if idx >= 1 && idx <= len(repoList) {
				repo := repoList[idx-1]
				selected[repo.alias] = repo.path
			}
		}
	}

	return selected
}

// askConfirmation asks for user confirmation
func askConfirmation(defaultYes bool) bool {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return defaultYes
	}

	return input == "y" || input == "yes"
}

// askChoice asks user to choose from a list of options
func askChoice(options []string, defaultChoice int) string {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return options[defaultChoice-1]
	}

	var choice int
	if _, err := fmt.Sscanf(input, "%d", &choice); err == nil {
		if choice >= 1 && choice <= len(options) {
			return options[choice-1]
		}
	}

	return options[defaultChoice-1]
}
