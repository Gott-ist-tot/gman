package errors

import (
	"fmt"
	"strings"
	"testing"
)

func TestGmanError_Basic(t *testing.T) {
	// Test using new builder pattern
	err := NotFoundError("repository", "/nonexistent/path")
	
	if err.Type != ErrTypeRepoNotFound {
		t.Errorf("Expected error type %s, got %s", ErrTypeRepoNotFound, err.Type)
	}
	
	// Severity field removed in simplified version
	if !IsCritical(err) {
		t.Error("Expected repo not found error to be critical")
	}
	
	// Builder pattern doesn't automatically add suggestions, but they can be added
	err = err.WithSuggestion("Check the repository path")
	if len(err.Suggestions) == 0 {
		t.Error("Expected suggestions to be provided after adding them")
	}
	
	// Context field removed in simplified version - path is in the message
	if !strings.Contains(err.Message, "/nonexistent/path") {
		t.Error("Expected path to be in error message")
	}
	
	// Also test legacy factory function for backward compatibility
	legacyErr := NewRepoNotFoundError("/legacy/path")
	if legacyErr.Type != ErrTypeRepoNotFound {
		t.Error("Legacy factory function should still work")
	}
	if len(legacyErr.Suggestions) == 0 {
		t.Error("Legacy factory should include suggestions")
	}
}

func TestGmanError_WithCause(t *testing.T) {
	baseErr := fmt.Errorf("underlying error")
	// Test using new builder pattern
	err := InternalError("test", "something went wrong", baseErr)
	
	if err.Cause != baseErr {
		t.Error("Expected cause to be preserved")
	}
	
	if err.Unwrap() != baseErr {
		t.Error("Expected Unwrap to return the cause")
	}
	
	// Also test legacy factory function
	legacyErr := NewInternalError("test", "something went wrong").WithCause(baseErr)
	if legacyErr.Cause != baseErr {
		t.Error("Legacy factory should preserve cause when chained")
	}
}

func TestGmanError_Suggestions(t *testing.T) {
	err := NewGmanError(ErrTypeInvalidInput, "test error")
	
	err.WithSuggestion("First suggestion")
	err.WithSuggestions("Second suggestion", "Third suggestion")
	
	if len(err.Suggestions) != 3 {
		t.Errorf("Expected 3 suggestions, got %d", len(err.Suggestions))
	}
}

func TestGmanError_Context(t *testing.T) {
	// Context functionality removed in simplified version
	// This test is now a placeholder to maintain test structure
	err := NewGmanError(ErrTypeCommandFailed, "test error")
	
	if err.Type != ErrTypeCommandFailed {
		t.Error("Expected error type to be preserved")
	}
	
	if err.Message != "test error" {
		t.Error("Expected error message to be preserved")
	}
}

func TestErrorFormatting(t *testing.T) {
	// Test with both builder pattern and legacy factory
	builderErr := NotFoundError("repository", "/test/path")
	legacyErr := NewRepoNotFoundError("/test/path")
	
	// Test formatting with simplified formatter
	formatter := NewErrorFormatter()
	
	// Test compact formatting for builder error
	compact := formatter.WithCompact(true).Format(builderErr)
	if compact == "" {
		t.Error("Expected non-empty compact format for builder error")
	}
	
	// Test detailed formatting for legacy error
	detailed := formatter.WithCompact(false).Format(legacyErr)
	if detailed == "" {
		t.Error("Expected non-empty detailed format for legacy error")
	}
	
	// Both should format properly
	legacyCompact := formatter.WithCompact(true).Format(legacyErr)
	if legacyCompact == "" {
		t.Error("Expected legacy errors to format properly")
	}
}

func TestErrorTypeChecking(t *testing.T) {
	// Test with both builder pattern and legacy factory
	builderErr := NotFoundError("repository", "/test/path")
	legacyErr := NewRepoNotFoundError("/test/path")
	
	// Test builder error
	if !IsType(builderErr, ErrTypeRepoNotFound) {
		t.Error("Expected builder error to be identified as repo not found")
	}
	
	if IsType(builderErr, ErrTypeMergeConflict) {
		t.Error("Expected builder error not to be identified as merge conflict")
	}
	
	if !IsCritical(builderErr) {
		t.Error("Expected builder repo not found error to be critical")
	}
	
	// Test legacy error
	if !IsType(legacyErr, ErrTypeRepoNotFound) {
		t.Error("Expected legacy error to be identified as repo not found")
	}
	
	if !IsCritical(legacyErr) {
		t.Error("Expected legacy repo not found error to be critical")
	}
}

func TestRecoverableErrors(t *testing.T) {
	// Network errors should be recoverable
	networkErr := NewNetworkTimeoutError("test operation", "30s")
	if !IsRecoverable(networkErr) {
		t.Error("Expected network timeout to be recoverable")
	}
	
	// Test with builder pattern
	builderNetworkErr := NetworkError("test operation", fmt.Errorf("timeout"))
	if !IsRecoverable(builderNetworkErr) {
		t.Error("Expected builder network error to be recoverable")
	}
	
	// Tool availability errors should be recoverable
	toolErr := NewToolNotAvailableError("test-tool", "install instructions")
	if !IsRecoverable(toolErr) {
		t.Error("Expected tool not available error to be recoverable")
	}
	
	// Test with builder pattern
	builderToolErr := ExternalToolError("test-tool", "test operation", fmt.Errorf("not found"))
	if !IsRecoverable(builderToolErr) {
		t.Error("Expected builder tool error to be recoverable")
	}
	
	// Config errors should not be recoverable
	configErr := NewConfigNotFoundError("/test/config")
	if IsRecoverable(configErr) {
		t.Error("Expected config not found error to not be recoverable")
	}
}

func TestRetryableError(t *testing.T) {
	baseErr := NewNetworkTimeoutError("test", "30s")
	retryErr := WithRetry(baseErr, 3, "5s")
	
	if retryErr.MaxRetries != 3 {
		t.Errorf("Expected max retries to be 3, got %d", retryErr.MaxRetries)
	}
	
	if retryErr.Delay != "5s" {
		t.Errorf("Expected delay to be 5s, got %s", retryErr.Delay)
	}
	
	// Test retryable checking
	if rErr, ok := IsRetryable(retryErr); !ok {
		t.Error("Expected error to be identified as retryable")
	} else if rErr.MaxRetries != 3 {
		t.Error("Expected retryable error to preserve retry info")
	}
}

func TestToGmanError(t *testing.T) {
	// Test with standard error
	stdErr := fmt.Errorf("standard error")
	gErr := ToGmanError(stdErr)
	
	if gErr.Type != ErrTypeInternal {
		t.Error("Expected standard error to be converted to internal error")
	}
	
	if gErr.Cause != stdErr {
		t.Error("Expected cause to be preserved")
	}
	
	// Test with existing GmanError from legacy factory
	existingErr := NewRepoNotFoundError("/test")
	converted := ToGmanError(existingErr)
	
	if converted != existingErr {
		t.Error("Expected existing GmanError to be returned unchanged")
	}
	
	// Test with existing GmanError from builder
	builderErr := NotFoundError("repository", "/test")
	convertedBuilder := ToGmanError(builderErr)
	
	if convertedBuilder != builderErr {
		t.Error("Expected existing builder GmanError to be returned unchanged")
	}
}