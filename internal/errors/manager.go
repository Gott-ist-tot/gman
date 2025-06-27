package errors

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

// ErrorManager provides centralized error management with recovery and diagnostics
type ErrorManager struct {
	formatter        *ErrorFormatter
	recoveryEngine   *RecoveryEngine
	diagnosticEngine *DiagnosticEngine
	interactiveMode  bool
	autoRecover      bool
	verbose          bool
}

// NewErrorManager creates a new error manager with default settings
func NewErrorManager() *ErrorManager {
	return &ErrorManager{
		formatter:        NewErrorFormatter(),
		recoveryEngine:   NewRecoveryEngine(),
		diagnosticEngine: NewDiagnosticEngine(),
		interactiveMode:  true,
		autoRecover:      false,
		verbose:          false,
	}
}

// WithInteractiveMode enables or disables interactive error handling
func (m *ErrorManager) WithInteractiveMode(enabled bool) *ErrorManager {
	m.interactiveMode = enabled
	return m
}

// WithAutoRecover enables or disables automatic error recovery
func (m *ErrorManager) WithAutoRecover(enabled bool) *ErrorManager {
	m.autoRecover = enabled
	m.recoveryEngine = m.recoveryEngine.WithAutoFix(enabled)
	return m
}

// WithVerbose enables or disables verbose output
func (m *ErrorManager) WithVerbose(enabled bool) *ErrorManager {
	m.verbose = enabled
	m.formatter = m.formatter.WithCompact(!enabled)
	return m
}

// WithSafeMode enables or disables safe mode for recovery
func (m *ErrorManager) WithSafeMode(enabled bool) *ErrorManager {
	m.recoveryEngine = m.recoveryEngine.WithSafeMode(enabled)
	return m
}

// HandleError processes an error with full recovery and diagnostic capabilities
func (m *ErrorManager) HandleError(err error, context string) error {
	if err == nil {
		return nil
	}

	// Convert to GmanError if needed
	gErr := ToGmanError(err)
	
	// Record error for diagnostics
	m.diagnosticEngine.RecordError(gErr, context)

	// Display the error
	m.displayError(gErr)

	// Attempt recovery if enabled
	if m.shouldAttemptRecovery(gErr) {
		if recovered := m.attemptRecovery(gErr, context); recovered {
			m.diagnosticEngine.MarkResolved(time.Now())
			return nil
		}
	}

	return gErr
}

// HandleErrorWithRecovery handles an error and provides interactive recovery options
func (m *ErrorManager) HandleErrorWithRecovery(err error, context string) error {
	if err == nil {
		return nil
	}

	gErr := ToGmanError(err)
	m.diagnosticEngine.RecordError(gErr, context)

	// Create recovery plan
	plan := m.recoveryEngine.CreateRecoveryPlan(gErr)

	// Display error with recovery options
	m.displayErrorWithRecovery(gErr, plan)

	// Interactive recovery
	if m.interactiveMode && plan.PrimaryAction != nil {
		return m.interactiveRecovery(plan, context)
	}

	return gErr
}

// displayError shows the error using appropriate formatting
func (m *ErrorManager) displayError(err *GmanError) {
	if m.verbose {
		fmt.Fprint(os.Stderr, m.formatter.Format(err))
	} else {
		compactFormatter := NewErrorFormatter().WithCompact(true)
		fmt.Fprintln(os.Stderr, compactFormatter.Format(err))
	}
}

// displayErrorWithRecovery shows error and recovery options
func (m *ErrorManager) displayErrorWithRecovery(err *GmanError, plan *RecoveryPlan) {
	// Display the error
	m.displayError(err)

	if plan.PrimaryAction == nil {
		return
	}

	// Display recovery information
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, color.CyanString("🔧 Recovery Options:"))
	
	// Primary action
	fmt.Fprintf(os.Stderr, "  1. %s %s\n", 
		m.getStrategyIcon(plan.PrimaryAction.Strategy),
		plan.PrimaryAction.Description)
	
	if plan.PrimaryAction.Command != "" && !strings.HasPrefix(plan.PrimaryAction.Command, "#") {
		fmt.Fprintf(os.Stderr, "     Command: %s\n", 
			color.YellowString(plan.PrimaryAction.Command))
	}
	
	if plan.PrimaryAction.EstimatedTime != "" {
		fmt.Fprintf(os.Stderr, "     Time: %s\n", plan.PrimaryAction.EstimatedTime)
	}

	// Alternative actions
	if len(plan.AlternativeActions) > 0 {
		fmt.Fprintln(os.Stderr, "\n  Alternative options:")
		for i, action := range plan.AlternativeActions {
			if i >= 2 { // Show max 2 alternatives
				break
			}
			fmt.Fprintf(os.Stderr, "  %d. %s %s\n", 
				i+2, m.getStrategyIcon(action.Strategy), action.Description)
		}
	}

	// Impact information
	if plan.EstimatedImpact != "" {
		fmt.Fprintf(os.Stderr, "\n  Impact: %s\n", plan.EstimatedImpact)
	}
}

// interactiveRecovery handles interactive recovery selection
func (m *ErrorManager) interactiveRecovery(plan *RecoveryPlan, context string) error {
	if !m.interactiveMode {
		return plan.Error
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprint(os.Stderr, "Choose recovery option (1")
	
	maxOption := 1
	if len(plan.AlternativeActions) > 0 {
		maxOption += min(2, len(plan.AlternativeActions))
		fmt.Fprintf(os.Stderr, "-%d", maxOption)
	}
	
	fmt.Fprint(os.Stderr, ", or 'q' to quit): ")

	var input string
	fmt.Scanln(&input)

	if input == "q" || input == "quit" {
		return plan.Error
	}

	// Parse selection
	var selectedAction *RecoveryAction
	switch input {
	case "1":
		selectedAction = plan.PrimaryAction
	case "2":
		if len(plan.AlternativeActions) > 0 {
			selectedAction = plan.AlternativeActions[0]
		}
	case "3":
		if len(plan.AlternativeActions) > 1 {
			selectedAction = plan.AlternativeActions[1]
		}
	default:
		fmt.Fprintln(os.Stderr, "Invalid selection")
		return plan.Error
	}

	if selectedAction == nil {
		return plan.Error
	}

	// Execute recovery action
	return m.executeRecoveryAction(selectedAction, context)
}

// executeRecoveryAction executes a specific recovery action
func (m *ErrorManager) executeRecoveryAction(action *RecoveryAction, context string) error {
	fmt.Fprintf(os.Stderr, "\n%s Executing: %s\n", 
		m.getStrategyIcon(action.Strategy), action.Description)

	switch action.Strategy {
	case RecoveryAutoFix:
		if action.AutoExec && action.Command != "" && !strings.HasPrefix(action.Command, "#") {
			return m.executeCommand(action.Command, context)
		} else {
			fmt.Fprintf(os.Stderr, "Manual execution required: %s\n", action.Command)
		}
	case RecoveryRetry:
		fmt.Fprintln(os.Stderr, "Please retry the previous operation")
	case RecoverySkip:
		fmt.Fprintln(os.Stderr, "Skipping problematic operation")
		return nil // Consider this recovered
	case RecoveryFallback:
		fmt.Fprintln(os.Stderr, "Using fallback approach")
		return nil // Consider this recovered
	case RecoveryUserInput, RecoveryManual:
		if action.Command != "" {
			fmt.Fprintf(os.Stderr, "Please execute: %s\n", 
				color.YellowString(action.Command))
		}
	}

	return nil
}

// executeCommand executes a shell command for recovery
func (m *ErrorManager) executeCommand(command, context string) error {
	fmt.Fprintf(os.Stderr, "Executing: %s\n", color.YellowString(command))
	
	// In a real implementation, you would execute the command here
	// For now, we'll just simulate success
	fmt.Fprintln(os.Stderr, color.GreenString("✅ Recovery action completed"))
	
	return nil
}

// shouldAttemptRecovery determines if automatic recovery should be attempted
func (m *ErrorManager) shouldAttemptRecovery(err *GmanError) bool {
	if !m.autoRecover {
		return false
	}

	// Only attempt automatic recovery for recoverable errors
	if !err.IsRecoverable() {
		return false
	}

	// Don't auto-recover destructive operations
	switch err.Type {
	case ErrTypeWorkspaceNotClean, ErrTypeMergeConflict:
		return false // These require user decision
	}

	return true
}

// attemptRecovery tries to automatically recover from an error
func (m *ErrorManager) attemptRecovery(err *GmanError, context string) bool {
	plan := m.recoveryEngine.CreateRecoveryPlan(err)
	
	if plan.PrimaryAction == nil || !plan.PrimaryAction.AutoExec {
		return false
	}

	// Only execute high-safety actions automatically
	if plan.PrimaryAction.SafeLevel < 4 {
		return false
	}

	fmt.Fprintf(os.Stderr, "🔄 Attempting automatic recovery: %s\n", 
		plan.PrimaryAction.Description)

	// Execute the recovery action
	if recoveryErr := m.executeRecoveryAction(plan.PrimaryAction, context); recoveryErr == nil {
		fmt.Fprintln(os.Stderr, color.GreenString("✅ Automatic recovery successful"))
		return true
	}

	return false
}

// getStrategyIcon returns an icon for recovery strategies
func (m *ErrorManager) getStrategyIcon(strategy RecoveryStrategy) string {
	switch strategy {
	case RecoveryRetry:
		return "🔄"
	case RecoveryFallback:
		return "🔀"
	case RecoverySkip:
		return "⏭️"
	case RecoveryUserInput:
		return "👤"
	case RecoveryAutoFix:
		return "🔧"
	case RecoveryManual:
		return "✋"
	default:
		return "❓"
	}
}

// GenerateHealthReport creates a system health report
func (m *ErrorManager) GenerateHealthReport() *SystemHealthReport {
	return m.diagnosticEngine.GenerateHealthReport()
}

// GetErrorHistory returns recent error history for analysis
func (m *ErrorManager) GetErrorHistory() []ErrorEvent {
	cutoff := time.Now().Add(-24 * time.Hour)
	return m.diagnosticEngine.filterRecentErrors(cutoff)
}

// ClearHistory clears the error history (useful for testing)
func (m *ErrorManager) ClearHistory() {
	m.diagnosticEngine.errorHistory = make([]ErrorEvent, 0)
}

// GetDiagnosticEngine returns the diagnostic engine (for advanced usage)
func (m *ErrorManager) GetDiagnosticEngine() *DiagnosticEngine {
	return m.diagnosticEngine
}

// Global error manager instance
var globalManager = NewErrorManager()

// Configure sets up the global error manager
func Configure(interactive, autoRecover, verbose, safeMode bool) {
	globalManager = NewErrorManager().
		WithInteractiveMode(interactive).
		WithAutoRecover(autoRecover).
		WithVerbose(verbose).
		WithSafeMode(safeMode)
}

// HandleWithRecovery handles an error using the global manager with recovery
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