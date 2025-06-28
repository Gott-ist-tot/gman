package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gman/internal/di"
	"gman/internal/external"
	"gman/internal/fzf"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	findGroupFilter string
	findEditor      string
)

// findCmd represents the find command
var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Real-time search across repositories using modern tools",
	Long: `Search for files, content, and commits across all managed repositories using real-time tools.
This command uses fd for lightning-fast file search, ripgrep for powerful content search,
and native git log for always-current commit search, all integrated with fzf for an
interactive selection experience.

Examples:
  gman find file                        # Browse all files with fd
  gman find file config                 # Search files matching "config"
  gman find file --group backend        # Search files in backend group
  gman find content "TODO"              # Search file content with ripgrep
  gman find commit                      # Browse all commits with git log
  gman find commit "fix bug"            # Search commits matching "fix bug"`,
}

// findFileCmd represents the find file command
var findFileCmd = &cobra.Command{
	Use:   "file [initial_pattern]",
	Short: "Search for files across repositories",
	Long: `Search for files across all managed repositories using real-time search.
Uses fd for lightning-fast file discovery across repositories. The search supports fuzzy matching
and provides real-time preview of file contents.

Examples:
  gman find file                    # Browse all files
  gman find file config.yml        # Search for config.yml files
  gman find file --group frontend  # Search only in frontend group
  
Key bindings in fzf:
  Enter       - Print selected file path
  Ctrl-O      - Open file in default editor
  Ctrl-C      - Cancel selection`,
	Args: cobra.MaximumNArgs(1),
	RunE: runFindFile,
}

// findCommitCmd represents the find commit command
var findCommitCmd = &cobra.Command{
	Use:   "commit [initial_pattern]",
	Short: "Search for commits across repositories",
	Long: `Search for commits across all managed repositories using real-time git log.
Uses native git log with grep patterns for instant, always-current results.
The search supports fuzzy matching on commit messages and provides real-time 
preview of commit diffs.

Examples:
  gman find commit                  # Browse all commits
  gman find commit "fix bug"        # Search for commits with "fix bug"
  gman find commit --group backend  # Search only in backend repositories
  
Key bindings in fzf:
  Enter       - Print selected commit hash
  Ctrl-C      - Cancel selection`,
	Args: cobra.MaximumNArgs(1),
	RunE: runFindCommit,
}

// findContentCmd represents the find content command using rg
var findContentCmd = &cobra.Command{
	Use:   "content <pattern>",
	Short: "Search file contents across repositories using ripgrep",
	Long: `Search for text within files across all managed repositories using ripgrep.
This provides real-time content search with regex support and is much faster
than traditional file indexing approaches.

Examples:
  gman find content "TODO"               # Search for TODO comments
  gman find content "func.*Error"       # Search using regex patterns
  gman find content "import.*react" --group frontend  # Search in specific group
  
Key bindings in fzf:
  Enter       - Print selected file path with line number
  Ctrl-C      - Cancel selection`,
	Args: cobra.ExactArgs(1),
	RunE: runFindContent,
}

func init() {
	// Command is now available via: gman tools find
	// Removed direct rootCmd registration to avoid duplication
	findCmd.AddCommand(findFileCmd)
	findCmd.AddCommand(findCommitCmd)
	findCmd.AddCommand(findContentCmd)

	// Common flags
	findCmd.PersistentFlags().StringVar(&findGroupFilter, "group", "", "Filter by repository group")

	// File-specific flags
	findFileCmd.Flags().StringVar(&findEditor, "editor", "", "Editor to use when opening files (default: $EDITOR)")
	
	// Content-specific flags
	findContentCmd.Flags().StringVar(&findEditor, "editor", "", "Editor to use when opening files (default: $EDITOR)")
}

func runFindFile(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	if len(cfg.Repositories) == 0 {
		return fmt.Errorf("no repositories configured. Use 'gman repo add' to add repositories")
	}

	// Get initial search query
	var initialQuery string
	if len(args) > 0 {
		initialQuery = args[0]
	}

	// Initialize smart searcher (automatically handles fallbacks)
	verbose := cmd.Flag("verbose") != nil && cmd.Flag("verbose").Value.String() == "true"
	searcher := external.NewSmartSearcher(verbose)
	
	// Show optimization tips if tools are missing
	if searcher.GetDiagnostics().GetReadiness() < 100 {
		fmt.Fprintf(os.Stderr, "⚡ %s file search (tool availability: %d%%)\n", 
			color.BlueString("Starting"), searcher.GetDiagnostics().GetReadiness())
		if verbose {
			searcher.ShowOptimizationTips()
			fmt.Fprintln(os.Stderr)
		}
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", color.BlueString("🔍 Searching files with optimized tools..."))
	}

	// Search for files using intelligent search strategy
	results, err := searcher.SearchFiles(initialQuery, cfg.Repositories, findGroupFilter)
	if err != nil {
		return fmt.Errorf("failed to search files: %w", err)
	}

	if len(results) == 0 {
		fmt.Printf("%s No files found", color.YellowString("⚠️"))
		if initialQuery != "" {
			fmt.Printf(" matching '%s'", initialQuery)
		}
		if findGroupFilter != "" {
			fmt.Printf(" in group '%s'", findGroupFilter)
		}
		fmt.Println()
		return nil
	}

	fmt.Fprintf(os.Stderr, "%s Found %d files. Starting selection...\n", 
		color.GreenString("✅"), len(results))

	// Use intelligent selection (fzf or fallback)
	prompt := fmt.Sprintf("📁 Select a file to view (%d results)", len(results))
	selectedFile, err := searcher.SelectFromResults(results, prompt)
	if err != nil {
		if strings.Contains(err.Error(), "cancelled") {
			fmt.Println("Selection cancelled.")
			return nil
		}
		return fmt.Errorf("selection failed: %w", err)
	}

	// Output the selected file path
	fmt.Println(selectedFile.FullPath)
	return nil
}

func runFindCommit(cmd *cobra.Command, args []string) error {
	// Check if fzf is available
	if !fzf.IsAvailable() {
		fmt.Fprintf(os.Stderr, "%s\n", color.RedString("❌ fzf not found"))
		fmt.Fprintf(os.Stderr, "%s\n\n", fzf.GetInstallInstructions())
		return fmt.Errorf("fzf is required for this command")
	}

	// Load configuration
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	if len(cfg.Repositories) == 0 {
		return fmt.Errorf("no repositories configured. Use 'gman add' to add repositories")
	}

	// Get initial search query
	var initialQuery string
	if len(args) > 0 {
		initialQuery = args[0]
	}

	fmt.Fprintf(os.Stderr, "%s\n", color.BlueString("🔍 Searching commits with real-time git log..."))

	// Filter repositories by group if specified
	repositories := cfg.Repositories
	if findGroupFilter != "" {
		groupRepos, err := configMgr.GetGroupRepositories(findGroupFilter)
		if err != nil {
			return fmt.Errorf("failed to get group repositories: %w", err)
		}
		repositories = groupRepos
	}

	// Collect commits from all repositories using git log
	var allCommits []string
	var totalCommits int

	for alias, path := range repositories {
		// Build git log command - search commit messages if query provided
		gitArgs := []string{"log", "--oneline", "--all", "--decorate", "--color=always", "-n", "100"}
		if initialQuery != "" {
			gitArgs = append(gitArgs, fmt.Sprintf("--grep=%s", initialQuery))
		}

		// Execute git log
		gitCmd := exec.Command("git", gitArgs...)
		gitCmd.Dir = path
		output, err := gitCmd.Output()
		if err != nil {
			// Skip repositories that don't have commits or have errors
			continue
		}

		if len(output) > 0 {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, line := range lines {
				if line != "" {
					// Prefix each commit with repository alias
					commitEntry := fmt.Sprintf("[%s] %s", color.CyanString(alias), line)
					allCommits = append(allCommits, commitEntry)
					totalCommits++
				}
			}
		}
	}

	if len(allCommits) == 0 {
		fmt.Printf("%s No commits found", color.YellowString("⚠️"))
		if initialQuery != "" {
			fmt.Printf(" matching '%s'", initialQuery)
		}
		if findGroupFilter != "" {
			fmt.Printf(" in group '%s'", findGroupFilter)
		}
		fmt.Println()
		return nil
	}

	// Create fzf finder
	finder, err := fzf.NewFinder()
	if err != nil {
		return fmt.Errorf("failed to create fzf finder: %w", err)
	}

	// Configure fzf options for commit search
	opts := fzf.DefaultCommitOptions()
	opts.InitialQuery = initialQuery
	
	// Add header with stats
	statsInfo := fmt.Sprintf("Found %d commits across %d repositories", totalCommits, len(repositories))
	if findGroupFilter != "" {
		statsInfo += fmt.Sprintf(" in group '%s'", findGroupFilter)
	}
	opts.Header = statsInfo + " | Press Enter to select, Ctrl-C to cancel"

	// Set up preview command for commit details
	opts.Preview = `
		line={1}
		repo=$(echo "$line" | sed 's/\[//g' | sed 's/\].*//g')
		hash=$(echo "$line" | awk '{print $2}')
		if [ ! -z "$repo" ] && [ ! -z "$hash" ]; then
			# Find repository path
			repo_path=""
			` + func() string {
		var paths []string
		for alias, path := range repositories {
			paths = append(paths, fmt.Sprintf(`if [ "$repo" = "%s" ]; then repo_path="%s"; fi`, alias, path))
		}
		return strings.Join(paths, "\n			")
	}() + `
			if [ ! -z "$repo_path" ]; then
				cd "$repo_path" && git show --color=always "$hash"
			fi
		fi
	`

	fmt.Fprintf(os.Stderr, "%s\n", color.GreenString("✅ Ready. Launching fzf..."))

	// Launch fzf with commit data
	selection, err := finder.FindSingle(allCommits, opts)
	if err != nil {
		if strings.Contains(err.Error(), "canceled") {
			fmt.Fprintf(os.Stderr, "%s\n", color.YellowString("Selection canceled"))
			return nil
		}
		return fmt.Errorf("fzf selection failed: %w", err)
	}

	// Parse selection to extract commit hash
	parts := strings.Fields(selection)
	if len(parts) >= 2 {
		// Extract commit hash (second field after repository name)
		commitHash := parts[1]
		fmt.Println(commitHash)
	} else {
		return fmt.Errorf("invalid selection format")
	}

	return nil
}

// getEditorCommand returns the editor command to use
func getEditorCommand() string {
	// Check command line flag first
	if findEditor != "" {
		return findEditor
	}

	// Check environment variables
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}

	// Try common editors
	editors := []string{"nvim", "vim", "nano", "emacs", "code", "subl"}
	for _, editor := range editors {
		if _, err := exec.LookPath(editor); err == nil {
			return editor
		}
	}

	return ""
}

func runFindContent(cmd *cobra.Command, args []string) error {
	// Check dependencies
	missingTools := external.CheckDependencies(external.RipGrep, external.FZF)
	if len(missingTools) > 0 {
		fmt.Fprintf(os.Stderr, "%s\n", color.RedString("❌ Missing required tools: %s", strings.Join(missingTools, ", ")))
		fmt.Fprintf(os.Stderr, "%s\n", external.GetMissingToolsMessage(external.RipGrep, external.FZF))
		return fmt.Errorf("required tools not available")
	}

	// Load configuration
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	if len(cfg.Repositories) == 0 {
		return fmt.Errorf("no repositories configured. Use 'gman repo add' to add repositories")
	}

	// Get search pattern (required)
	searchPattern := args[0]

	// Initialize rg searcher
	searcher := external.NewRGSearcher()
	
	fmt.Fprintf(os.Stderr, "%s\n", color.BlueString("🔍 Searching content with ripgrep..."))

	// Search for content using rg
	results, err := searcher.SearchContent(searchPattern, cfg.Repositories, findGroupFilter)
	if err != nil {
		return fmt.Errorf("failed to search content: %w", err)
	}

	if len(results) == 0 {
		fmt.Printf("%s No content found", color.YellowString("⚠️"))
		fmt.Printf(" matching '%s'", searchPattern)
		if findGroupFilter != "" {
			fmt.Printf(" in group '%s'", findGroupFilter)
		}
		fmt.Println()
		return nil
	}

	// Format results for fzf
	fzfInput := searcher.FormatForFZF(results)
	fzfLines := strings.Split(fzfInput, "\n")

	// Create fzf finder
	finder, err := fzf.NewFinder()
	if err != nil {
		return fmt.Errorf("failed to create fzf finder: %w", err)
	}

	// Configure fzf options for content search
	opts := fzf.DefaultFileOptions()
	opts.Multi = false
	opts.Height = "80%"
	opts.Layout = "reverse"
	opts.Border = true

	// Add header with stats
	statsInfo := fmt.Sprintf("Found %d matches for '%s'", len(results), searchPattern)
	if findGroupFilter != "" {
		statsInfo += fmt.Sprintf(" in group '%s'", findGroupFilter)
	}
	opts.Header = statsInfo + " | Press Enter to select, Ctrl-O to open in editor, Ctrl-C to cancel"

	// Add key bindings
	editorCmd := getEditorCommand()
	if editorCmd != "" {
		// Create editor binding that opens the file at the specific line
		// New format is: "absolute_path:line_number:display_text"
		// We extract the path (field 1) and line number (field 2) using cut
		editorBinding := fmt.Sprintf("ctrl-o:execute(%s +$(echo {} | cut -d: -f2) \"$(echo {} | cut -d: -f1)\")", editorCmd)
		opts.BindKeys = []string{editorBinding}
	}

	fmt.Fprintf(os.Stderr, "%s\n", color.GreenString("✅ Search complete. Launching fzf..."))

	// Launch fzf
	selection, err := finder.FindSingle(fzfLines, opts)
	if err != nil {
		if strings.Contains(err.Error(), "canceled") {
			fmt.Fprintf(os.Stderr, "%s\n", color.YellowString("Selection canceled"))
			return nil
		}
		return fmt.Errorf("fzf selection failed: %w", err)
	}

	// Parse selection and get file path with line number
	selectedResult, err := searcher.ParseFZFSelection(selection, results)
	if err != nil {
		return fmt.Errorf("failed to parse selection: %w", err)
	}

	// Output the selected file path with line number for easy editor navigation
	// Use FullPath for absolute path access
	fmt.Printf("%s:%d\n", selectedResult.FullPath, selectedResult.LineNumber)
	return nil
}

// Helper function to check if a command exists
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
