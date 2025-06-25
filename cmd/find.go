package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"gman/internal/config"
	"gman/internal/fzf"
	"gman/internal/index"
	"github.com/fatih/color"
)

var (
	findGroupFilter  string
	findRebuildIndex bool
	findEditor       string
)

// findCmd represents the find command
var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Search across repositories using fzf",
	Long: `Search for files and commits across all managed repositories using fuzzy search.
This command integrates with fzf to provide a fast, interactive search experience
with preview capabilities.

Examples:
  gman find file                        # Search all files
  gman find file config                 # Search files matching "config"
  gman find file --group backend        # Search files in backend group
  gman find commit                      # Search all commits
  gman find commit "fix bug"            # Search commits matching "fix bug"`,
}

// findFileCmd represents the find file command
var findFileCmd = &cobra.Command{
	Use:   "file [initial_pattern]",
	Short: "Search for files across repositories",
	Long: `Search for files across all managed repositories using fuzzy search.
Files are indexed for fast searching. The search supports fuzzy matching
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
	Long: `Search for commits across all managed repositories using fuzzy search.
Commits are indexed for fast searching. The search supports fuzzy matching
on commit messages, authors, and provides real-time preview of commit diffs.

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

func init() {
	rootCmd.AddCommand(findCmd)
	findCmd.AddCommand(findFileCmd)
	findCmd.AddCommand(findCommitCmd)

	// Common flags
	findCmd.PersistentFlags().StringVar(&findGroupFilter, "group", "", "Filter by repository group")
	findCmd.PersistentFlags().BoolVar(&findRebuildIndex, "rebuild", false, "Force rebuild the search index")

	// File-specific flags
	findFileCmd.Flags().StringVar(&findEditor, "editor", "", "Editor to use when opening files (default: $EDITOR)")
}

func runFindFile(cmd *cobra.Command, args []string) error {
	// Check if fzf is available
	if !fzf.IsAvailable() {
		fmt.Fprintf(os.Stderr, "%s\n", color.RedString("❌ fzf not found"))
		fmt.Fprintf(os.Stderr, "%s\n\n", fzf.GetInstallInstructions())
		return fmt.Errorf("fzf is required for this command")
	}

	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	if len(cfg.Repositories) == 0 {
		return fmt.Errorf("no repositories configured. Use 'gman add' to add repositories")
	}

	// Initialize searcher
	searcher, err := index.NewSearcher(configMgr)
	if err != nil {
		return fmt.Errorf("failed to initialize searcher: %w", err)
	}
	defer searcher.Close()

	// Ensure index exists
	fmt.Fprintf(os.Stderr, "%s\n", color.BlueString("🔍 Preparing search index..."))
	if err := searcher.EnsureIndex(cfg.Repositories, findRebuildIndex); err != nil {
		return fmt.Errorf("failed to prepare search index: %w", err)
	}

	// Get initial search query
	var initialQuery string
	if len(args) > 0 {
		initialQuery = args[0]
	}

	// Search for files
	results, err := searcher.SearchFiles(initialQuery, findGroupFilter, cfg.Repositories)
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

	// Format results for fzf
	fzfInput := searcher.FormatFileSearchResults(results)

	// Create fzf finder
	finder, err := fzf.NewFinder()
	if err != nil {
		return fmt.Errorf("failed to create fzf finder: %w", err)
	}

	// Set up preview
	previewGen := fzf.NewPreviewGenerator()
	previewCmd := previewGen.BuildFilePreviewCommand(cfg.Repositories)

	// Configure fzf options
	opts := fzf.DefaultFileOptions()
	opts.Preview = previewCmd
	opts.InitialQuery = initialQuery
	
	// Add header with stats
	statsInfo := fmt.Sprintf("Found %d files", len(results))
	if findGroupFilter != "" {
		statsInfo += fmt.Sprintf(" in group '%s'", findGroupFilter)
	}
	opts.Header = statsInfo + " | Press Enter to select, Ctrl-O to open in editor, Ctrl-C to cancel"

	// Add key bindings
	editorCmd := getEditorCommand()
	if editorCmd != "" {
		// Create editor binding that opens the file
		editorBinding := fmt.Sprintf("ctrl-o:execute(%s {2})", editorCmd)
		opts.BindKeys = []string{editorBinding}
	}

	fmt.Fprintf(os.Stderr, "%s\n", color.GreenString("✅ Index ready. Launching fzf..."))

	// Launch fzf
	selection, err := finder.FindSingle(fzfInput, opts)
	if err != nil {
		if strings.Contains(err.Error(), "canceled") {
			fmt.Fprintf(os.Stderr, "%s\n", color.YellowString("Selection canceled"))
			return nil
		}
		return fmt.Errorf("fzf selection failed: %w", err)
	}

	// Parse selection and get file path
	selectedResult, err := searcher.ParseFileSelection(selection, results)
	if err != nil {
		return fmt.Errorf("failed to parse selection: %w", err)
	}

	// Output the selected file path
	fmt.Println(selectedResult.Path)
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
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	if len(cfg.Repositories) == 0 {
		return fmt.Errorf("no repositories configured. Use 'gman add' to add repositories")
	}

	// Initialize searcher
	searcher, err := index.NewSearcher(configMgr)
	if err != nil {
		return fmt.Errorf("failed to initialize searcher: %w", err)
	}
	defer searcher.Close()

	// Ensure index exists
	fmt.Fprintf(os.Stderr, "%s\n", color.BlueString("🔍 Preparing search index..."))
	if err := searcher.EnsureIndex(cfg.Repositories, findRebuildIndex); err != nil {
		return fmt.Errorf("failed to prepare search index: %w", err)
	}

	// Get initial search query
	var initialQuery string
	if len(args) > 0 {
		initialQuery = args[0]
	}

	// Search for commits
	results, err := searcher.SearchCommits(initialQuery, findGroupFilter, cfg.Repositories)
	if err != nil {
		return fmt.Errorf("failed to search commits: %w", err)
	}

	if len(results) == 0 {
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

	// Format results for fzf
	fzfInput := searcher.FormatCommitSearchResults(results)

	// Create fzf finder
	finder, err := fzf.NewFinder()
	if err != nil {
		return fmt.Errorf("failed to create fzf finder: %w", err)
	}

	// Set up preview
	previewGen := fzf.NewPreviewGenerator()
	previewCmd := previewGen.BuildCommitPreviewCommand(cfg.Repositories)

	// Configure fzf options
	opts := fzf.DefaultCommitOptions()
	opts.Preview = previewCmd
	opts.InitialQuery = initialQuery
	
	// Add header with stats
	statsInfo := fmt.Sprintf("Found %d commits", len(results))
	if findGroupFilter != "" {
		statsInfo += fmt.Sprintf(" in group '%s'", findGroupFilter)
	}
	opts.Header = statsInfo + " | Press Enter to select, Ctrl-C to cancel"

	fmt.Fprintf(os.Stderr, "%s\n", color.GreenString("✅ Index ready. Launching fzf..."))

	// Launch fzf
	selection, err := finder.FindSingle(fzfInput, opts)
	if err != nil {
		if strings.Contains(err.Error(), "canceled") {
			fmt.Fprintf(os.Stderr, "%s\n", color.YellowString("Selection canceled"))
			return nil
		}
		return fmt.Errorf("fzf selection failed: %w", err)
	}

	// Parse selection and get commit hash
	selectedResult, err := searcher.ParseCommitSelection(selection, results)
	if err != nil {
		return fmt.Errorf("failed to parse selection: %w", err)
	}

	// Output the selected commit hash
	fmt.Println(selectedResult.Hash)
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

// Helper function to check if a command exists
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}