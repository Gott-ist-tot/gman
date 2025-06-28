package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"gman/internal/di"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	initForce    bool
	initShellAll bool
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize gman for optimal usage",
	Long: `Initialize gman with automated setup for shell integration and dependencies.
	
This command helps set up gman for the best user experience by:
- Installing shell integration for directory switching
- Setting up command completion
- Detecting available dependencies and providing installation guidance
- Configuring optimal settings for your environment

Examples:
  gman init shell                    # Set up shell integration only
  gman init shell --all              # Set up for all supported shells
  gman init shell --force            # Overwrite existing shell integration`,
}

// initShellCmd represents the init shell command
var initShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Set up shell integration for directory switching",
	Long: `Automatically set up shell integration for gman directory switching.

This command will:
1. Detect your current shell (bash, zsh, fish)
2. Add the gman wrapper function to your shell configuration
3. Enable command completion
4. Provide guidance on dependency installation

The shell integration enables:
- 'gman switch' to actually change directories
- Tab completion for all gman commands
- Optional dependency warnings for enhanced features

Supported shells: bash, zsh, fish`,
	RunE: runInitShell,
}

func init() {
	// Add init command to tools
	initCmd.AddCommand(initShellCmd)
	
	// Shell-specific flags
	initShellCmd.Flags().BoolVar(&initForce, "force", false, "Overwrite existing shell integration")
	initShellCmd.Flags().BoolVar(&initShellAll, "all", false, "Set up integration for all supported shells")
}

func runInitShell(cmd *cobra.Command, args []string) error {
	fmt.Printf("%s\n", color.BlueString("🔧 Setting up gman shell integration..."))
	
	// Detect current shell
	shell := detectShell()
	if shell == "" {
		return fmt.Errorf("unable to detect shell type. Please set up manually using scripts/shell-integration.sh")
	}
	
	fmt.Printf("Detected shell: %s\n", color.CyanString(shell))
	
	// Get shell config file
	configFile, err := getShellConfigFile(shell)
	if err != nil {
		return fmt.Errorf("failed to determine shell config file: %w", err)
	}
	
	fmt.Printf("Shell config file: %s\n", color.YellowString(configFile))
	
	// Check if integration already exists
	if !initForce {
		hasIntegration, err := checkExistingIntegration(configFile)
		if err != nil {
			fmt.Printf("Warning: Could not check existing integration: %v\n", err)
		} else if hasIntegration {
			fmt.Printf("%s\n", color.YellowString("⚠️  Shell integration already exists."))
			fmt.Println("Use --force to overwrite, or remove manually and try again.")
			return nil
		}
	}
	
	// Add shell integration
	err = addShellIntegration(configFile, shell)
	if err != nil {
		return fmt.Errorf("failed to add shell integration: %w", err)
	}
	
	fmt.Printf("%s Shell integration added successfully!\n", color.GreenString("✅"))
	fmt.Println()
	fmt.Printf("To activate the integration, either:\n")
	fmt.Printf("1. Restart your terminal\n")
	fmt.Printf("2. Run: %s\n", color.CyanString(fmt.Sprintf("source %s", configFile)))
	fmt.Println()
	
	// Check and warn about dependencies
	checkDependencies()
	
	// Show usage examples
	showUsageExamples()
	
	return nil
}

func detectShell() string {
	// Check SHELL environment variable
	shell := os.Getenv("SHELL")
	if shell != "" {
		if strings.Contains(shell, "zsh") {
			return "zsh"
		} else if strings.Contains(shell, "bash") {
			return "bash"
		} else if strings.Contains(shell, "fish") {
			return "fish"
		}
	}
	
	// Fallback: check if we're in a known shell environment
	if os.Getenv("ZSH_VERSION") != "" {
		return "zsh"
	}
	if os.Getenv("BASH_VERSION") != "" {
		return "bash"
	}
	
	return ""
}

func getShellConfigFile(shell string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	
	homeDir := usr.HomeDir
	
	switch shell {
	case "zsh":
		// Check for .zshrc first, fall back to .zprofile
		zshrc := filepath.Join(homeDir, ".zshrc")
		if _, err := os.Stat(zshrc); err == nil {
			return zshrc, nil
		}
		return filepath.Join(homeDir, ".zprofile"), nil
		
	case "bash":
		// Check for .bashrc first, fall back to .bash_profile
		bashrc := filepath.Join(homeDir, ".bashrc")
		if _, err := os.Stat(bashrc); err == nil {
			return bashrc, nil
		}
		return filepath.Join(homeDir, ".bash_profile"), nil
		
	case "fish":
		configDir := filepath.Join(homeDir, ".config", "fish")
		return filepath.Join(configDir, "config.fish"), nil
		
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
}

func checkExistingIntegration(configFile string) (bool, error) {
	file, err := os.Open(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // File doesn't exist, no integration
		}
		return false, err
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "gman()") || strings.Contains(line, "gman shell integration") {
			return true, nil
		}
	}
	
	return false, scanner.Err()
}

func addShellIntegration(configFile, shell string) error {
	// Ensure the directory exists
	dir := filepath.Dir(configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// Open file for appending
	file, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Add integration based on shell type
	integration := getShellIntegrationCode(shell)
	
	_, err = file.WriteString(integration)
	return err
}

func getShellIntegrationCode(shell string) string {
	switch shell {
	case "fish":
		return `
# gman shell integration - added by 'gman init shell'
# Set environment variable to indicate shell integration is active
set -gx GMAN_SHELL_INTEGRATION 1

function gman
    if test "$argv[1]" = "switch" -o "$argv[1]" = "sw" -o "$argv[1]" = "cd"
        set output (command gman $argv 2>&1)
        set exit_code $status
        
        set gman_cd_line (echo "$output" | grep "^GMAN_CD:")
        
        if test -n "$gman_cd_line"
            set target_dir (echo "$gman_cd_line" | string replace "GMAN_CD:" "")
            echo "$output" | grep -v "^GMAN_CD:" | grep -v "^\$"
            
            if test -d "$target_dir"
                cd "$target_dir"
                echo "Switched to: $target_dir"
            else
                echo "Error: Directory not found: $target_dir" >&2
                return 1
            end
        else
            echo "$output"
        end
        return $exit_code
    else
        command gman $argv
    end
end

# Enable gman completion for fish
if command -v gman > /dev/null
    gman completion fish | source
end
`
	default: // bash/zsh
		return `
# gman shell integration - added by 'gman init shell'
# Set environment variable to indicate shell integration is active
export GMAN_SHELL_INTEGRATION=1

gman() {
    # Check if the first argument is 'switch' or its aliases
    if [[ "$1" == "switch" || "$1" == "sw" || "$1" == "cd" ]]; then
        local output gman_cd_line
        # For switch commands, capture output to handle GMAN_CD
        output=$(command gman "$@" 2>&1)
        local exit_code=$?
        
        # Extract GMAN_CD line while preserving other output
        gman_cd_line=$(echo "$output" | grep "^GMAN_CD:")
        
        if [[ -n "$gman_cd_line" ]]; then
            local target_dir="${gman_cd_line#GMAN_CD:}"
            # Print all non-GMAN_CD output first
            echo "$output" | grep -v "^GMAN_CD:" | grep -v "^$"
            
            if [ -d "$target_dir" ]; then
                cd "$target_dir"
                echo "Switched to: $target_dir"
            else
                echo "Error: Directory not found: $target_dir" >&2
                return 1
            fi
        else
            echo "$output"
        fi
        return $exit_code
    else
        # For all other commands, execute directly
        command gman "$@"
    fi
}

# Enable gman completion
if command -v gman &> /dev/null; then
    if [ -n "$BASH_VERSION" ]; then
        eval "$(gman completion bash)"
    elif [ -n "$ZSH_VERSION" ]; then
        eval "$(gman completion zsh)"
    fi
fi
`
	}
}

func checkDependencies() {
	fmt.Printf("%s\n", color.BlueString("🔍 Checking optional dependencies..."))
	
	dependencies := []struct {
		name        string
		command     string
		description string
		installCmd  string
	}{
		{"fd", "fd", "Lightning-fast file search", "brew install fd (macOS) or apt install fd-find (Ubuntu)"},
		{"ripgrep", "rg", "Fast content search with regex", "brew install ripgrep (macOS) or apt install ripgrep (Ubuntu)"},
		{"fzf", "fzf", "Interactive fuzzy finder", "brew install fzf (macOS) or apt install fzf (Ubuntu)"},
	}
	
	var missing []string
	var available []string
	
	for _, dep := range dependencies {
		if commandExists(dep.command) {
			available = append(available, dep.name)
		} else {
			missing = append(missing, dep.name)
		}
	}
	
	if len(available) > 0 {
		fmt.Printf("%s Available: %s\n", color.GreenString("✅"), strings.Join(available, ", "))
	}
	
	if len(missing) > 0 {
		fmt.Printf("%s Missing: %s\n", color.YellowString("⚠️"), strings.Join(missing, ", "))
		fmt.Println()
		fmt.Printf("These tools are optional but greatly enhance gman's search capabilities:\n")
		for _, dep := range dependencies {
			for _, m := range missing {
				if dep.name == m {
					fmt.Printf("  • %s: %s\n", color.CyanString(dep.name), dep.description)
					fmt.Printf("    Install: %s\n", dep.installCmd)
				}
			}
		}
		fmt.Println()
		fmt.Printf("You can install all dependencies at once with:\n")
		fmt.Printf("  %s\n", color.CyanString("./scripts/setup-dependencies.sh"))
	} else {
		fmt.Printf("%s All optional dependencies are available!\n", color.GreenString("🚀"))
	}
}

func showUsageExamples() {
	fmt.Printf("%s\n", color.BlueString("📚 Quick start examples:"))
	fmt.Println()
	
	// Check if repositories are configured
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err == nil {
		cfg := configMgr.GetConfig()
		if len(cfg.Repositories) > 0 {
			fmt.Printf("Try these commands with your configured repositories:\n")
		} else {
			fmt.Printf("First, add some repositories:\n")
			fmt.Printf("  %s\n", color.CyanString("gman repo add myproject /path/to/project"))
			fmt.Printf("  %s\n", color.CyanString("gman setup discover ~/Projects"))
			fmt.Println()
			fmt.Printf("Then try these commands:\n")
		}
	}
	
	fmt.Printf("  %s     # Switch between repositories\n", color.CyanString("gman switch"))
	fmt.Printf("  %s          # Check status of all repositories\n", color.CyanString("gman work status"))
	fmt.Printf("  %s       # Interactive dashboard\n", color.CyanString("gman tools dashboard"))
	fmt.Printf("  %s  # Search for files\n", color.CyanString("gman tools find file"))
	fmt.Printf("  %s # Search commits\n", color.CyanString("gman tools find commit"))
	fmt.Println()
	fmt.Printf("For more help: %s\n", color.CyanString("gman --help"))
}