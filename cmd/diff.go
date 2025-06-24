package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"gman/internal/config"
	"gman/internal/git"
)

var diffTool string

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare files across branches or repositories",
	Long: `Compare files across different branches within a repository or across different repositories.

Examples:
  # Compare a file between two branches in a repository
  gman diff <repo> <branch1> <branch2> -- <file_path>
  
  # Use an external diff tool
  gman diff <repo> <branch1> <branch2> --tool meld -- <file_path>
  
  # Compare the same file across two different repositories
  gman diff-cross-repo <repo1> <repo2> -- <file_path>`,
}

// diffFileCmd represents the diff-file command for comparing files between branches
var diffFileCmd = &cobra.Command{
	Use:   "file <repo> <branch1> <branch2> -- <file_path>",
	Short: "Compare a file between two branches in a repository",
	Long: `Compare a specific file between two branches within the same repository.

This command helps identify differences in specific files across branches,
which is useful for ensuring patches and fixes are applied consistently.

Examples:
  gman diff file myrepo main feature-branch -- src/main.go
  gman diff file myrepo ax65 ax66 -- config/settings.yml
  gman diff file myrepo --tool vimdiff main dev -- src/utils.py`,
	Args: func(cmd *cobra.Command, args []string) error {
		// Find the "--" separator
		dashIndex := -1
		for i, arg := range args {
			if arg == "--" {
				dashIndex = i
				break
			}
		}

		if dashIndex == -1 {
			return fmt.Errorf("file path must be specified after '--' separator")
		}

		if dashIndex != 3 {
			return fmt.Errorf("expected format: <repo> <branch1> <branch2> -- <file_path>")
		}

		if len(args) != 5 {
			return fmt.Errorf("expected format: <repo> <branch1> <branch2> -- <file_path>")
		}

		return nil
	},
	RunE: runDiffFile,
}

// diffCrossRepoCmd represents the diff-cross-repo command
var diffCrossRepoCmd = &cobra.Command{
	Use:   "cross-repo <repo1> <repo2> -- <file_path>",
	Short: "Compare a file between two different repositories",
	Long: `Compare the same file between two different repositories.

This is useful when you have multiple clones of the same repository or 
similar projects that should maintain consistency in certain files.

Examples:
  gman diff cross-repo project-clone1 project-clone2 -- Dockerfile
  gman diff cross-repo backend frontend -- shared/config.json`,
	Args: func(cmd *cobra.Command, args []string) error {
		// Find the "--" separator
		dashIndex := -1
		for i, arg := range args {
			if arg == "--" {
				dashIndex = i
				break
			}
		}

		if dashIndex == -1 {
			return fmt.Errorf("file path must be specified after '--' separator")
		}

		if dashIndex != 2 {
			return fmt.Errorf("expected format: <repo1> <repo2> -- <file_path>")
		}

		if len(args) != 4 {
			return fmt.Errorf("expected format: <repo1> <repo2> -- <file_path>")
		}

		return nil
	},
	RunE: runDiffCrossRepo,
}

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.AddCommand(diffFileCmd)
	diffCmd.AddCommand(diffCrossRepoCmd)
	
	// Add common flags
	diffFileCmd.Flags().StringVar(&diffTool, "tool", "", "External diff tool to use (e.g., meld, vimdiff, code)")
	diffCrossRepoCmd.Flags().StringVar(&diffTool, "tool", "", "External diff tool to use (e.g., meld, vimdiff, code)")
}

func runDiffFile(cmd *cobra.Command, args []string) error {
	repoAlias := args[0]
	branch1 := args[1]
	branch2 := args[2]
	filePath := args[4] // After "--"

	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	repoPath, exists := cfg.Repositories[repoAlias]
	if !exists {
		return fmt.Errorf("repository '%s' not found. Use 'gman list' to see available repositories", repoAlias)
	}

	// Create git manager and perform diff
	gitMgr := git.NewManager()
	
	if diffTool != "" {
		return runExternalDiffTool(gitMgr, repoPath, branch1, branch2, filePath, diffTool)
	}

	output, err := gitMgr.DiffFileBetweenBranches(repoPath, branch1, branch2, filePath)
	if err != nil {
		return fmt.Errorf("failed to diff file: %w", err)
	}

	if output == "" {
		fmt.Printf("No differences found in '%s' between branches '%s' and '%s'\n", filePath, branch1, branch2)
	} else {
		fmt.Print(output)
	}

	return nil
}

func runDiffCrossRepo(cmd *cobra.Command, args []string) error {
	repo1Alias := args[0]
	repo2Alias := args[1]
	filePath := args[3] // After "--"

	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	
	repo1Path, exists1 := cfg.Repositories[repo1Alias]
	if !exists1 {
		return fmt.Errorf("repository '%s' not found", repo1Alias)
	}
	
	repo2Path, exists2 := cfg.Repositories[repo2Alias]
	if !exists2 {
		return fmt.Errorf("repository '%s' not found", repo2Alias)
	}

	// Create git manager and perform cross-repo diff
	gitMgr := git.NewManager()
	
	if diffTool != "" {
		return runExternalCrossRepoDiffTool(gitMgr, repo1Path, repo2Path, filePath, diffTool)
	}

	output, err := gitMgr.DiffFileBetweenRepos(repo1Path, repo2Path, filePath)
	if err != nil {
		return fmt.Errorf("failed to diff file between repositories: %w", err)
	}

	if output == "" {
		fmt.Printf("No differences found in '%s' between repositories '%s' and '%s'\n", filePath, repo1Alias, repo2Alias)
	} else {
		fmt.Print(output)
	}

	return nil
}

func runExternalDiffTool(gitMgr *git.Manager, repoPath, branch1, branch2, filePath, tool string) error {
	// Get temporary files for both branches
	file1, err := gitMgr.GetFileContentFromBranch(repoPath, branch1, filePath)
	if err != nil {
		return fmt.Errorf("failed to get file content from branch %s: %w", branch1, err)
	}

	file2, err := gitMgr.GetFileContentFromBranch(repoPath, branch2, filePath)
	if err != nil {
		return fmt.Errorf("failed to get file content from branch %s: %w", branch2, err)
	}

	// Create temporary files
	tmpDir := os.TempDir()
	tmpFile1 := filepath.Join(tmpDir, fmt.Sprintf("%s_%s_%s", branch1, filepath.Base(filePath), "tmp"))
	tmpFile2 := filepath.Join(tmpDir, fmt.Sprintf("%s_%s_%s", branch2, filepath.Base(filePath), "tmp"))

	if err := os.WriteFile(tmpFile1, []byte(file1), 0644); err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile1)

	if err := os.WriteFile(tmpFile2, []byte(file2), 0644); err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile2)

	// Launch external diff tool
	cmd := exec.Command(tool, tmpFile1, tmpFile2)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func runExternalCrossRepoDiffTool(gitMgr *git.Manager, repo1Path, repo2Path, filePath, tool string) error {
	// Get file paths from both repositories
	file1Path := filepath.Join(repo1Path, filePath)
	file2Path := filepath.Join(repo2Path, filePath)

	// Check if files exist
	if _, err := os.Stat(file1Path); os.IsNotExist(err) {
		return fmt.Errorf("file '%s' does not exist in first repository", filePath)
	}
	if _, err := os.Stat(file2Path); os.IsNotExist(err) {
		return fmt.Errorf("file '%s' does not exist in second repository", filePath)
	}

	// Launch external diff tool
	cmd := exec.Command(tool, file1Path, file2Path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}