package errors

import (
	"fmt"
	"strings"
	"testing"
)

func TestOperationError(t *testing.T) {
	baseErr := fmt.Errorf("underlying error")
	
	err := OperationError("test operation", baseErr)
	
	if err == nil {
		t.Fatal("OperationError() returned nil")
	}
	
	if err.Type != ErrTypeCommandFailed {
		t.Errorf("Expected error type %s, got %s", ErrTypeCommandFailed, err.Type)
	}
	
	expectedMessage := "failed to test operation"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, err.Message)
	}
	
	if err.Cause != baseErr {
		t.Error("Expected cause to be preserved")
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		field    string
		value    string
		reason   string
		expected string
	}{
		{
			field:    "repository",
			value:    "invalid-repo",
			reason:   "not found",
			expected: "invalid repository 'invalid-repo': not found",
		},
		{
			field:    "group",
			value:    "test-group",
			reason:   "already exists",
			expected: "invalid group 'test-group': already exists",
		},
		{
			field:    "path",
			value:    "/invalid/path",
			reason:   "does not exist",
			expected: "invalid path '/invalid/path': does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.field, tt.reason), func(t *testing.T) {
			err := ValidationError(tt.field, tt.value, tt.reason)
			
			if err == nil {
				t.Fatal("ValidationError() returned nil")
			}
			
			if err.Type != ErrTypeInvalidInput {
				t.Errorf("Expected error type %s, got %s", ErrTypeInvalidInput, err.Type)
			}
			
			if err.Message != tt.expected {
				t.Errorf("Expected message '%s', got '%s'", tt.expected, err.Message)
			}
		})
	}
}

func TestNotFoundError(t *testing.T) {
	tests := []struct {
		resource     string
		identifier   string
		expectedType ErrorType
		expectedMsg  string
	}{
		{
			resource:     "repository",
			identifier:   "test-repo",
			expectedType: ErrTypeRepoNotFound,
			expectedMsg:  "repository 'test-repo' not found",
		},
		{
			resource:     "repo",
			identifier:   "test-repo",
			expectedType: ErrTypeRepoNotFound,
			expectedMsg:  "repo 'test-repo' not found",
		},
		{
			resource:     "branch",
			identifier:   "feature-branch",
			expectedType: ErrTypeBranchNotFound,
			expectedMsg:  "branch 'feature-branch' not found",
		},
		{
			resource:     "config",
			identifier:   "settings.yml",
			expectedType: ErrTypeConfigNotFound,
			expectedMsg:  "config 'settings.yml' not found",
		},
		{
			resource:     "configuration",
			identifier:   "app.conf",
			expectedType: ErrTypeConfigNotFound,
			expectedMsg:  "configuration 'app.conf' not found",
		},
		{
			resource:     "unknown",
			identifier:   "test",
			expectedType: ErrTypeRepoNotFound, // defaults to repo not found
			expectedMsg:  "unknown 'test' not found",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.resource, tt.identifier), func(t *testing.T) {
			err := NotFoundError(tt.resource, tt.identifier)
			
			if err == nil {
				t.Fatal("NotFoundError() returned nil")
			}
			
			if err.Type != tt.expectedType {
				t.Errorf("Expected error type %s, got %s", tt.expectedType, err.Type)
			}
			
			if err.Message != tt.expectedMsg {
				t.Errorf("Expected message '%s', got '%s'", tt.expectedMsg, err.Message)
			}
		})
	}
}

func TestConfigurationError(t *testing.T) {
	baseErr := fmt.Errorf("parsing error")
	
	err := ConfigurationError("loading", baseErr)
	
	if err == nil {
		t.Fatal("ConfigurationError() returned nil")
	}
	
	if err.Type != ErrTypeConfigInvalid {
		t.Errorf("Expected error type %s, got %s", ErrTypeConfigInvalid, err.Type)
	}
	
	expectedMessage := "configuration error during loading"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, err.Message)
	}
	
	if err.Cause != baseErr {
		t.Error("Expected cause to be preserved")
	}
	
	// Should have suggestion
	if len(err.Suggestions) == 0 {
		t.Error("Expected suggestions to be added")
	}
	
	expectedSuggestion := "Check your configuration file at ~/.config/gman/config.yml"
	if !containsSuggestion(err.Suggestions, expectedSuggestion) {
		t.Errorf("Expected suggestion '%s' to be present", expectedSuggestion)
	}
}

func TestRepositoryError(t *testing.T) {
	baseErr := fmt.Errorf("git error")
	
	err := RepositoryError("status check", "test-repo", baseErr)
	
	if err == nil {
		t.Fatal("RepositoryError() returned nil")
	}
	
	if err.Type != ErrTypeNotGitRepo {
		t.Errorf("Expected error type %s, got %s", ErrTypeNotGitRepo, err.Type)
	}
	
	expectedMessage := "repository error in 'test-repo' during status check"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, err.Message)
	}
	
	if err.Cause != baseErr {
		t.Error("Expected cause to be preserved")
	}
	
	// Should have suggestion
	if len(err.Suggestions) == 0 {
		t.Error("Expected suggestions to be added")
	}
	
	expectedSuggestion := "Verify that 'test-repo' is a valid Git repository"
	if !containsSuggestion(err.Suggestions, expectedSuggestion) {
		t.Errorf("Expected suggestion '%s' to be present", expectedSuggestion)
	}
}

func TestExternalToolError(t *testing.T) {
	baseErr := fmt.Errorf("command not found")
	
	err := ExternalToolError("git", "status check", baseErr)
	
	if err == nil {
		t.Fatal("ExternalToolError() returned nil")
	}
	
	if err.Type != ErrTypeToolNotAvailable {
		t.Errorf("Expected error type %s, got %s", ErrTypeToolNotAvailable, err.Type)
	}
	
	expectedMessage := "git not available for status check"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, err.Message)
	}
	
	if err.Cause != baseErr {
		t.Error("Expected cause to be preserved")
	}
	
	// Should have suggestion
	expectedSuggestion := "Install git to enable this functionality"
	if !containsSuggestion(err.Suggestions, expectedSuggestion) {
		t.Errorf("Expected suggestion '%s' to be present", expectedSuggestion)
	}
}

func TestNetworkError(t *testing.T) {
	baseErr := fmt.Errorf("connection timeout")
	
	err := NetworkError("fetching updates", baseErr)
	
	if err == nil {
		t.Fatal("NetworkError() returned nil")
	}
	
	if err.Type != ErrTypeConnectFailed {
		t.Errorf("Expected error type %s, got %s", ErrTypeConnectFailed, err.Type)
	}
	
	expectedMessage := "network error during fetching updates"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, err.Message)
	}
	
	if err.Cause != baseErr {
		t.Error("Expected cause to be preserved")
	}
	
	// Should have multiple suggestions
	if len(err.Suggestions) < 3 {
		t.Error("Expected multiple network troubleshooting suggestions")
	}
	
	expectedSuggestions := []string{
		"Check your internet connection",
		"Verify remote repository URLs are correct",
		"Try again later if this is a temporary issue",
	}
	
	for _, suggestion := range expectedSuggestions {
		if !containsSuggestion(err.Suggestions, suggestion) {
			t.Errorf("Expected suggestion '%s' to be present", suggestion)
		}
	}
}

func TestGitOperationError(t *testing.T) {
	baseErr := fmt.Errorf("merge conflict")
	
	err := GitOperationError("merge", "test-repo", baseErr)
	
	if err == nil {
		t.Fatal("GitOperationError() returned nil")
	}
	
	if err.Type != ErrTypeCommandFailed {
		t.Errorf("Expected error type %s, got %s", ErrTypeCommandFailed, err.Type)
	}
	
	expectedMessage := "Git merge failed in repository 'test-repo'"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, err.Message)
	}
	
	if err.Cause != baseErr {
		t.Error("Expected cause to be preserved")
	}
}

func TestWorkspaceError(t *testing.T) {
	err := WorkspaceError("test-repo", "uncommitted changes")
	
	if err == nil {
		t.Fatal("WorkspaceError() returned nil")
	}
	
	if err.Type != ErrTypeWorkspaceNotClean {
		t.Errorf("Expected error type %s, got %s", ErrTypeWorkspaceNotClean, err.Type)
	}
	
	expectedMessage := "workspace issue in 'test-repo': uncommitted changes"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, err.Message)
	}
	
	// Should have workspace-specific suggestions
	expectedSuggestions := []string{
		"Commit or stash your changes",
		"Use 'git status' to see uncommitted changes",
		"Run 'gman work status' to see all repository states",
	}
	
	for _, suggestion := range expectedSuggestions {
		if !containsSuggestion(err.Suggestions, suggestion) {
			t.Errorf("Expected suggestion '%s' to be present", suggestion)
		}
	}
}

func TestPermissionError(t *testing.T) {
	baseErr := fmt.Errorf("access denied")
	
	err := PermissionError("writing file", "/protected/file", baseErr)
	
	if err == nil {
		t.Fatal("PermissionError() returned nil")
	}
	
	if err.Type != ErrTypePermissionDenied {
		t.Errorf("Expected error type %s, got %s", ErrTypePermissionDenied, err.Type)
	}
	
	expectedMessage := "permission denied during writing file on /protected/file"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, err.Message)
	}
	
	if err.Cause != baseErr {
		t.Error("Expected cause to be preserved")
	}
	
	// Should have permission-specific suggestions
	expectedSuggestions := []string{
		"Check file/directory permissions",
		"Ensure you have write access to the target location",
		"Try running with appropriate permissions",
	}
	
	for _, suggestion := range expectedSuggestions {
		if !containsSuggestion(err.Suggestions, suggestion) {
			t.Errorf("Expected suggestion '%s' to be present", suggestion)
		}
	}
}

func TestUserCancelledError(t *testing.T) {
	err := UserCancelledError("repository selection")
	
	if err == nil {
		t.Fatal("UserCancelledError() returned nil")
	}
	
	if err.Type != ErrTypeOperationCancelled {
		t.Errorf("Expected error type %s, got %s", ErrTypeOperationCancelled, err.Type)
	}
	
	expectedMessage := "repository selection cancelled by user"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, err.Message)
	}
	
	// Should not have a cause
	if err.Cause != nil {
		t.Error("User cancelled errors should not have a cause")
	}
}

func TestGroupError(t *testing.T) {
	baseErr := fmt.Errorf("group not found")
	
	err := GroupError("listing repositories", "test-group", baseErr)
	
	if err == nil {
		t.Fatal("GroupError() returned nil")
	}
	
	if err.Type != ErrTypeInvalidInput {
		t.Errorf("Expected error type %s, got %s", ErrTypeInvalidInput, err.Type)
	}
	
	expectedMessage := "group error during listing repositories for group 'test-group'"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, err.Message)
	}
	
	if err.Cause != baseErr {
		t.Error("Expected cause to be preserved")
	}
	
	// Should have group-specific suggestions
	expectedSuggestions := []string{
		"Use 'gman repo group list' to see available groups",
		"Check group name spelling",
		"Create the group first if it doesn't exist",
	}
	
	for _, suggestion := range expectedSuggestions {
		if !containsSuggestion(err.Suggestions, suggestion) {
			t.Errorf("Expected suggestion '%s' to be present", suggestion)
		}
	}
}

func TestBranchError(t *testing.T) {
	baseErr := fmt.Errorf("branch not found")
	
	err := BranchError("checkout", "feature-branch", "test-repo", baseErr)
	
	if err == nil {
		t.Fatal("BranchError() returned nil")
	}
	
	if err.Type != ErrTypeBranchNotFound {
		t.Errorf("Expected error type %s, got %s", ErrTypeBranchNotFound, err.Type)
	}
	
	expectedMessage := "branch error during checkout: branch 'feature-branch' in repository 'test-repo'"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, err.Message)
	}
	
	if err.Cause != baseErr {
		t.Error("Expected cause to be preserved")
	}
	
	// Should have branch-specific suggestions
	expectedSuggestions := []string{
		"Use 'git branch -a' to see available branches",
		"Check branch name spelling",
		"Fetch remote branches if needed",
	}
	
	for _, suggestion := range expectedSuggestions {
		if !containsSuggestion(err.Suggestions, suggestion) {
			t.Errorf("Expected suggestion '%s' to be present", suggestion)
		}
	}
}

func TestWorktreeError(t *testing.T) {
	baseErr := fmt.Errorf("worktree already exists")
	
	err := WorktreeError("creation", "/path/to/worktree", baseErr)
	
	if err == nil {
		t.Fatal("WorktreeError() returned nil")
	}
	
	if err.Type != ErrTypeWorktreeExists {
		t.Errorf("Expected error type %s, got %s", ErrTypeWorktreeExists, err.Type)
	}
	
	expectedMessage := "worktree error during creation at '/path/to/worktree'"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, err.Message)
	}
	
	if err.Cause != baseErr {
		t.Error("Expected cause to be preserved")
	}
	
	// Should have worktree-specific suggestions
	expectedSuggestions := []string{
		"Use 'git worktree list' to see existing worktrees",
		"Choose a different path for the worktree",
		"Remove existing worktree if no longer needed",
	}
	
	for _, suggestion := range expectedSuggestions {
		if !containsSuggestion(err.Suggestions, suggestion) {
			t.Errorf("Expected suggestion '%s' to be present", suggestion)
		}
	}
}

func TestSearchError(t *testing.T) {
	baseErr := fmt.Errorf("invalid regex")
	
	err := SearchError("file", "*.go[", baseErr)
	
	if err == nil {
		t.Fatal("SearchError() returned nil")
	}
	
	if err.Type != ErrTypeCommandFailed {
		t.Errorf("Expected error type %s, got %s", ErrTypeCommandFailed, err.Type)
	}
	
	expectedMessage := "file search failed for pattern '*.go['"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, err.Message)
	}
	
	if err.Cause != baseErr {
		t.Error("Expected cause to be preserved")
	}
	
	// Should have search-specific suggestions
	expectedSuggestions := []string{
		"Check the search pattern syntax",
		"Ensure the target repositories exist",
		"Try a simpler search pattern",
	}
	
	for _, suggestion := range expectedSuggestions {
		if !containsSuggestion(err.Suggestions, suggestion) {
			t.Errorf("Expected suggestion '%s' to be present", suggestion)
		}
	}
}

func TestInternalError(t *testing.T) {
	baseErr := fmt.Errorf("panic recovered")
	
	err := InternalError("config parser", "loading defaults", baseErr)
	
	if err == nil {
		t.Fatal("InternalError() returned nil")
	}
	
	if err.Type != ErrTypeInternal {
		t.Errorf("Expected error type %s, got %s", ErrTypeInternal, err.Type)
	}
	
	expectedMessage := "internal error in config parser during loading defaults"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, err.Message)
	}
	
	if err.Cause != baseErr {
		t.Error("Expected cause to be preserved")
	}
	
	// Should have internal error suggestions
	expectedSuggestions := []string{
		"This appears to be an internal issue",
		"Please report this bug with the error details",
		"Try restarting the operation",
	}
	
	for _, suggestion := range expectedSuggestions {
		if !containsSuggestion(err.Suggestions, suggestion) {
			t.Errorf("Expected suggestion '%s' to be present", suggestion)
		}
	}
}

func TestQuickError(t *testing.T) {
	err := QuickError(ErrTypeInvalidInput, "test message")
	
	if err == nil {
		t.Fatal("QuickError() returned nil")
	}
	
	if err.Type != ErrTypeInvalidInput {
		t.Errorf("Expected error type %s, got %s", ErrTypeInvalidInput, err.Type)
	}
	
	if err.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", err.Message)
	}
	
	// Quick errors should not have cause or suggestions by default
	if err.Cause != nil {
		t.Error("Quick errors should not have a cause by default")
	}
	
	if len(err.Suggestions) > 0 {
		t.Error("Quick errors should not have suggestions by default")
	}
}

func TestErrorBuildersChaining(t *testing.T) {
	// Test that error builders return proper GmanError instances that can be chained
	baseErr := fmt.Errorf("base error")
	
	err := OperationError("test", baseErr)
	err = err.WithSuggestion("Additional suggestion")
	
	if len(err.Suggestions) == 0 {
		t.Error("Expected suggestions to be present after chaining")
	}
	
	// Verify the suggestion was added
	if !containsSuggestion(err.Suggestions, "Additional suggestion") {
		t.Error("Expected chained suggestion to be present")
	}
}

// Helper function to check if a suggestion is present in the suggestions slice
func containsSuggestion(suggestions []string, target string) bool {
	for _, suggestion := range suggestions {
		if strings.Contains(suggestion, target) {
			return true
		}
	}
	return false
}