package errors

import (
	"testing"
)

func TestRecoveryEngine_CreateRecoveryPlan(t *testing.T) {
	engine := NewRecoveryEngine()
	
	// Test repo not found error
	err := NewRepoNotFoundError("/test/path")
	plan := engine.CreateRecoveryPlan(err)
	
	if plan.Error != err {
		t.Error("Expected plan to reference the original error")
	}
	
	if plan.PrimaryAction == nil {
		t.Error("Expected primary action to be provided")
	}
	
	if len(plan.AlternativeActions) == 0 {
		t.Error("Expected alternative actions to be provided")
	}
	
	if plan.EstimatedImpact == "" {
		t.Error("Expected impact estimation to be provided")
	}
}

func TestRecoveryEngine_SafeMode(t *testing.T) {
	engine := NewRecoveryEngine().WithSafeMode(true)
	
	// Create an error that would normally have low-safety actions
	err := NewWorkspaceNotCleanError("/test/repo")
	plan := engine.CreateRecoveryPlan(err)
	
	// Verify that only safe actions are included
	if plan.PrimaryAction != nil && plan.PrimaryAction.SafeLevel < 3 {
		t.Error("Expected safe mode to filter out unsafe actions")
	}
	
	for _, action := range plan.AlternativeActions {
		if action.SafeLevel < 3 {
			t.Error("Expected safe mode to filter out unsafe alternative actions")
		}
	}
}

func TestRecoveryEngine_AutoFix(t *testing.T) {
	engine := NewRecoveryEngine().WithAutoFix(true)
	
	err := NewNotGitRepoError("/test/path")
	plan := engine.CreateRecoveryPlan(err)
	
	// Should have auto-executable actions when auto-fix is enabled
	hasAutoExec := false
	if plan.PrimaryAction != nil && plan.PrimaryAction.AutoExec {
		hasAutoExec = true
	}
	
	for _, action := range plan.AlternativeActions {
		if action.AutoExec {
			hasAutoExec = true
			break
		}
	}
	
	if !hasAutoExec {
		t.Error("Expected auto-fix mode to enable auto-executable actions")
	}
}

func TestRecoveryAction_Properties(t *testing.T) {
	actions := []*RecoveryAction{
		{
			Strategy:      RecoveryRetry,
			Description:   "Test retry action",
			SafeLevel:     5,
			EstimatedTime: "30s",
		},
		{
			Strategy:      RecoveryAutoFix,
			Description:   "Test auto-fix action",
			Command:       "git init",
			AutoExec:      true,
			SafeLevel:     4,
			EstimatedTime: "5s",
		},
	}
	
	for _, action := range actions {
		if action.Strategy == "" {
			t.Error("Expected strategy to be set")
		}
		
		if action.Description == "" {
			t.Error("Expected description to be set")
		}
		
		if action.SafeLevel < 1 || action.SafeLevel > 5 {
			t.Error("Expected safe level to be between 1 and 5")
		}
	}
}

func TestRecoveryPlan_DiagnosticInfo(t *testing.T) {
	engine := NewRecoveryEngine()
	err := NewNetworkTimeoutError("test operation", "30s")
	plan := engine.CreateRecoveryPlan(err)
	
	// Should have diagnostic information
	if len(plan.DiagnosticInfo) == 0 {
		t.Error("Expected diagnostic information to be populated")
	}
	
	// Should have specific keys
	expectedKeys := []string{"error_type", "severity", "timestamp"}
	for _, key := range expectedKeys {
		if _, exists := plan.DiagnosticInfo[key]; !exists {
			t.Errorf("Expected diagnostic info to contain key: %s", key)
		}
	}
}

func TestRecoveryEngine_ErrorTypeHandling(t *testing.T) {
	engine := NewRecoveryEngine()
	
	testCases := []struct {
		name      string
		errorFunc func() *GmanError
		expectActions bool
	}{
		{
			name:      "RepoNotFound",
			errorFunc: func() *GmanError { return NewRepoNotFoundError("/test") },
			expectActions: true,
		},
		{
			name:      "MergeConflict",
			errorFunc: func() *GmanError { return NewMergeConflictError("/test") },
			expectActions: true,
		},
		{
			name:      "ToolNotAvailable",
			errorFunc: func() *GmanError { return NewToolNotAvailableError("test-tool", "") },
			expectActions: true,
		},
		{
			name:      "NetworkTimeout",
			errorFunc: func() *GmanError { return NewNetworkTimeoutError("test", "30s") },
			expectActions: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.errorFunc()
			plan := engine.CreateRecoveryPlan(err)
			
			if tc.expectActions {
				if plan.PrimaryAction == nil {
					t.Errorf("Expected primary action for error type %s", err.Type)
				}
			}
		})
	}
}

func TestDiagnosticEngine_RecordError(t *testing.T) {
	engine := NewDiagnosticEngine()
	
	err := NewRepoNotFoundError("/test/path")
	context := "test operation"
	
	// Record the error
	engine.RecordError(err, context)
	
	// Verify it was recorded
	if len(engine.errorHistory) != 1 {
		t.Errorf("Expected 1 error in history, got %d", len(engine.errorHistory))
	}
	
	event := engine.errorHistory[0]
	if event.Error != err {
		t.Error("Expected recorded error to match original")
	}
	
	if event.Context != context {
		t.Error("Expected recorded context to match original")
	}
	
	if event.Resolved {
		t.Error("Expected newly recorded error to not be resolved")
	}
}

func TestDiagnosticEngine_AnalyzePatterns(t *testing.T) {
	engine := NewDiagnosticEngine()
	
	// Record multiple errors of the same type
	for i := 0; i < 3; i++ {
		err := NewRepoNotFoundError("/test/path")
		engine.RecordError(err, "test")
	}
	
	// Record one error of different type
	err2 := NewNetworkTimeoutError("test", "30s")
	engine.RecordError(err2, "test")
	
	patterns := engine.AnalyzePatterns()
	
	if len(patterns) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(patterns))
	}
	
	// First pattern should be the most frequent (RepoNotFound)
	if patterns[0].ErrorType != ErrTypeRepoNotFound {
		t.Error("Expected most frequent pattern to be RepoNotFound")
	}
	
	if patterns[0].Frequency != 3 {
		t.Errorf("Expected frequency of 3, got %d", patterns[0].Frequency)
	}
}

func TestDiagnosticEngine_HealthCalculation(t *testing.T) {
	engine := NewDiagnosticEngine()
	
	// Test excellent health (no errors)
	report := engine.GenerateHealthReport()
	if report.OverallHealth != HealthExcellent {
		t.Error("Expected excellent health with no errors")
	}
	
	// Add some critical errors
	for i := 0; i < 6; i++ {
		err := NewConfigInvalidError("test")
		engine.RecordError(err, "test")
	}
	
	report = engine.GenerateHealthReport()
	if report.OverallHealth != HealthCritical {
		t.Error("Expected critical health with many critical errors")
	}
}

func TestErrorManager_Integration(t *testing.T) {
	manager := NewErrorManager().
		WithInteractiveMode(false).
		WithAutoRecover(true).
		WithVerbose(false)
	
	// Test handling a recoverable error
	err := NewNetworkTimeoutError("test operation", "30s")
	result := manager.HandleError(err, "test context")
	
	// Should have recorded the error
	history := manager.GetErrorHistory()
	if len(history) == 0 {
		t.Error("Expected error to be recorded in history")
	}
	
	// The error should still be returned since we're not in interactive mode
	if result == nil {
		t.Error("Expected error to be returned in non-interactive mode")
	}
}

func TestErrorManager_HealthReporting(t *testing.T) {
	manager := NewErrorManager()
	
	// Add some test errors
	manager.HandleError(NewRepoNotFoundError("/test1"), "test")
	manager.HandleError(NewRepoNotFoundError("/test2"), "test")
	manager.HandleError(NewNetworkTimeoutError("sync", "30s"), "test")
	
	report := manager.GenerateHealthReport()
	
	if report.OverallHealth == HealthExcellent {
		t.Error("Expected health to be degraded with recorded errors")
	}
	
	if len(report.ErrorPatterns) == 0 {
		t.Error("Expected error patterns to be identified")
	}
	
	if len(report.RecommendedActions) == 0 {
		t.Error("Expected recommended actions to be provided")
	}
}

func TestErrorManager_ClearHistory(t *testing.T) {
	manager := NewErrorManager()
	
	// Add some errors
	manager.HandleError(NewRepoNotFoundError("/test"), "test")
	
	// Verify errors were recorded
	if len(manager.GetErrorHistory()) == 0 {
		t.Error("Expected errors to be recorded")
	}
	
	// Clear history
	manager.ClearHistory()
	
	// Verify history is empty
	if len(manager.GetErrorHistory()) != 0 {
		t.Error("Expected history to be cleared")
	}
}