package errors

import (
	"fmt"
)

// Domain-specific error builders to consolidate repetitive error creation patterns
// These builders eliminate the need for repeated fmt.Errorf calls throughout the codebase

// OperationError creates standardized operation failure errors
// Consolidates the common pattern: fmt.Errorf("failed to %s: %w", operation, err)
func OperationError(operation string, err error) *GmanError {
	return &GmanError{
		Type:    ErrTypeCommandFailed,
		Message: fmt.Sprintf("failed to %s", operation),
		Cause:   err,
	}
}

// ValidationError creates standardized validation errors
func ValidationError(field, value, reason string) *GmanError {
	return &GmanError{
		Type:    ErrTypeInvalidInput,
		Message: fmt.Sprintf("invalid %s '%s': %s", field, value, reason),
	}
}

// NotFoundError creates standardized "not found" errors
func NotFoundError(resource, identifier string) *GmanError {
	errorType := ErrTypeRepoNotFound
	switch resource {
	case "repository", "repo":
		errorType = ErrTypeRepoNotFound
	case "branch":
		errorType = ErrTypeBranchNotFound
	case "config", "configuration":
		errorType = ErrTypeConfigNotFound
	}

	return &GmanError{
		Type:    errorType,
		Message: fmt.Sprintf("%s '%s' not found", resource, identifier),
	}
}

// ConfigurationError creates configuration-related errors
func ConfigurationError(operation string, err error) *GmanError {
	gerr := &GmanError{
		Type:    ErrTypeConfigInvalid,
		Message: fmt.Sprintf("configuration error during %s", operation),
		Cause:   err,
	}
	return gerr.WithSuggestion("Check your configuration file at ~/.config/gman/config.yml")
}

// RepositoryError creates repository-related errors with context
func RepositoryError(operation, repo string, err error) *GmanError {
	gerr := &GmanError{
		Type:    ErrTypeNotGitRepo,
		Message: fmt.Sprintf("repository error in '%s' during %s", repo, operation),
		Cause:   err,
	}
	return gerr.WithSuggestion(fmt.Sprintf("Verify that '%s' is a valid Git repository", repo))
}

// ExternalToolError creates errors for missing or failing external tools
func ExternalToolError(tool, operation string, err error) *GmanError {
	gerr := &GmanError{
		Type:    ErrTypeToolNotAvailable,
		Message: fmt.Sprintf("%s not available for %s", tool, operation),
		Cause:   err,
	}
	return gerr.WithSuggestion(fmt.Sprintf("Install %s to enable this functionality", tool))
}

// NetworkError creates network-related errors
func NetworkError(operation string, err error) *GmanError {
	gerr := &GmanError{
		Type:    ErrTypeConnectFailed,
		Message: fmt.Sprintf("network error during %s", operation),
		Cause:   err,
	}
	return gerr.WithSuggestions(
		"Check your internet connection",
		"Verify remote repository URLs are correct",
		"Try again later if this is a temporary issue",
	)
}

// GitOperationError creates Git-specific operation errors
func GitOperationError(operation, repo string, err error) *GmanError {
	return &GmanError{
		Type:    ErrTypeCommandFailed,
		Message: fmt.Sprintf("Git %s failed in repository '%s'", operation, repo),
		Cause:   err,
	}
}

// WorkspaceError creates workspace state errors
func WorkspaceError(repo, issue string) *GmanError {
	gerr := &GmanError{
		Type:    ErrTypeWorkspaceNotClean,
		Message: fmt.Sprintf("workspace issue in '%s': %s", repo, issue),
	}
	return gerr.WithSuggestions(
		"Commit or stash your changes",
		"Use 'git status' to see uncommitted changes",
		"Run 'gman work status' to see all repository states",
	)
}

// PermissionError creates permission-related errors
func PermissionError(operation, resource string, err error) *GmanError {
	gerr := &GmanError{
		Type:    ErrTypePermissionDenied,
		Message: fmt.Sprintf("permission denied during %s on %s", operation, resource),
		Cause:   err,
	}
	return gerr.WithSuggestions(
		"Check file/directory permissions",
		"Ensure you have write access to the target location",
		"Try running with appropriate permissions",
	)
}

// UserCancelledError creates user cancellation errors
func UserCancelledError(operation string) *GmanError {
	return &GmanError{
		Type:    ErrTypeOperationCancelled,
		Message: fmt.Sprintf("%s cancelled by user", operation),
	}
}

// GroupError creates repository group-related errors
func GroupError(operation, groupName string, err error) *GmanError {
	gerr := &GmanError{
		Type:    ErrTypeInvalidInput,
		Message: fmt.Sprintf("group error during %s for group '%s'", operation, groupName),
		Cause:   err,
	}
	return gerr.WithSuggestions(
		"Use 'gman repo group list' to see available groups",
		"Check group name spelling",
		"Create the group first if it doesn't exist",
	)
}

// BranchError creates branch-related errors
func BranchError(operation, branch, repo string, err error) *GmanError {
	gerr := &GmanError{
		Type:    ErrTypeBranchNotFound,
		Message: fmt.Sprintf("branch error during %s: branch '%s' in repository '%s'", operation, branch, repo),
		Cause:   err,
	}
	return gerr.WithSuggestions(
		"Use 'git branch -a' to see available branches",
		"Check branch name spelling",
		"Fetch remote branches if needed",
	)
}

// WorktreeError creates worktree-related errors
func WorktreeError(operation, path string, err error) *GmanError {
	gerr := &GmanError{
		Type:    ErrTypeWorktreeExists,
		Message: fmt.Sprintf("worktree error during %s at '%s'", operation, path),
		Cause:   err,
	}
	return gerr.WithSuggestions(
		"Use 'git worktree list' to see existing worktrees",
		"Choose a different path for the worktree",
		"Remove existing worktree if no longer needed",
	)
}

// SearchError creates search-related errors
func SearchError(searchType, pattern string, err error) *GmanError {
	gerr := &GmanError{
		Type:    ErrTypeCommandFailed,
		Message: fmt.Sprintf("%s search failed for pattern '%s'", searchType, pattern),
		Cause:   err,
	}
	return gerr.WithSuggestions(
		"Check the search pattern syntax",
		"Ensure the target repositories exist",
		"Try a simpler search pattern",
	)
}

// InternalError creates internal system errors
func InternalError(component, operation string, err error) *GmanError {
	gerr := &GmanError{
		Type:    ErrTypeInternal,
		Message: fmt.Sprintf("internal error in %s during %s", component, operation),
		Cause:   err,
	}
	return gerr.WithSuggestions(
		"This appears to be an internal issue",
		"Please report this bug with the error details",
		"Try restarting the operation",
	)
}

// QuickError creates a simple error with minimal context (for backward compatibility)
func QuickError(errType ErrorType, message string) *GmanError {
	return &GmanError{
		Type:    errType,
		Message: message,
	}
}