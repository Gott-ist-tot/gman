// Package errors provides structured error handling for gman
package errors

import (
	"errors"
	"fmt"
)

// Common error variables for easy comparison
var (
	ErrRepoNotFound     = &GmanError{Type: ErrTypeRepoNotFound}
	ErrNotGitRepo       = &GmanError{Type: ErrTypeNotGitRepo}
	ErrMergeConflict    = &GmanError{Type: ErrTypeMergeConflict}
	ErrRemoteUnreachable = &GmanError{Type: ErrTypeRemoteUnreachable}
	ErrWorkspaceNotClean = &GmanError{Type: ErrTypeWorkspaceNotClean}
	ErrBranchNotFound   = &GmanError{Type: ErrTypeBranchNotFound}
	ErrConfigInvalid    = &GmanError{Type: ErrTypeConfigInvalid}
	ErrToolNotAvailable = &GmanError{Type: ErrTypeToolNotAvailable}
)

// Wrap wraps a standard error with a GmanError
func Wrap(err error, errorType ErrorType, message string) *GmanError {
	return NewGmanError(errorType, message).WithCause(err)
}

// Wrapf wraps a standard error with a formatted message
func Wrapf(err error, errorType ErrorType, format string, args ...interface{}) *GmanError {
	message := fmt.Sprintf(format, args...)
	return NewGmanError(errorType, message).WithCause(err)
}

// Is checks if an error is of a specific type (supports both GmanError and standard errors)
func Is(err error, target error) bool {
	return errors.Is(err, target)
}

// As attempts to extract a GmanError from an error chain
func As(err error) (*GmanError, bool) {
	var gErr *GmanError
	if errors.As(err, &gErr) {
		return gErr, true
	}
	return nil, false
}

// IsType checks if an error is of a specific GmanError type
func IsType(err error, errorType ErrorType) bool {
	if gErr, ok := As(err); ok {
		return gErr.Type == errorType
	}
	return false
}

// GetType returns the error type of a GmanError, or empty string for other errors
func GetType(err error) ErrorType {
	if gErr, ok := As(err); ok {
		return gErr.Type
	}
	return ""
}

// GetSeverity returns the severity of a GmanError (simplified - always returns SeverityError)
func GetSeverity(err error) Severity {
	// Simplified - severity removed from GmanError
	return SeverityError
}

// IsRecoverable checks if an error might be recoverable
func IsRecoverable(err error) bool {
	if gErr, ok := As(err); ok {
		return gErr.IsRecoverable()
	}
	return false
}

// IsCritical checks if an error is critical (simplified - based on error type)
func IsCritical(err error) bool {
	if gErr, ok := As(err); ok {
		// Simplified - determine criticality based on error type
		switch gErr.Type {
		case ErrTypeConfigNotFound, ErrTypeRepoNotFound, ErrTypeInternal:
			return true
		default:
			return false
		}
	}
	return false
}

// GetSuggestions returns suggestions from a GmanError, or empty slice for other errors
func GetSuggestions(err error) []string {
	if gErr, ok := As(err); ok {
		return gErr.Suggestions
	}
	return nil
}

// ToGmanError converts any error to a GmanError
// If it's already a GmanError, returns it unchanged
// Otherwise wraps it as an internal error
func ToGmanError(err error) *GmanError {
	if gErr, ok := As(err); ok {
		return gErr
	}
	
	return NewInternalError("unknown", err.Error()).WithCause(err)
}

// CombineErrors combines multiple errors into a single GmanError
func CombineErrors(errors []error, message string) *GmanError {
	if len(errors) == 0 {
		return nil
	}
	
	if len(errors) == 1 {
		return ToGmanError(errors[0])
	}
	
	// Create a combined error
	combined := NewGmanError(ErrTypeInternal, message)
	
	// Add each error as context
	for i, err := range errors {
		key := fmt.Sprintf("error_%d", i+1)
		combined.WithContext(key, err.Error())
	}
	
	// Add suggestions based on the first error if it's a GmanError
	if gErr, ok := As(errors[0]); ok {
		combined.WithSuggestions(gErr.Suggestions...)
	}
	
	return combined
}

// SafeExecute executes a function and converts any panic to a GmanError
func SafeExecute(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = NewInternalError("panic", fmt.Sprintf("Panic occurred: %v", r))
		}
	}()
	
	return fn()
}

// RetryableError marks an error as retryable (simplified)
type RetryableError struct {
	*GmanError
	MaxRetries int
	Delay      string
}

// NewRetryableError creates a retryable error
func NewRetryableError(base *GmanError, maxRetries int, delay string) *RetryableError {
	return &RetryableError{
		GmanError:  base,
		MaxRetries: maxRetries,
		Delay:      delay,
	}
}

// Error implements the error interface for RetryableError
func (r *RetryableError) Error() string {
	return r.GmanError.Error()
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) (*RetryableError, bool) {
	if rErr, ok := err.(*RetryableError); ok {
		return rErr, true
	}
	return nil, false
}

// WithRetry adds retry information to a GmanError
func WithRetry(err *GmanError, maxRetries int, delay string) *RetryableError {
	return NewRetryableError(err, maxRetries, delay)
}