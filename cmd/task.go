package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"gman/internal/di"
	"gman/internal/external"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// taskCmd represents the task command group
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage file task collections",
	Long: `Manage task-oriented file collections across repositories.

Tasks allow you to group related files across multiple repositories 
for development workflows and external tool integration.

Examples:
  gman task create feature-auth                    # Create new task
  gman task add feature-auth src/auth.go          # Add files to task
  gman task add feature-auth --interactive        # Interactive file selection
  gman task list-files feature-auth               # Output file paths for external tools
  gman task list-files feature-auth | xargs aider # Integration with aider
  echo "src/config.go" | gman task add feature-auth --from-stdin  # Pipe integration`,
}

// taskCreateCmd creates a new task
var taskCreateCmd = &cobra.Command{
	Use:   "create <task-name> [description]",
	Short: "Create a new task collection",
	Long: `Create a new empty task collection.

Examples:
  gman task create feature-auth
  gman task create feature-auth "Authentication system refactoring"`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runTaskCreate,
}

// taskDeleteCmd deletes a task
var taskDeleteCmd = &cobra.Command{
	Use:     "delete <task-name>",
	Short:   "Delete a task collection",
	Long:    `Delete the specified task collection. This removes the task and all its file references.`,
	Aliases: []string{"rm", "remove"},
	Args:    cobra.ExactArgs(1),
	RunE:    runTaskDelete,
}

// taskListCmd lists all tasks
var taskListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all task collections",
	Long:    `Display all configured task collections with their file counts and metadata.`,
	Aliases: []string{"ls"},
	RunE:    runTaskList,
}

// taskAddCmd adds files to a task
var taskAddCmd = &cobra.Command{
	Use:   "add <task-name> [files...]",
	Short: "Add files to a task collection",
	Long: `Add files to an existing task collection.

Examples:
  gman task add feature-auth src/auth.go src/login.go
  gman task add feature-auth --interactive        # Interactive file selection
  echo "src/config.go" | gman task add feature-auth --from-stdin`,
	Args: cobra.MinimumNArgs(1),
	RunE: runTaskAdd,
}

// taskRemoveCmd removes files from a task
var taskRemoveCmd = &cobra.Command{
	Use:   "remove <task-name> [files...]",
	Short: "Remove files from a task collection", 
	Long: `Remove files from an existing task collection.

Examples:
  gman task remove feature-auth src/old.go
  gman task remove feature-auth --interactive     # Interactive file removal`,
	Args: cobra.MinimumNArgs(1),
	RunE: runTaskRemove,
}

// taskListFilesCmd outputs file paths for external tool integration
var taskListFilesCmd = &cobra.Command{
	Use:   "list-files [task-name]",
	Short: "Output file paths for external tool integration",
	Long: `Output the file paths for a task in plain text format, suitable for piping to external tools.

If no task name is provided, lists files from all tasks.

Examples:
  gman task list-files feature-auth               # List files in specific task
  gman task list-files feature-auth | xargs aider # Integration with aider
  gman task list-files                             # List all files from all tasks`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTaskListFiles,
}

var (
	taskInteractive bool
	taskFromStdin   bool
)

func init() {
	// Add subcommands
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskDeleteCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskAddCmd)
	taskCmd.AddCommand(taskRemoveCmd)
	taskCmd.AddCommand(taskListFilesCmd)

	// Add flags
	taskAddCmd.Flags().BoolVarP(&taskInteractive, "interactive", "i", false, "Interactive file selection")
	taskAddCmd.Flags().BoolVar(&taskFromStdin, "from-stdin", false, "Read file paths from stdin")
	taskRemoveCmd.Flags().BoolVarP(&taskInteractive, "interactive", "i", false, "Interactive file removal")
}

func runTaskCreate(cmd *cobra.Command, args []string) error {
	taskName := args[0]
	description := ""
	if len(args) > 1 {
		description = args[1]
	}

	// Load configuration
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create task
	if err := configMgr.CreateTask(taskName, description); err != nil {
		return err
	}

	fmt.Printf("%s Created task '%s'\n", color.GreenString("✅"), taskName)
	if description != "" {
		fmt.Printf("   Description: %s\n", description)
	}

	return nil
}

func runTaskDelete(cmd *cobra.Command, args []string) error {
	taskName := args[0]

	// Load configuration
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Delete task
	if err := configMgr.DeleteTask(taskName); err != nil {
		return err
	}

	fmt.Printf("%s Deleted task '%s'\n", color.GreenString("✅"), taskName)
	return nil
}

func runTaskList(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	tasks := configMgr.GetTasks()
	if len(tasks) == 0 {
		fmt.Println("No tasks configured. Use 'gman task create' to create tasks.")
		return nil
	}

	// Sort tasks by name
	var taskNames []string
	for name := range tasks {
		taskNames = append(taskNames, name)
	}
	sort.Strings(taskNames)

	fmt.Printf("\n%s (%d tasks):\n\n", color.CyanString("Task Collections"), len(tasks))

	for _, name := range taskNames {
		task := tasks[name]
		fmt.Printf("%s %s (%d files)\n",
			color.YellowString("📋"),
			color.GreenString(name),
			len(task.Files))

		if task.Description != "" {
			fmt.Printf("   %s\n", color.WhiteString(task.Description))
		}

		// Show file count by repository
		repoCounts := make(map[string]int)
		for _, file := range task.Files {
			repoCounts[file.Repository]++
		}

		if len(repoCounts) > 0 {
			var repoInfo []string
			for repo, count := range repoCounts {
				repoInfo = append(repoInfo, fmt.Sprintf("%s (%d)", repo, count))
			}
			fmt.Printf("   Repositories: %s\n", color.BlueString(strings.Join(repoInfo, ", ")))
		}

		fmt.Printf("   Created: %s, Updated: %s\n\n",
			task.CreatedAt.Format("2006-01-02 15:04"),
			task.UpdatedAt.Format("2006-01-02 15:04"))
	}

	return nil
}

func runTaskAdd(cmd *cobra.Command, args []string) error {
	taskName := args[0]
	
	// Load configuration
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	var filePaths []string

	// Get file paths from different sources
	if taskFromStdin {
		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				filePaths = append(filePaths, line)
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading from stdin: %w", err)
		}
	} else if taskInteractive {
		// Interactive file selection using fd
		fdSearcher := external.NewFDSearcher()
		
		// Get all repositories for search scope
		config := configMgr.GetConfig()
		
		// Search for files across all repositories
		results, err := fdSearcher.SearchFiles("", config.Repositories, "")
		if err != nil {
			return fmt.Errorf("error searching files: %w", err)
		}
		
		if len(results) == 0 {
			fmt.Println("No files found.")
			return nil
		}
		
		// Use simple interactive selection for files
		var fileOptions []string
		for _, result := range results {
			fileOptions = append(fileOptions, result.FullPath)
		}
		
		// For now, use a simple selection mechanism
		// TODO: Implement proper multi-selection in interactive package
		fmt.Println("Available files:")
		for i, file := range fileOptions {
			fmt.Printf("%d) %s\n", i+1, file)
		}
		fmt.Print("Enter file numbers separated by spaces (e.g., 1 3 5): ")
		
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return fmt.Errorf("no input provided")
		}
		
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			return fmt.Errorf("no selection made")
		}
		
		// Parse selected indices
		parts := strings.Fields(input)
		for _, part := range parts {
			index, err := strconv.Atoi(part)
			if err != nil || index < 1 || index > len(fileOptions) {
				return fmt.Errorf("invalid selection: %s", part)
			}
			filePaths = append(filePaths, fileOptions[index-1])
		}
	} else {
		// Use files from command line arguments
		filePaths = args[1:]
	}

	if len(filePaths) == 0 {
		return fmt.Errorf("no files specified")
	}

	// Add files to task
	if err := configMgr.AddFilesToTask(taskName, filePaths); err != nil {
		return err
	}

	fmt.Printf("%s Added %d files to task '%s'\n",
		color.GreenString("✅"), len(filePaths), taskName)

	// Show added files
	for _, path := range filePaths {
		fmt.Printf("   + %s\n", color.BlueString(path))
	}

	return nil
}

func runTaskRemove(cmd *cobra.Command, args []string) error {
	taskName := args[0]
	
	// Load configuration
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	var filePaths []string

	if taskInteractive {
		// Interactive file removal - show current task files
		task, err := configMgr.GetTask(taskName)
		if err != nil {
			return err
		}
		
		if len(task.Files) == 0 {
			fmt.Printf("Task '%s' has no files.\n", taskName)
			return nil
		}
		
		// Prepare file options for selection
		var fileOptions []string
		for _, taskFile := range task.Files {
			fileOptions = append(fileOptions, taskFile.FullPath)
		}
		
		// Use simple interactive selection for removal
		fmt.Printf("Files in task '%s':\n", taskName)
		for i, file := range fileOptions {
			fmt.Printf("%d) %s\n", i+1, file)
		}
		fmt.Print("Enter file numbers to remove (separated by spaces): ")
		
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return fmt.Errorf("no input provided")
		}
		
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			return fmt.Errorf("no selection made")
		}
		
		// Parse selected indices
		parts := strings.Fields(input)
		for _, part := range parts {
			index, err := strconv.Atoi(part)
			if err != nil || index < 1 || index > len(fileOptions) {
				return fmt.Errorf("invalid selection: %s", part)
			}
			filePaths = append(filePaths, fileOptions[index-1])
		}
	} else {
		// Use files from command line arguments
		filePaths = args[1:]
	}

	if len(filePaths) == 0 {
		return fmt.Errorf("no files specified")
	}

	// Remove files from task
	if err := configMgr.RemoveFilesFromTask(taskName, filePaths); err != nil {
		return err
	}

	fmt.Printf("%s Removed %d files from task '%s'\n",
		color.GreenString("✅"), len(filePaths), taskName)

	// Show removed files
	for _, path := range filePaths {
		fmt.Printf("   - %s\n", color.RedString(path))
	}

	return nil
}

func runTaskListFiles(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	var allFiles []string

	if len(args) == 0 {
		// List files from all tasks
		tasks := configMgr.GetTasks()
		for _, task := range tasks {
			for _, taskFile := range task.Files {
				// Verify file still exists
				if _, err := os.Stat(taskFile.FullPath); err == nil {
					allFiles = append(allFiles, taskFile.FullPath)
				}
			}
		}
	} else {
		// List files from specific task
		taskName := args[0]
		files, err := configMgr.GetTaskFiles(taskName)
		if err != nil {
			return err
		}
		allFiles = files
	}

	// Output file paths (one per line) for external tool integration
	for _, file := range allFiles {
		fmt.Println(file)
	}

	return nil
}