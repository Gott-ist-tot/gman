package errors

import (
	"fmt"
	"os"
)

// ErrorManager provides simple error display for CLI tool
type ErrorManager struct {
	verbose bool
}

// NewErrorManager creates a new simplified error manager
func NewErrorManager() *ErrorManager {
	return &ErrorManager{
		verbose: false,
	}
}

// WithVerbose enables or disables verbose output
func (m *ErrorManager) WithVerbose(enabled bool) *ErrorManager {
	m.verbose = enabled
	return m
}

// WithInteractiveMode enables or disables interactive error handling (no-op for compatibility)
func (m *ErrorManager) WithInteractiveMode(enabled bool) *ErrorManager {
	// Simplified - no interactive mode needed for CLI tool
	return m
}

// WithAutoRecover enables or disables automatic error recovery (no-op for compatibility)
func (m *ErrorManager) WithAutoRecover(enabled bool) *ErrorManager {
	// Simplified - no auto recovery needed for CLI tool
	return m
}

// WithSafeMode enables or disables safe mode for recovery (no-op for compatibility)
func (m *ErrorManager) WithSafeMode(enabled bool) *ErrorManager {
	// Simplified - no safe mode needed for CLI tool
	return m
}

// HandleError processes an error with simple display
func (m *ErrorManager) HandleError(err error, context string) error {
	if err == nil {
		return nil
	}

	// Convert to GmanError if needed for better display
	gErr := ToGmanError(err)
	
	// Display the error
	m.displayError(gErr, context)

	return gErr
}

// HandleErrorWithRecovery handles an error using simple display (no recovery)
func (m *ErrorManager) HandleErrorWithRecovery(err error, context string) error {
	return m.HandleError(err, context)
}

// displayError shows the error with suggestions
func (m *ErrorManager) displayError(err *GmanError, context string) {
	// Show error with context if provided
	if context != "" {
		fmt.Fprintf(os.Stderr, "Error in %s: %s\n", context, err.Error())
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	}
	
	// Show suggestions if available
	if len(err.Suggestions) > 0 {
		fmt.Fprintln(os.Stderr, "\nSuggestions:")
		for i, suggestion := range err.Suggestions {
			fmt.Fprintf(os.Stderr, "  %d. %s\n", i+1, suggestion)
		}
	}
	
	// Show verbose details if enabled
	if m.verbose && err.Cause != nil {
		fmt.Fprintf(os.Stderr, "Caused by: %s\n", err.Cause.Error())
	}
}

// GenerateHealthReport creates a simple system health report (for compatibility)
func (m *ErrorManager) GenerateHealthReport() *SystemHealthReport {
	return &SystemHealthReport{
		OverallHealth:  "healthy",
		ErrorPatterns:  []ErrorPattern{},
	}
}

// GetErrorHistory returns empty history (simplified)
func (m *ErrorManager) GetErrorHistory() []ErrorEvent {
	return []ErrorEvent{}
}

// ClearHistory is a no-op (for compatibility)
func (m *ErrorManager) ClearHistory() {
	// No history to clear in simplified version
}

// SystemHealthReport represents basic system health (simplified)
type SystemHealthReport struct {
	OverallHealth string         `json:"overall_health"`
	ErrorPatterns []ErrorPattern `json:"error_patterns"`
}

// ErrorPattern represents an error pattern (simplified)
type ErrorPattern struct {
	ErrorType string `json:"error_type"`
	Frequency int    `json:"frequency"`
}

// ErrorEvent represents an error event (simplified)
type ErrorEvent struct {
	Timestamp string `json:"timestamp"`
	ErrorType string `json:"error_type"`
	Message   string `json:"message"`
}

// Global error manager instance
var globalManager = NewErrorManager()

// Configure sets up the global error manager (simplified)
func Configure(interactive, autoRecover, verbose, safeMode bool) {
	globalManager = NewErrorManager().WithVerbose(verbose)
}

// HandleWithRecovery handles an error using the global manager (simplified)
func HandleWithRecovery(err error, context string) error {
	return globalManager.HandleErrorWithRecovery(err, context)
}

// GetGlobalManager returns the global error manager
func GetGlobalManager() *ErrorManager {
	return globalManager
}

// GenerateGlobalHealthReport generates a health report using the global manager
func GenerateGlobalHealthReport() *SystemHealthReport {
	return globalManager.GenerateHealthReport()
}