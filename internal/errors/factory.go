package errors

import "fmt"

// NewGmanError creates a new GmanError with the specified type and message
func NewGmanError(errorType ErrorType, message string) *GmanError {
	return &GmanError{
		Type:    errorType,
		Message: message,
	}
}

// Repository-related error factories

// NewRepoNotFoundError creates an error for when a repository is not found
func NewRepoNotFoundError(path string) *GmanError {
	return NewGmanError(ErrTypeRepoNotFound, fmt.Sprintf("Repository not found at path: %s", path)).
		WithSuggestions(
			"Check if the path exists and is accessible",
			"Verify the repository path in your configuration",
			"Use 'gman repo list' to see configured repositories",
		)
}

// NewNotGitRepoError creates an error for when a directory is not a git repository
func NewNotGitRepoError(path string) *GmanError {
	return NewGmanError(ErrTypeNotGitRepo, fmt.Sprintf("Directory is not a Git repository: %s", path)).
		WithSuggestions(
			"Initialize a Git repository with 'git init'",
			"Clone an existing repository",
			"Check if .git directory exists",
		)
}

// NewRepoAlreadyExistsError creates an error for when a repository alias already exists
func NewRepoAlreadyExistsError(alias string) *GmanError {
	return NewGmanError(ErrTypeRepoAlreadyExists, fmt.Sprintf("Repository alias '%s' already exists", alias)).
		WithSuggestions(
			"Use a different alias name",
			"Remove the existing repository first with 'gman repo remove "+alias+"'",
			"Use 'gman repo list' to see existing repositories",
		)
}

// Git operation error factories

// NewMergeConflictError creates an error for merge conflicts
func NewMergeConflictError(repoPath string) *GmanError {
	return NewGmanError(ErrTypeMergeConflict, fmt.Sprintf("Merge conflict detected in repository: %s", repoPath)).
		WithSuggestions(
			"Resolve conflicts manually in your editor",
			"Use 'git status' to see conflicted files",
			"Run 'git add <file>' after resolving conflicts",
			"Complete merge with 'git commit'",
		)
}

// NewRemoteUnreachableError creates an error for unreachable remotes
func NewRemoteUnreachableError(remote string, repoPath string) *GmanError {
	return NewGmanError(ErrTypeRemoteUnreachable, fmt.Sprintf("Cannot reach remote '%s' for repository: %s", remote, repoPath)).
		WithSuggestions(
			"Check your internet connection",
			"Verify remote URL with 'git remote -v'",
			"Check if you have access to the remote repository",
			"Try again later if it's a temporary network issue",
		)
}

// NewWorkspaceNotCleanError creates an error for uncommitted changes
func NewWorkspaceNotCleanError(repoPath string) *GmanError {
	return NewGmanError(ErrTypeWorkspaceNotClean, fmt.Sprintf("Repository has uncommitted changes: %s", repoPath)).
		WithSuggestions(
			"Commit your changes with 'git commit -am \"message\"'",
			"Stash changes with 'git stash'",
			"Discard changes with 'git checkout -- .' (warning: destructive)",
			"Use 'git status' to see what needs to be committed",
		)
}

// NewBranchNotFoundError creates an error for non-existent branches
func NewBranchNotFoundError(branch string, repoPath string) *GmanError {
	return NewGmanError(ErrTypeBranchNotFound, fmt.Sprintf("Branch '%s' not found in repository: %s", branch, repoPath)).
		WithSuggestions(
			"Check available branches with 'git branch -a'",
			"Create the branch with 'git checkout -b "+branch+"'",
			"Switch to an existing branch",
		)
}

// NewWorktreeExistsError creates an error for existing worktrees
func NewWorktreeExistsError(path string) *GmanError {
	return NewGmanError(ErrTypeWorktreeExists, fmt.Sprintf("Worktree already exists at path: %s", path)).
		WithSuggestions(
			"Use a different path for the worktree",
			"Remove existing worktree with 'git worktree remove "+path+"'",
			"List existing worktrees with 'git worktree list'",
		)
}

// Configuration error factories

// NewConfigInvalidError creates an error for invalid configuration
func NewConfigInvalidError(reason string) *GmanError {
	return NewGmanError(ErrTypeConfigInvalid, fmt.Sprintf("Configuration is invalid: %s", reason)).
		WithSuggestions(
			"Check YAML syntax in configuration file",
			"Restore from backup if available",
			"Reset configuration with 'gman setup'",
		)
}

// NewConfigNotFoundError creates an error for missing configuration
func NewConfigNotFoundError(path string) *GmanError {
	return NewGmanError(ErrTypeConfigNotFound, fmt.Sprintf("Configuration file not found: %s", path)).
		WithSuggestions(
			"Run 'gman setup' to create initial configuration",
			"Check if the configuration directory exists",
			"Verify file permissions",
		)
}

// NewPermissionDeniedError creates an error for permission issues
func NewPermissionDeniedError(resource string) *GmanError {
	return NewGmanError(ErrTypePermissionDenied, fmt.Sprintf("Permission denied accessing: %s", resource)).
		WithSuggestions(
			"Check file/directory permissions",
			"Run with appropriate user privileges",
			"Verify ownership of files and directories",
		)
}

// External tool error factories

// NewToolNotAvailableError creates an error for missing tools
func NewToolNotAvailableError(tool string, installInstructions string) *GmanError {
	err := NewGmanError(ErrTypeToolNotAvailable, fmt.Sprintf("Required tool '%s' is not available", tool))
		
	if installInstructions != "" {
		err.WithSuggestion("Install with: " + installInstructions)
	}
	err.WithSuggestions(
		"Use 'gman tools check' to verify tool availability",
		"Check if the tool is in your PATH",
	)
	
	return err
}

// NewCommandFailedError creates an error for failed commands
func NewCommandFailedError(command string, exitCode int, output string) *GmanError {
	message := fmt.Sprintf("Command failed: %s (exit code: %d)", command, exitCode)
	if output != "" {
		message += fmt.Sprintf("\nOutput: %s", output)
	}
	
	return NewGmanError(ErrTypeCommandFailed, message).
		WithSuggestions(
			"Check command syntax and arguments",
			"Verify required permissions",
			"Review error output for specific issues",
		)
}

// Network error factories

// NewNetworkTimeoutError creates an error for network timeouts
func NewNetworkTimeoutError(operation string, timeout string) *GmanError {
	return NewGmanError(ErrTypeNetworkTimeout, fmt.Sprintf("Network timeout during %s (timeout: %s)", operation, timeout)).
		WithSuggestions(
			"Check your internet connection",
			"Try again with a longer timeout",
			"Verify the remote server is responding",
		)
}

// NewConnectFailedError creates an error for connection failures
func NewConnectFailedError(target string, reason string) *GmanError {
	return NewGmanError(ErrTypeConnectFailed, fmt.Sprintf("Failed to connect to %s: %s", target, reason)).
		WithSuggestions(
			"Check network connectivity",
			"Verify the target address is correct",
			"Check firewall and proxy settings",
		)
}

// User input error factories

// NewInvalidInputError creates an error for invalid user input
func NewInvalidInputError(input string, reason string) *GmanError {
	return NewGmanError(ErrTypeInvalidInput, fmt.Sprintf("Invalid input '%s': %s", input, reason)).
		WithSuggestions(
			"Check the input format and try again",
			"Use --help to see valid options",
		)
}

// NewOperationCancelledError creates an error for cancelled operations
func NewOperationCancelledError(operation string) *GmanError {
	return NewGmanError(ErrTypeOperationCancelled, fmt.Sprintf("Operation cancelled: %s", operation))
}

// Internal error factories

// NewInternalError creates an error for internal system errors
func NewInternalError(component string, reason string) *GmanError {
	return NewGmanError(ErrTypeInternal, fmt.Sprintf("Internal error in %s: %s", component, reason)).
		WithSuggestions(
			"This is likely a bug - please report it",
			"Try restarting the operation",
			"Check for software updates",
		)
}

// NewNotImplementedError creates an error for unimplemented features
func NewNotImplementedError(feature string) *GmanError {
	return NewGmanError(ErrTypeNotImplemented, fmt.Sprintf("Feature not implemented: %s", feature)).
		WithSuggestions(
			"This feature is planned for a future release",
			"Check documentation for alternative approaches",
		)
}

