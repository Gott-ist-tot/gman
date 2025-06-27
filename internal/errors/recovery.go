package errors

import (
	"fmt"
	"strings"
	"time"
)

// RecoveryStrategy represents different approaches to error recovery
type RecoveryStrategy string

const (
	RecoveryRetry     RecoveryStrategy = "RETRY"
	RecoveryFallback  RecoveryStrategy = "FALLBACK"
	RecoverySkip      RecoveryStrategy = "SKIP"
	RecoveryUserInput RecoveryStrategy = "USER_INPUT"
	RecoveryAutoFix   RecoveryStrategy = "AUTO_FIX"
	RecoveryManual    RecoveryStrategy = "MANUAL"
)

// RecoveryAction represents a specific action that can be taken to recover from an error
type RecoveryAction struct {
	Strategy    RecoveryStrategy  `json:"strategy"`
	Description string            `json:"description"`
	Command     string            `json:"command,omitempty"`
	AutoExec    bool              `json:"auto_exec"`
	SafeLevel   int               `json:"safe_level"` // 1-5, where 5 is safest
	EstimatedTime string          `json:"estimated_time,omitempty"`
	Prerequisites []string        `json:"prerequisites,omitempty"`
}

// RecoveryPlan contains multiple recovery options for an error
type RecoveryPlan struct {
	Error           *GmanError       `json:"error"`
	PrimaryAction   *RecoveryAction  `json:"primary_action"`
	AlternativeActions []*RecoveryAction `json:"alternative_actions"`
	DiagnosticInfo  map[string]string `json:"diagnostic_info"`
	EstimatedImpact string           `json:"estimated_impact"`
}

// RecoveryEngine provides intelligent error recovery recommendations
type RecoveryEngine struct {
	safeMode bool
	autoFix  bool
}

// NewRecoveryEngine creates a new recovery engine
func NewRecoveryEngine() *RecoveryEngine {
	return &RecoveryEngine{
		safeMode: true,
		autoFix:  false,
	}
}

// WithSafeMode enables or disables safe mode (conservative recovery)
func (r *RecoveryEngine) WithSafeMode(enabled bool) *RecoveryEngine {
	r.safeMode = enabled
	return r
}

// WithAutoFix enables or disables automatic fixes
func (r *RecoveryEngine) WithAutoFix(enabled bool) *RecoveryEngine {
	r.autoFix = enabled
	return r
}

// CreateRecoveryPlan generates a comprehensive recovery plan for an error
func (r *RecoveryEngine) CreateRecoveryPlan(err *GmanError) *RecoveryPlan {
	plan := &RecoveryPlan{
		Error:          err,
		DiagnosticInfo: make(map[string]string),
	}

	// Add diagnostic information
	r.addDiagnosticInfo(plan)

	// Generate recovery actions based on error type
	actions := r.generateRecoveryActions(err)
	if len(actions) > 0 {
		plan.PrimaryAction = actions[0]
		if len(actions) > 1 {
			plan.AlternativeActions = actions[1:]
		}
	}

	// Estimate impact
	plan.EstimatedImpact = r.estimateImpact(err)

	return plan
}

// generateRecoveryActions creates recovery actions based on error type
func (r *RecoveryEngine) generateRecoveryActions(err *GmanError) []*RecoveryAction {
	var actions []*RecoveryAction

	switch err.Type {
	case ErrTypeRepoNotFound:
		actions = r.handleRepoNotFound(err)
	case ErrTypeNotGitRepo:
		actions = r.handleNotGitRepo(err)
	case ErrTypeMergeConflict:
		actions = r.handleMergeConflict(err)
	case ErrTypeRemoteUnreachable:
		actions = r.handleRemoteUnreachable(err)
	case ErrTypeWorkspaceNotClean:
		actions = r.handleWorkspaceNotClean(err)
	case ErrTypeBranchNotFound:
		actions = r.handleBranchNotFound(err)
	case ErrTypeToolNotAvailable:
		actions = r.handleToolNotAvailable(err)
	case ErrTypeNetworkTimeout:
		actions = r.handleNetworkTimeout(err)
	case ErrTypeConfigInvalid:
		actions = r.handleConfigInvalid(err)
	case ErrTypePermissionDenied:
		actions = r.handlePermissionDenied(err)
	default:
		actions = r.handleGeneric(err)
	}

	// Filter actions based on safe mode
	if r.safeMode {
		actions = r.filterSafeActions(actions)
	}

	return actions
}

// handleRepoNotFound creates recovery actions for repository not found errors
func (r *RecoveryEngine) handleRepoNotFound(err *GmanError) []*RecoveryAction {
	repoPath := err.Context["path"]
	
	return []*RecoveryAction{
		{
			Strategy:    RecoveryUserInput,
			Description: "Remove invalid repository from configuration",
			Command:     fmt.Sprintf("gman repo remove %s", getAliasFromPath(repoPath)),
			SafeLevel:   5,
			EstimatedTime: "5 seconds",
		},
		{
			Strategy:    RecoveryUserInput,
			Description: "Update repository path in configuration",
			Command:     "gman repo list # then update the correct path",
			SafeLevel:   4,
			EstimatedTime: "1 minute",
		},
		{
			Strategy:    RecoveryManual,
			Description: "Create the repository directory and initialize Git",
			Command:     fmt.Sprintf("mkdir -p %s && cd %s && git init", repoPath, repoPath),
			SafeLevel:   3,
			EstimatedTime: "2 minutes",
		},
	}
}

// handleNotGitRepo creates recovery actions for non-git directory errors
func (r *RecoveryEngine) handleNotGitRepo(err *GmanError) []*RecoveryAction {
	repoPath := err.Context["path"]
	
	return []*RecoveryAction{
		{
			Strategy:    RecoveryAutoFix,
			Description: "Initialize Git repository in the directory",
			Command:     fmt.Sprintf("git -C %s init", repoPath),
			AutoExec:    r.autoFix,
			SafeLevel:   4,
			EstimatedTime: "5 seconds",
		},
		{
			Strategy:    RecoveryUserInput,
			Description: "Clone an existing repository to this location",
			Command:     fmt.Sprintf("git clone <repository-url> %s", repoPath),
			SafeLevel:   5,
			EstimatedTime: "1-5 minutes",
		},
		{
			Strategy:    RecoveryUserInput,
			Description: "Remove from gman configuration",
			Command:     fmt.Sprintf("gman repo remove %s", getAliasFromPath(repoPath)),
			SafeLevel:   5,
			EstimatedTime: "5 seconds",
		},
	}
}

// handleMergeConflict creates recovery actions for merge conflicts
func (r *RecoveryEngine) handleMergeConflict(err *GmanError) []*RecoveryAction {
	repoPath := err.Context["repository"]
	
	return []*RecoveryAction{
		{
			Strategy:    RecoveryManual,
			Description: "Open merge tool to resolve conflicts interactively",
			Command:     fmt.Sprintf("cd %s && git mergetool", repoPath),
			SafeLevel:   4,
			EstimatedTime: "5-30 minutes",
			Prerequisites: []string{"Merge tool configured (git config merge.tool)"},
		},
		{
			Strategy:    RecoveryUserInput,
			Description: "Abort the merge and return to previous state",
			Command:     fmt.Sprintf("git -C %s merge --abort", repoPath),
			SafeLevel:   5,
			EstimatedTime: "5 seconds",
		},
		{
			Strategy:    RecoveryManual,
			Description: "Resolve conflicts manually in editor",
			Command:     fmt.Sprintf("cd %s && git status # Edit conflicted files, then git add .", repoPath),
			SafeLevel:   3,
			EstimatedTime: "10-60 minutes",
		},
	}
}

// handleRemoteUnreachable creates recovery actions for remote connectivity issues
func (r *RecoveryEngine) handleRemoteUnreachable(err *GmanError) []*RecoveryAction {
	repoPath := err.Context["repository"]
	remote := err.Context["remote"]
	
	return []*RecoveryAction{
		{
			Strategy:    RecoveryRetry,
			Description: "Retry the operation (network issues are often temporary)",
			Command:     "# Retry the previous operation",
			SafeLevel:   5,
			EstimatedTime: "30 seconds",
		},
		{
			Strategy:    RecoveryManual,
			Description: "Check and update remote URL",
			Command:     fmt.Sprintf("git -C %s remote -v # Then: git remote set-url %s <new-url>", repoPath, remote),
			SafeLevel:   4,
			EstimatedTime: "2 minutes",
		},
		{
			Strategy:    RecoverySkip,
			Description: "Skip remote operations and work locally",
			Command:     "# Continue with local operations only",
			SafeLevel:   5,
			EstimatedTime: "Immediate",
		},
	}
}

// handleWorkspaceNotClean creates recovery actions for dirty workspace
func (r *RecoveryEngine) handleWorkspaceNotClean(err *GmanError) []*RecoveryAction {
	repoPath := err.Context["repository"]
	
	return []*RecoveryAction{
		{
			Strategy:    RecoveryAutoFix,
			Description: "Stash uncommitted changes temporarily",
			Command:     fmt.Sprintf("git -C %s stash push -m 'Auto-stash by gman'", repoPath),
			AutoExec:    r.autoFix,
			SafeLevel:   4,
			EstimatedTime: "5 seconds",
		},
		{
			Strategy:    RecoveryUserInput,
			Description: "Commit the changes with a message",
			Command:     fmt.Sprintf("git -C %s add . && git commit -m 'WIP: uncommitted changes'", repoPath),
			SafeLevel:   4,
			EstimatedTime: "30 seconds",
		},
		{
			Strategy:    RecoveryManual,
			Description: "Review and selectively commit changes",
			Command:     fmt.Sprintf("cd %s && git status && git add -p", repoPath),
			SafeLevel:   5,
			EstimatedTime: "2-10 minutes",
		},
		{
			Strategy:    RecoveryManual,
			Description: "Discard all uncommitted changes (DESTRUCTIVE)",
			Command:     fmt.Sprintf("git -C %s checkout -- .", repoPath),
			SafeLevel:   1,
			EstimatedTime: "5 seconds",
		},
	}
}

// handleBranchNotFound creates recovery actions for missing branches
func (r *RecoveryEngine) handleBranchNotFound(err *GmanError) []*RecoveryAction {
	repoPath := err.Context["repository"]
	branch := err.Context["branch"]
	
	return []*RecoveryAction{
		{
			Strategy:    RecoveryAutoFix,
			Description: "Create the branch from current HEAD",
			Command:     fmt.Sprintf("git -C %s checkout -b %s", repoPath, branch),
			AutoExec:    r.autoFix,
			SafeLevel:   4,
			EstimatedTime: "5 seconds",
		},
		{
			Strategy:    RecoveryUserInput,
			Description: "Create branch from specific commit or tag",
			Command:     fmt.Sprintf("git -C %s checkout -b %s <commit-hash>", repoPath, branch),
			SafeLevel:   3,
			EstimatedTime: "30 seconds",
		},
		{
			Strategy:    RecoveryFallback,
			Description: "Switch to the default branch instead",
			Command:     fmt.Sprintf("git -C %s checkout main || git checkout master", repoPath),
			SafeLevel:   5,
			EstimatedTime: "5 seconds",
		},
	}
}

// handleToolNotAvailable creates recovery actions for missing tools
func (r *RecoveryEngine) handleToolNotAvailable(err *GmanError) []*RecoveryAction {
	tool := err.Context["tool"]
	
	actions := []*RecoveryAction{
		{
			Strategy:    RecoveryFallback,
			Description: "Continue with reduced functionality (tool will be skipped)",
			Command:     "# Continue operation without the tool",
			SafeLevel:   5,
			EstimatedTime: "Immediate",
		},
		{
			Strategy:    RecoveryUserInput,
			Description: "Install the missing tool",
			Command:     fmt.Sprintf("gman tools check --fix # or install %s manually", tool),
			SafeLevel:   4,
			EstimatedTime: "1-5 minutes",
		},
	}

	return actions
}

// handleNetworkTimeout creates recovery actions for network timeouts
func (r *RecoveryEngine) handleNetworkTimeout(err *GmanError) []*RecoveryAction {
	operation := err.Context["operation"]
	
	return []*RecoveryAction{
		{
			Strategy:    RecoveryRetry,
			Description: "Retry with longer timeout",
			Command:     fmt.Sprintf("# Retry %s with extended timeout", operation),
			SafeLevel:   5,
			EstimatedTime: "1-2 minutes",
		},
		{
			Strategy:    RecoverySkip,
			Description: "Skip network-dependent operations",
			Command:     "# Continue with offline mode",
			SafeLevel:   4,
			EstimatedTime: "Immediate",
		},
		{
			Strategy:    RecoveryManual,
			Description: "Check network connectivity and proxy settings",
			Command:     "ping 8.8.8.8 # Check internet, verify proxy settings",
			SafeLevel:   5,
			EstimatedTime: "2 minutes",
		},
	}
}

// handleConfigInvalid creates recovery actions for configuration errors
func (r *RecoveryEngine) handleConfigInvalid(err *GmanError) []*RecoveryAction {
	return []*RecoveryAction{
		{
			Strategy:    RecoveryAutoFix,
			Description: "Reset configuration to defaults",
			Command:     "gman setup --reset",
			AutoExec:    r.autoFix,
			SafeLevel:   3,
			EstimatedTime: "30 seconds",
		},
		{
			Strategy:    RecoveryUserInput,
			Description: "Run setup wizard to recreate configuration",
			Command:     "gman setup",
			SafeLevel:   5,
			EstimatedTime: "2-5 minutes",
		},
		{
			Strategy:    RecoveryManual,
			Description: "Edit configuration file manually",
			Command:     "edit ~/.config/gman/config.yml",
			SafeLevel:   2,
			EstimatedTime: "5-15 minutes",
		},
	}
}

// handlePermissionDenied creates recovery actions for permission errors
func (r *RecoveryEngine) handlePermissionDenied(err *GmanError) []*RecoveryAction {
	resource := err.Context["resource"]
	
	return []*RecoveryAction{
		{
			Strategy:    RecoveryManual,
			Description: "Fix file/directory permissions",
			Command:     fmt.Sprintf("chmod u+rw %s # or chown if needed", resource),
			SafeLevel:   3,
			EstimatedTime: "1 minute",
		},
		{
			Strategy:    RecoveryUserInput,
			Description: "Run with elevated permissions",
			Command:     "sudo gman <command>",
			SafeLevel:   2,
			EstimatedTime: "30 seconds",
		},
		{
			Strategy:    RecoverySkip,
			Description: "Skip operations requiring this resource",
			Command:     "# Continue without accessing the restricted resource",
			SafeLevel:   4,
			EstimatedTime: "Immediate",
		},
	}
}

// handleGeneric creates generic recovery actions for unknown errors
func (r *RecoveryEngine) handleGeneric(err *GmanError) []*RecoveryAction {
	return []*RecoveryAction{
		{
			Strategy:    RecoveryRetry,
			Description: "Retry the operation",
			Command:     "# Retry the previous command",
			SafeLevel:   4,
			EstimatedTime: "30 seconds",
		},
		{
			Strategy:    RecoveryManual,
			Description: "Check system logs and documentation",
			Command:     "gman --help # or check documentation",
			SafeLevel:   5,
			EstimatedTime: "5 minutes",
		},
	}
}

// filterSafeActions removes actions with low safety levels in safe mode
func (r *RecoveryEngine) filterSafeActions(actions []*RecoveryAction) []*RecoveryAction {
	filtered := make([]*RecoveryAction, 0)
	
	for _, action := range actions {
		if action.SafeLevel >= 3 { // Only include moderately safe or safer actions
			filtered = append(filtered, action)
		}
	}
	
	return filtered
}

// addDiagnosticInfo adds contextual diagnostic information to the recovery plan
func (r *RecoveryEngine) addDiagnosticInfo(plan *RecoveryPlan) {
	err := plan.Error
	
	plan.DiagnosticInfo["error_type"] = string(err.Type)
	plan.DiagnosticInfo["severity"] = err.Severity.String()
	plan.DiagnosticInfo["timestamp"] = time.Now().Format(time.RFC3339)
	plan.DiagnosticInfo["recoverable"] = fmt.Sprintf("%t", err.IsRecoverable())
	
	// Add error-specific diagnostic info
	switch err.Type {
	case ErrTypeRemoteUnreachable, ErrTypeNetworkTimeout:
		plan.DiagnosticInfo["network_dependent"] = "true"
		plan.DiagnosticInfo["retry_recommended"] = "true"
	case ErrTypeToolNotAvailable:
		plan.DiagnosticInfo["degraded_functionality"] = "true"
		plan.DiagnosticInfo["fallback_available"] = "true"
	case ErrTypeMergeConflict, ErrTypeWorkspaceNotClean:
		plan.DiagnosticInfo["requires_user_decision"] = "true"
		plan.DiagnosticInfo["data_at_risk"] = "true"
	}
}

// estimateImpact provides an assessment of the error's impact
func (r *RecoveryEngine) estimateImpact(err *GmanError) string {
	switch err.Severity {
	case SeverityCritical:
		return "High - Operation cannot continue"
	case SeverityError:
		return "Medium - Current operation failed but system is stable"
	case SeverityWarning:
		return "Low - Reduced functionality but operation can continue"
	case SeverityInfo:
		return "Minimal - Informational only"
	default:
		return "Unknown impact level"
	}
}

// getAliasFromPath attempts to extract alias from repository path
func getAliasFromPath(path string) string {
	// This is a simplified implementation
	// In practice, you'd look up the alias from the configuration
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "unknown"
}