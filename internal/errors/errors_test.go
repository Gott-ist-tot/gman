package errors

import (
	"fmt"
	"strings"
	"testing"
)

func TestGmanError_Basic(t *testing.T) {
	err := NewRepoNotFoundError("/nonexistent/path")
	
	if err.Type != ErrTypeRepoNotFound {
		t.Errorf("Expected error type %s, got %s", ErrTypeRepoNotFound, err.Type)
	}
	
	// Severity field removed in simplified version
	if !IsCritical(err) {
		t.Error("Expected repo not found error to be critical")
	}
	
	if len(err.Suggestions) == 0 {
		t.Error("Expected suggestions to be provided")
	}
	
	// Context field removed in simplified version - path is in the message
	if !strings.Contains(err.Message, "/nonexistent/path") {
		t.Error("Expected path to be in error message")
	}
}

func TestGmanError_WithCause(t *testing.T) {
	baseErr := fmt.Errorf("underlying error")
	err := NewInternalError("test", "something went wrong").WithCause(baseErr)
	
	if err.Cause != baseErr {
		t.Error("Expected cause to be preserved")
	}
	
	if err.Unwrap() != baseErr {
		t.Error("Expected Unwrap to return the cause")
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
	err := NewRepoNotFoundError("/test/path")
	
	// Test formatting with simplified formatter
	formatter := NewErrorFormatter()
	
	// Test compact formatting
	compact := formatter.WithCompact(true).Format(err)
	if compact == "" {
		t.Error("Expected non-empty compact format")
	}
	
	// Test detailed formatting
	detailed := formatter.WithCompact(false).Format(err)
	if detailed == "" {
		t.Error("Expected non-empty detailed format")
	}
	
	// Compact format should be shorter than detailed
	if len(compact) >= len(detailed) {
		t.Error("Expected compact format to be shorter than detailed")
	}
}

func TestErrorTypeChecking(t *testing.T) {
	err := NewRepoNotFoundError("/test/path")
	
	if !IsType(err, ErrTypeRepoNotFound) {
		t.Error("Expected error to be identified as repo not found")
	}
	
	if IsType(err, ErrTypeMergeConflict) {
		t.Error("Expected error not to be identified as merge conflict")
	}
	
	if !IsCritical(err) {
		t.Error("Expected repo not found error to be critical")
	}
}

func TestRecoverableErrors(t *testing.T) {
	// Network errors should be recoverable
	networkErr := NewNetworkTimeoutError("test operation", "30s")
	if !IsRecoverable(networkErr) {
		t.Error("Expected network timeout to be recoverable")
	}
	
	// Tool availability errors should be recoverable
	toolErr := NewToolNotAvailableError("test-tool", "install instructions")
	if !IsRecoverable(toolErr) {
		t.Error("Expected tool not available error to be recoverable")
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
	
	// Test with existing GmanError
	existingErr := NewRepoNotFoundError("/test")
	converted := ToGmanError(existingErr)
	
	if converted != existingErr {
		t.Error("Expected existing GmanError to be returned unchanged")
	}
}