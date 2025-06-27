package errors

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// DiagnosticPattern represents a pattern of errors that can be analyzed
type DiagnosticPattern struct {
	ErrorType     ErrorType `json:"error_type"`
	Frequency     int       `json:"frequency"`
	LastOccurrence time.Time `json:"last_occurrence"`
	Contexts      []string  `json:"contexts"`
	Suggestions   []string  `json:"suggestions"`
}

// SystemHealthReport provides an overview of error patterns and system health
type SystemHealthReport struct {
	GeneratedAt       time.Time            `json:"generated_at"`
	OverallHealth     HealthStatus         `json:"overall_health"`
	ErrorPatterns     []DiagnosticPattern  `json:"error_patterns"`
	PreventiveTips    []string             `json:"preventive_tips"`
	SystemMetrics     map[string]string    `json:"system_metrics"`
	RecommendedActions []string            `json:"recommended_actions"`
}

// HealthStatus represents the overall health of the system
type HealthStatus string

const (
	HealthExcellent HealthStatus = "EXCELLENT"
	HealthGood      HealthStatus = "GOOD"
	HealthFair      HealthStatus = "FAIR"
	HealthPoor      HealthStatus = "POOR"
	HealthCritical  HealthStatus = "CRITICAL"
)

// DiagnosticEngine analyzes error patterns and provides system health insights
type DiagnosticEngine struct {
	errorHistory    []ErrorEvent
	maxHistorySize  int
	analysisWindow  time.Duration
}

// ErrorEvent represents a single error occurrence for analysis
type ErrorEvent struct {
	Error     *GmanError `json:"error"`
	Timestamp time.Time  `json:"timestamp"`
	Context   string     `json:"context"`
	Resolved  bool       `json:"resolved"`
}

// NewDiagnosticEngine creates a new diagnostic engine
func NewDiagnosticEngine() *DiagnosticEngine {
	return &DiagnosticEngine{
		errorHistory:   make([]ErrorEvent, 0),
		maxHistorySize: 1000,
		analysisWindow: 24 * time.Hour, // Analyze last 24 hours by default
	}
}

// WithHistorySize sets the maximum number of error events to keep
func (d *DiagnosticEngine) WithHistorySize(size int) *DiagnosticEngine {
	d.maxHistorySize = size
	return d
}

// WithAnalysisWindow sets the time window for analysis
func (d *DiagnosticEngine) WithAnalysisWindow(window time.Duration) *DiagnosticEngine {
	d.analysisWindow = window
	return d
}

// RecordError adds an error to the diagnostic history
func (d *DiagnosticEngine) RecordError(err *GmanError, context string) {
	event := ErrorEvent{
		Error:     err,
		Timestamp: time.Now(),
		Context:   context,
		Resolved:  false,
	}

	d.errorHistory = append(d.errorHistory, event)

	// Trim history if it exceeds max size
	if len(d.errorHistory) > d.maxHistorySize {
		d.errorHistory = d.errorHistory[1:]
	}
}

// MarkResolved marks an error as resolved for analysis purposes
func (d *DiagnosticEngine) MarkResolved(timestamp time.Time) {
	for i := range d.errorHistory {
		if d.errorHistory[i].Timestamp.Equal(timestamp) {
			d.errorHistory[i].Resolved = true
			break
		}
	}
}

// AnalyzePatterns identifies common error patterns and their characteristics
func (d *DiagnosticEngine) AnalyzePatterns() []DiagnosticPattern {
	cutoff := time.Now().Add(-d.analysisWindow)
	recentErrors := d.filterRecentErrors(cutoff)

	// Group errors by type
	patterns := make(map[ErrorType]*DiagnosticPattern)

	for _, event := range recentErrors {
		errorType := event.Error.Type
		
		if pattern, exists := patterns[errorType]; exists {
			pattern.Frequency++
			if event.Timestamp.After(pattern.LastOccurrence) {
				pattern.LastOccurrence = event.Timestamp
			}
			pattern.Contexts = d.mergeContexts(pattern.Contexts, event.Context)
		} else {
			patterns[errorType] = &DiagnosticPattern{
				ErrorType:     errorType,
				Frequency:     1,
				LastOccurrence: event.Timestamp,
				Contexts:      []string{event.Context},
				Suggestions:   d.generatePatternSuggestions(errorType, []ErrorEvent{event}),
			}
		}
	}

	// Convert map to slice and sort by frequency
	result := make([]DiagnosticPattern, 0, len(patterns))
	for _, pattern := range patterns {
		result = append(result, *pattern)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Frequency > result[j].Frequency
	})

	return result
}

// GenerateHealthReport creates a comprehensive system health report
func (d *DiagnosticEngine) GenerateHealthReport() *SystemHealthReport {
	patterns := d.AnalyzePatterns()
	
	report := &SystemHealthReport{
		GeneratedAt:    time.Now(),
		ErrorPatterns:  patterns,
		SystemMetrics:  d.collectSystemMetrics(),
	}

	// Determine overall health
	report.OverallHealth = d.calculateOverallHealth(patterns)

	// Generate preventive tips
	report.PreventiveTips = d.generatePreventiveTips(patterns)

	// Generate recommended actions
	report.RecommendedActions = d.generateRecommendedActions(patterns, report.OverallHealth)

	return report
}

// calculateOverallHealth determines system health based on error patterns
func (d *DiagnosticEngine) calculateOverallHealth(patterns []DiagnosticPattern) HealthStatus {
	if len(patterns) == 0 {
		return HealthExcellent
	}

	// Count critical and error-level issues
	criticalCount := 0
	errorCount := 0
	totalFrequency := 0

	for _, pattern := range patterns {
		totalFrequency += pattern.Frequency
		
		// Determine severity based on error type
		switch pattern.ErrorType {
		case ErrTypeConfigInvalid, ErrTypeRepoNotFound:
			criticalCount += pattern.Frequency
		case ErrTypeMergeConflict, ErrTypePermissionDenied:
			errorCount += pattern.Frequency
		}
	}

	// Health calculation logic
	if criticalCount > 5 {
		return HealthCritical
	} else if criticalCount > 2 || errorCount > 10 {
		return HealthPoor
	} else if totalFrequency > 20 {
		return HealthFair
	} else if totalFrequency > 5 {
		return HealthGood
	}

	return HealthExcellent
}

// generatePreventiveTips creates tips to prevent common errors
func (d *DiagnosticEngine) generatePreventiveTips(patterns []DiagnosticPattern) []string {
	tips := make([]string, 0)
	seenTypes := make(map[ErrorType]bool)

	for _, pattern := range patterns {
		if !seenTypes[pattern.ErrorType] {
			seenTypes[pattern.ErrorType] = true
			tips = append(tips, d.getPreventiveTipsForType(pattern.ErrorType)...)
		}
	}

	// Add general tips
	tips = append(tips,
		"Regularly run 'gman tools check' to ensure optimal configuration",
		"Keep repositories clean with periodic 'gman work status --extended'",
		"Use 'gman setup discover' to maintain accurate repository configuration",
	)

	return tips
}

// generateRecommendedActions creates action items based on health status
func (d *DiagnosticEngine) generateRecommendedActions(patterns []DiagnosticPattern, health HealthStatus) []string {
	actions := make([]string, 0)

	switch health {
	case HealthCritical:
		actions = append(actions,
			"URGENT: Address critical configuration issues immediately",
			"Run 'gman setup --reset' to restore default configuration",
			"Review and fix repository paths with 'gman repo list'",
		)
	case HealthPoor:
		actions = append(actions,
			"Review and resolve frequent error patterns",
			"Consider running system maintenance commands",
			"Check repository health with 'gman work status --extended'",
		)
	case HealthFair:
		actions = append(actions,
			"Monitor error patterns for trends",
			"Optimize workflow to reduce common errors",
		)
	case HealthGood:
		actions = append(actions,
			"Continue current practices",
			"Consider periodic health checks",
		)
	case HealthExcellent:
		actions = append(actions,
			"System is operating optimally",
			"No immediate action required",
		)
	}

	// Add pattern-specific actions
	for _, pattern := range patterns[:min(3, len(patterns))] { // Top 3 patterns
		if pattern.Frequency > 2 {
			actions = append(actions, 
				fmt.Sprintf("Address recurring %s errors (occurred %d times)", 
					pattern.ErrorType, pattern.Frequency))
		}
	}

	return actions
}

// collectSystemMetrics gathers relevant system metrics
func (d *DiagnosticEngine) collectSystemMetrics() map[string]string {
	metrics := make(map[string]string)
	
	cutoff := time.Now().Add(-d.analysisWindow)
	recentErrors := d.filterRecentErrors(cutoff)
	
	metrics["total_errors_24h"] = fmt.Sprintf("%d", len(recentErrors))
	metrics["analysis_window"] = d.analysisWindow.String()
	metrics["history_size"] = fmt.Sprintf("%d", len(d.errorHistory))
	
	// Calculate resolution rate
	resolved := 0
	for _, event := range recentErrors {
		if event.Resolved {
			resolved++
		}
	}
	
	if len(recentErrors) > 0 {
		resolutionRate := float64(resolved) / float64(len(recentErrors)) * 100
		metrics["resolution_rate"] = fmt.Sprintf("%.1f%%", resolutionRate)
	} else {
		metrics["resolution_rate"] = "N/A"
	}
	
	// Most common error type
	if len(recentErrors) > 0 {
		typeCounts := make(map[ErrorType]int)
		for _, event := range recentErrors {
			typeCounts[event.Error.Type]++
		}
		
		maxCount := 0
		var mostCommon ErrorType
		for errorType, count := range typeCounts {
			if count > maxCount {
				maxCount = count
				mostCommon = errorType
			}
		}
		
		metrics["most_common_error"] = string(mostCommon)
	}
	
	return metrics
}

// generatePatternSuggestions creates specific suggestions for error patterns
func (d *DiagnosticEngine) generatePatternSuggestions(errorType ErrorType, events []ErrorEvent) []string {
	suggestions := make([]string, 0)

	switch errorType {
	case ErrTypeRepoNotFound:
		suggestions = append(suggestions,
			"Regularly verify repository paths with 'gman repo list'",
			"Use absolute paths when adding repositories",
			"Remove invalid repositories promptly to avoid confusion",
		)
	case ErrTypeNetworkTimeout:
		suggestions = append(suggestions,
			"Check network stability and proxy configuration",
			"Consider increasing timeout values for slow connections",
			"Use offline mode when network is unreliable",
		)
	case ErrTypeMergeConflict:
		suggestions = append(suggestions,
			"Keep feature branches up to date with main branch",
			"Use smaller, more frequent commits to reduce conflict scope",
			"Consider using merge tools for easier conflict resolution",
		)
	case ErrTypeToolNotAvailable:
		suggestions = append(suggestions,
			"Install recommended tools with 'gman tools check --fix'",
			"Set up tool installation scripts for team consistency",
			"Document tool requirements for new team members",
		)
	}

	return suggestions
}

// getPreventiveTipsForType returns preventive tips for specific error types
func (d *DiagnosticEngine) getPreventiveTipsForType(errorType ErrorType) []string {
	switch errorType {
	case ErrTypeRepoNotFound:
		return []string{
			"Use 'gman setup discover' to automatically find and configure repositories",
			"Verify repository paths before adding them to configuration",
		}
	case ErrTypeNetworkTimeout:
		return []string{
			"Test network connectivity before bulk operations",
			"Use 'gman work sync --dry-run' to preview network operations",
		}
	case ErrTypeMergeConflict:
		return []string{
			"Fetch and merge changes frequently to avoid large conflicts",
			"Use feature branches and keep them short-lived",
		}
	case ErrTypeToolNotAvailable:
		return []string{
			"Run 'gman tools check' after system updates",
			"Include tool installation in onboarding documentation",
		}
	default:
		return []string{}
	}
}

// filterRecentErrors returns errors within the analysis window
func (d *DiagnosticEngine) filterRecentErrors(cutoff time.Time) []ErrorEvent {
	filtered := make([]ErrorEvent, 0)
	for _, event := range d.errorHistory {
		if event.Timestamp.After(cutoff) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// mergeContexts combines context strings, avoiding duplicates
func (d *DiagnosticEngine) mergeContexts(existing []string, newContext string) []string {
	for _, ctx := range existing {
		if ctx == newContext {
			return existing
		}
	}
	return append(existing, newContext)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// FormatHealthReport creates a human-readable health report
func FormatHealthReport(report *SystemHealthReport) string {
	var result strings.Builder
	
	result.WriteString(fmt.Sprintf("🏥 System Health Report - %s\n", 
		report.GeneratedAt.Format("2006-01-02 15:04:05")))
	result.WriteString(strings.Repeat("=", 50) + "\n\n")
	
	// Overall health
	healthIcon := getHealthIcon(report.OverallHealth)
	result.WriteString(fmt.Sprintf("%s Overall Health: %s\n\n", healthIcon, report.OverallHealth))
	
	// System metrics
	result.WriteString("📊 System Metrics:\n")
	for key, value := range report.SystemMetrics {
		result.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
	}
	result.WriteString("\n")
	
	// Error patterns
	if len(report.ErrorPatterns) > 0 {
		result.WriteString("🔍 Error Patterns (Last 24h):\n")
		for i, pattern := range report.ErrorPatterns {
			if i >= 5 { // Show top 5 patterns
				break
			}
			result.WriteString(fmt.Sprintf("  %d. %s: %d occurrences\n", 
				i+1, pattern.ErrorType, pattern.Frequency))
		}
		result.WriteString("\n")
	}
	
	// Recommended actions
	if len(report.RecommendedActions) > 0 {
		result.WriteString("🎯 Recommended Actions:\n")
		for i, action := range report.RecommendedActions {
			result.WriteString(fmt.Sprintf("  %d. %s\n", i+1, action))
		}
		result.WriteString("\n")
	}
	
	// Preventive tips
	if len(report.PreventiveTips) > 0 {
		result.WriteString("💡 Preventive Tips:\n")
		for i, tip := range report.PreventiveTips[:min(3, len(report.PreventiveTips))] {
			result.WriteString(fmt.Sprintf("  %d. %s\n", i+1, tip))
		}
	}
	
	return result.String()
}

// getHealthIcon returns an appropriate icon for health status
func getHealthIcon(health HealthStatus) string {
	switch health {
	case HealthExcellent:
		return "🟢"
	case HealthGood:
		return "🟡"
	case HealthFair:
		return "🟠"
	case HealthPoor:
		return "🔴"
	case HealthCritical:
		return "🚨"
	default:
		return "❓"
	}
}