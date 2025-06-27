package errors

import (
	"fmt"
	"testing"
)

func TestGmanError_Basic(t *testing.T) {
	err := NewRepoNotFoundError("/nonexistent/path")
	
	if err.Type != ErrTypeRepoNotFound {
		t.Errorf("Expected error type %s, got %s", ErrTypeRepoNotFound, err.Type)
	}
	
	if err.Severity != SeverityCritical {
		t.Errorf("Expected severity %s, got %s", SeverityCritical, err.Severity)
	}
	
	if len(err.Suggestions) == 0 {
		t.Error("Expected suggestions to be provided")
	}
	
	if err.Context["path"] != "/nonexistent/path" {
		t.Error("Expected path context to be set")
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
	err := NewGmanError(ErrTypeCommandFailed, "test error")
	
	err.WithContext("key1", "value1")
	err.WithContextMap(map[string]string{
		"key2": "value2",
		"key3": "value3",
	})
	
	if len(err.Context) != 3 {
		t.Errorf("Expected 3 context entries, got %d", len(err.Context))
	}
	
	if err.Context["key1"] != "value1" {
		t.Error("Expected key1 context to be set")
	}
}

func TestErrorFormatting(t *testing.T) {
	err := NewRepoNotFoundError("/test/path")
	
	// Test simple formatting
	simple := FormatSimple(err)
	if simple == "" {
		t.Error("Expected non-empty simple format")
	}
	
	// Test detailed formatting
	detailed := FormatDetailed(err)
	if detailed == "" {
		t.Error("Expected non-empty detailed format")
	}
	
	// Simple format should be shorter than detailed
	if len(simple) >= len(detailed) {
		t.Error("Expected simple format to be shorter than detailed")
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