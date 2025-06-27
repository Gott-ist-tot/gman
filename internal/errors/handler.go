package errors

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// ErrorHandler provides centralized error handling for the CLI
type ErrorHandler struct {
	formatter *ErrorFormatter
	verbose   bool
}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		formatter: NewErrorFormatter(),
		verbose:   false,
	}
}

// WithVerbose enables verbose error output
func (h *ErrorHandler) WithVerbose(verbose bool) *ErrorHandler {
	h.verbose = verbose
	return h
}

// HandleError processes and displays an error appropriately
func (h *ErrorHandler) HandleError(err error) {
	if err == nil {
		return
	}

	// Check if it's a GmanError
	if gErr, ok := As(err); ok {
		h.handleGmanError(gErr)
	} else {
		h.handleStandardError(err)
	}
}

// handleGmanError handles GmanError instances with appropriate formatting
func (h *ErrorHandler) handleGmanError(err *GmanError) {
	if h.verbose {
		fmt.Fprint(os.Stderr, h.formatter.Format(err))
	} else {
		// Show compact format for non-verbose mode
		compactFormatter := NewErrorFormatter().WithCompact(true)
		fmt.Fprintln(os.Stderr, compactFormatter.Format(err))
		
		// Show suggestions for critical errors
		if err.IsCritical() && len(err.Suggestions) > 0 {
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, h.formatter.formatSuggestions(err.Suggestions))
		}
	}
}

// handleStandardError handles standard Go errors
func (h *ErrorHandler) handleStandardError(err error) {
	if h.verbose {
		fmt.Fprintf(os.Stderr, "❌ ERROR: %s\n", err.Error())
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	}
}

// ExitCode returns the appropriate exit code for an error
func (h *ErrorHandler) ExitCode(err error) int {
	if err == nil {
		return 0
	}

	if gErr, ok := As(err); ok {
		switch gErr.Severity {
		case SeverityCritical:
			return 2
		case SeverityError:
			return 1
		case SeverityWarning:
			return 0 // Warnings don't cause exit failure
		case SeverityInfo:
			return 0
		default:
			return 1
		}
	}

	return 1 // Standard errors return 1
}

// WrapCobraCommand wraps a cobra command to provide enhanced error handling
func WrapCobraCommand(cmd *cobra.Command) {
	originalRunE := cmd.RunE
	originalRun := cmd.Run

	errorHandler := NewErrorHandler()

	// Check for verbose flag
	if cmd.Flags().Changed("verbose") {
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			errorHandler = errorHandler.WithVerbose(true)
		}
	}

	if originalRunE != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			err := originalRunE(cmd, args)
			if err != nil {
				errorHandler.HandleError(err)
				os.Exit(errorHandler.ExitCode(err))
			}
			return nil
		}
		cmd.Run = nil
	} else if originalRun != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			originalRun(cmd, args)
			return nil
		}
		cmd.Run = nil
	}
}

// Global error handler instance
var globalHandler = NewErrorHandler()

// SetVerbose sets the global error handler to verbose mode
func SetVerbose(verbose bool) {
	globalHandler = globalHandler.WithVerbose(verbose)
}

// Handle processes an error using the global handler
func Handle(err error) {
	globalHandler.HandleError(err)
}

// Exit processes an error and exits with appropriate code
func Exit(err error) {
	if err != nil {
		globalHandler.HandleError(err)
		os.Exit(globalHandler.ExitCode(err))
	}
}

// Fatal processes an error and exits (convenience function)
func Fatal(err error) {
	Exit(err)
}

// HandleAndExit handles an error and exits if it's not nil (convenience function)
func HandleAndExit(err error) {
	if err != nil {
		Exit(err)
	}
}

// CreateUserFriendlyError converts a standard error to a user-friendly GmanError
func CreateUserFriendlyError(err error, context string) *GmanError {
	if gErr, ok := As(err); ok {
		return gErr
	}

	// Analyze the error message to provide better categorization
	errMsg := err.Error()
	
	// Network-related errors
	if containsAny(errMsg, []string{"timeout", "connection", "network", "unreachable"}) {
		return NewNetworkTimeoutError(context, "unknown").WithCause(err)
	}
	
	// Permission-related errors
	if containsAny(errMsg, []string{"permission", "access denied", "forbidden"}) {
		return NewPermissionDeniedError(context).WithCause(err)
	}
	
	// Git-related errors
	if containsAny(errMsg, []string{"not a git repository", ".git"}) {
		return NewNotGitRepoError(context).WithCause(err)
	}
	
	if containsAny(errMsg, []string{"merge conflict", "conflict"}) {
		return NewMergeConflictError(context).WithCause(err)
	}
	
	// File/path errors
	if containsAny(errMsg, []string{"no such file", "not found", "does not exist"}) {
		return NewRepoNotFoundError(context).WithCause(err)
	}
	
	// Command errors
	if containsAny(errMsg, []string{"command not found", "executable file not found"}) {
		return NewToolNotAvailableError("unknown", "").WithCause(err)
	}
	
	// Default to internal error
	return NewInternalError(context, errMsg).WithCause(err)
}

// containsAny checks if a string contains any of the given substrings
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}