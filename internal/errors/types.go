package errors

import (
	"fmt"
	"strings"
)

// ErrorType represents different categories of errors in gman
type ErrorType string

const (
	// Repository-related errors
	ErrTypeRepoNotFound     ErrorType = "REPO_NOT_FOUND"
	ErrTypeNotGitRepo       ErrorType = "NOT_GIT_REPO"
	ErrTypeRepoAlreadyExists ErrorType = "REPO_ALREADY_EXISTS"
	
	// Git operation errors
	ErrTypeMergeConflict     ErrorType = "MERGE_CONFLICT"
	ErrTypeRemoteUnreachable ErrorType = "REMOTE_UNREACHABLE"
	ErrTypeWorkspaceNotClean ErrorType = "WORKSPACE_NOT_CLEAN"
	ErrTypeBranchNotFound    ErrorType = "BRANCH_NOT_FOUND"
	ErrTypeWorktreeExists    ErrorType = "WORKTREE_EXISTS"
	
	// Configuration errors
	ErrTypeConfigInvalid   ErrorType = "CONFIG_INVALID"
	ErrTypeConfigNotFound  ErrorType = "CONFIG_NOT_FOUND"
	ErrTypePermissionDenied ErrorType = "PERMISSION_DENIED"
	
	// External tool errors
	ErrTypeToolNotAvailable ErrorType = "TOOL_NOT_AVAILABLE"
	ErrTypeCommandFailed    ErrorType = "COMMAND_FAILED"
	
	// Network and connectivity errors
	ErrTypeNetworkTimeout ErrorType = "NETWORK_TIMEOUT"
	ErrTypeConnectFailed  ErrorType = "CONNECT_FAILED"
	
	// User input errors
	ErrTypeInvalidInput    ErrorType = "INVALID_INPUT"
	ErrTypeOperationCancelled ErrorType = "OPERATION_CANCELLED"
	
	// Internal errors
	ErrTypeInternal        ErrorType = "INTERNAL_ERROR"
	ErrTypeNotImplemented  ErrorType = "NOT_IMPLEMENTED"
)

// Severity represents the severity level of an error
type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// GmanError represents a structured error in the gman system
type GmanError struct {
	Type        ErrorType         `json:"type"`
	Message     string            `json:"message"`
	Cause       error             `json:"cause,omitempty"`
	Severity    Severity          `json:"severity"`
	Suggestions []string          `json:"suggestions,omitempty"`
	Context     map[string]string `json:"context,omitempty"`
	Code        string            `json:"code,omitempty"`
}

// Error implements the error interface
func (e *GmanError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Cause.Error())
	}
	return e.Message
}

// Unwrap returns the underlying cause error
func (e *GmanError) Unwrap() error {
	return e.Cause
}

// Is checks if the error is of a specific type
func (e *GmanError) Is(target error) bool {
	if ge, ok := target.(*GmanError); ok {
		return e.Type == ge.Type
	}
	return false
}

// WithCause adds a cause error to the GmanError
func (e *GmanError) WithCause(cause error) *GmanError {
	e.Cause = cause
	return e
}

// WithSuggestion adds a suggestion to the error
func (e *GmanError) WithSuggestion(suggestion string) *GmanError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// WithSuggestions adds multiple suggestions to the error
func (e *GmanError) WithSuggestions(suggestions ...string) *GmanError {
	e.Suggestions = append(e.Suggestions, suggestions...)
	return e
}

// WithContext adds context information to the error
func (e *GmanError) WithContext(key, value string) *GmanError {
	if e.Context == nil {
		e.Context = make(map[string]string)
	}
	e.Context[key] = value
	return e
}

// WithContextMap adds multiple context entries to the error
func (e *GmanError) WithContextMap(context map[string]string) *GmanError {
	if e.Context == nil {
		e.Context = make(map[string]string)
	}
	for k, v := range context {
		e.Context[k] = v
	}
	return e
}

// GetSuggestions returns formatted suggestions for display
func (e *GmanError) GetSuggestions() string {
	if len(e.Suggestions) == 0 {
		return ""
	}
	
	var result strings.Builder
	result.WriteString("Suggestions:\n")
	for i, suggestion := range e.Suggestions {
		result.WriteString(fmt.Sprintf("  %d. %s\n", i+1, suggestion))
	}
	return result.String()
}

// GetContext returns formatted context information for display
func (e *GmanError) GetContext() string {
	if len(e.Context) == 0 {
		return ""
	}
	
	var result strings.Builder
	result.WriteString("Context:\n")
	for key, value := range e.Context {
		result.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
	}
	return result.String()
}

// IsCritical returns true if the error is critical
func (e *GmanError) IsCritical() bool {
	return e.Severity == SeverityCritical
}

// IsRecoverable returns true if the error might be recoverable
func (e *GmanError) IsRecoverable() bool {
	switch e.Type {
	case ErrTypeNetworkTimeout, ErrTypeConnectFailed, ErrTypeRemoteUnreachable:
		return true
	case ErrTypeToolNotAvailable:
		return true
	case ErrTypeWorkspaceNotClean:
		return true
	default:
		return false
	}
}