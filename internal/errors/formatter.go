package errors

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

// DisplayMode determines how errors are formatted and displayed
type DisplayMode string

const (
	DisplayModeCompact     DisplayMode = "compact"
	DisplayModeDetailed    DisplayMode = "detailed"
	DisplayModeInteractive DisplayMode = "interactive"
	DisplayModeJSON        DisplayMode = "json"
	DisplayModeTable       DisplayMode = "table"
)

// ErrorDisplayConfig contains configuration for error display formatting
type ErrorDisplayConfig struct {
	Mode           DisplayMode
	ShowTimestamp  bool
	ShowContext    bool
	ShowSuggestions bool
	ShowRecovery   bool
	ColorEnabled   bool
	MaxWidth       int
	IndentSize     int
	ShowIcons      bool
	Verbose        bool
}

// DefaultDisplayConfig returns a sensible default configuration
func DefaultDisplayConfig() *ErrorDisplayConfig {
	return &ErrorDisplayConfig{
		Mode:           DisplayModeDetailed,
		ShowTimestamp:  true,
		ShowContext:    true,
		ShowSuggestions: true,
		ShowRecovery:   true,
		ColorEnabled:   true,
		MaxWidth:       80,
		IndentSize:     2,
		ShowIcons:      true,
		Verbose:        false,
	}
}

// CompactDisplayConfig returns configuration for compact error display
func CompactDisplayConfig() *ErrorDisplayConfig {
	return &ErrorDisplayConfig{
		Mode:           DisplayModeCompact,
		ShowTimestamp:  false,
		ShowContext:    false,
		ShowSuggestions: false,
		ShowRecovery:   false,
		ColorEnabled:   true,
		MaxWidth:       120,
		IndentSize:     0,
		ShowIcons:      false,
		Verbose:        false,
	}
}

// ErrorFormatter provides different formatting options for errors (Legacy)
type ErrorFormatter struct {
	useColor    bool
	showContext bool
	showCode    bool
	compact     bool
}

// EnhancedErrorFormatter provides sophisticated error formatting capabilities
type EnhancedErrorFormatter struct {
	config *ErrorDisplayConfig
}

// NewErrorFormatter creates a new error formatter with default settings (Legacy)
func NewErrorFormatter() *ErrorFormatter {
	return &ErrorFormatter{
		useColor:    true,
		showContext: true,
		showCode:    false,
		compact:     false,
	}
}

// NewEnhancedErrorFormatter creates a new enhanced error formatter
func NewEnhancedErrorFormatter(config *ErrorDisplayConfig) *EnhancedErrorFormatter {
	if config == nil {
		config = DefaultDisplayConfig()
	}
	return &EnhancedErrorFormatter{
		config: config,
	}
}

// WithColor enables or disables colored output
func (f *ErrorFormatter) WithColor(enabled bool) *ErrorFormatter {
	f.useColor = enabled
	return f
}

// WithContext enables or disables context display
func (f *ErrorFormatter) WithContext(enabled bool) *ErrorFormatter {
	f.showContext = enabled
	return f
}

// WithCode enables or disables error code display
func (f *ErrorFormatter) WithCode(enabled bool) *ErrorFormatter {
	f.showCode = enabled
	return f
}

// WithCompact enables or disables compact formatting
func (f *ErrorFormatter) WithCompact(enabled bool) *ErrorFormatter {
	f.compact = enabled
	return f
}

// Format formats a GmanError for display
func (f *ErrorFormatter) Format(err *GmanError) string {
	if f.compact {
		return f.formatCompact(err)
	}
	return f.formatDetailed(err)
}

// formatCompact formats the error in a compact, single-line format
func (f *ErrorFormatter) formatCompact(err *GmanError) string {
	icon := f.getSeverityIcon(err.Severity)
	severityStr := f.formatSeverity(err.Severity)
	
	var result strings.Builder
	result.WriteString(fmt.Sprintf("%s %s: %s", icon, severityStr, err.Message))
	
	if f.showCode && err.Code != "" {
		result.WriteString(fmt.Sprintf(" [%s]", err.Code))
	}
	
	return result.String()
}

// formatDetailed formats the error with full details
func (f *ErrorFormatter) formatDetailed(err *GmanError) string {
	var result strings.Builder
	
	// Header with severity and main message
	icon := f.getSeverityIcon(err.Severity)
	severityStr := f.formatSeverity(err.Severity)
	
	result.WriteString(fmt.Sprintf("%s %s: %s\n", icon, severityStr, err.Message))
	
	// Error code if available
	if f.showCode && err.Code != "" {
		result.WriteString(fmt.Sprintf("Code: %s\n", f.formatCode(err.Code)))
	}
	
	// Underlying cause if available
	if err.Cause != nil {
		result.WriteString(fmt.Sprintf("Cause: %s\n", f.formatCause(err.Cause.Error())))
	}
	
	// Context information
	if f.showContext && len(err.Context) > 0 {
		result.WriteString("\n")
		result.WriteString(f.formatContext(err.Context))
	}
	
	// Suggestions
	if len(err.Suggestions) > 0 {
		result.WriteString("\n")
		result.WriteString(f.formatSuggestions(err.Suggestions))
	}
	
	return result.String()
}

// getSeverityIcon returns an icon for the severity level
func (f *ErrorFormatter) getSeverityIcon(severity Severity) string {
	switch severity {
	case SeverityInfo:
		return "ℹ️"
	case SeverityWarning:
		return "⚠️"
	case SeverityError:
		return "❌"
	case SeverityCritical:
		return "🚨"
	default:
		return "❓"
	}
}

// formatSeverity formats the severity level with optional color
func (f *ErrorFormatter) formatSeverity(severity Severity) string {
	severityStr := severity.String()
	
	if !f.useColor {
		return severityStr
	}
	
	switch severity {
	case SeverityInfo:
		return color.CyanString(severityStr)
	case SeverityWarning:
		return color.YellowString(severityStr)
	case SeverityError:
		return color.RedString(severityStr)
	case SeverityCritical:
		return color.New(color.FgRed, color.Bold).Sprint(severityStr)
	default:
		return severityStr
	}
}

// formatCode formats the error code
func (f *ErrorFormatter) formatCode(code string) string {
	if f.useColor {
		return color.CyanString(code)
	}
	return code
}

// formatCause formats the underlying cause
func (f *ErrorFormatter) formatCause(cause string) string {
	if f.useColor {
		return color.YellowString(cause)
	}
	return cause
}

// formatContext formats context information
func (f *ErrorFormatter) formatContext(context map[string]string) string {
	var result strings.Builder
	
	contextTitle := "Context:"
	if f.useColor {
		contextTitle = color.CyanString("Context:")
	}
	result.WriteString(contextTitle + "\n")
	
	for key, value := range context {
		if f.useColor {
			result.WriteString(fmt.Sprintf("  %s: %s\n", 
				color.BlueString(key), color.WhiteString(value)))
		} else {
			result.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}
	
	return result.String()
}

// formatSuggestions formats suggestions for resolving the error
func (f *ErrorFormatter) formatSuggestions(suggestions []string) string {
	var result strings.Builder
	
	suggestionTitle := "💡 Suggestions:"
	if f.useColor {
		suggestionTitle = color.GreenString("💡 Suggestions:")
	}
	result.WriteString(suggestionTitle + "\n")
	
	for i, suggestion := range suggestions {
		if f.useColor {
			result.WriteString(fmt.Sprintf("  %s. %s\n", 
				color.GreenString(fmt.Sprintf("%d", i+1)), suggestion))
		} else {
			result.WriteString(fmt.Sprintf("  %d. %s\n", i+1, suggestion))
		}
	}
	
	return result.String()
}

// FormatSimple formats any error (GmanError or standard error) in a simple format
func FormatSimple(err error) string {
	if gErr, ok := err.(*GmanError); ok {
		formatter := NewErrorFormatter().WithCompact(true)
		return formatter.Format(gErr)
	}
	
	// Format standard errors
	return fmt.Sprintf("❌ ERROR: %s", err.Error())
}

// FormatDetailed formats any error (GmanError or standard error) with full details
func FormatDetailed(err error) string {
	if gErr, ok := err.(*GmanError); ok {
		formatter := NewErrorFormatter()
		return formatter.Format(gErr)
	}
	
	// Format standard errors with basic structure
	return fmt.Sprintf("❌ ERROR: %s\n\n💡 Suggestions:\n  1. Check the operation and try again\n  2. Review any error messages for specific guidance\n", err.Error())
}

// PrettyPrint prints a formatted error to the console
func PrettyPrint(err error) {
	fmt.Print(FormatDetailed(err))
}

// PrettyPrintSimple prints a simple formatted error to the console
func PrettyPrintSimple(err error) {
	fmt.Println(FormatSimple(err))
}

// Enhanced Error Formatter Methods

// Format formats an error according to the configured display mode
func (f *EnhancedErrorFormatter) Format(err *GmanError) string {
	switch f.config.Mode {
	case DisplayModeCompact:
		return f.formatCompactEnhanced(err)
	case DisplayModeDetailed:
		return f.formatDetailedEnhanced(err)
	case DisplayModeInteractive:
		return f.formatInteractive(err)
	case DisplayModeJSON:
		return f.formatJSON(err)
	case DisplayModeTable:
		return f.formatTable(err)
	default:
		return f.formatDetailedEnhanced(err)
	}
}

// FormatWithRecovery formats an error with recovery plan
func (f *EnhancedErrorFormatter) FormatWithRecovery(err *GmanError, plan *RecoveryPlan) string {
	errorDisplay := f.Format(err)
	
	if !f.config.ShowRecovery || plan == nil || plan.PrimaryAction == nil {
		return errorDisplay
	}

	recoveryDisplay := f.formatRecoveryPlan(plan)
	return errorDisplay + "\n" + recoveryDisplay
}

// formatCompactEnhanced creates a single-line error representation
func (f *EnhancedErrorFormatter) formatCompactEnhanced(err *GmanError) string {
	icon := f.getSeverityIconEnhanced(err.Severity)
	severityColor := f.getSeverityColorEnhanced(err.Severity)
	
	message := err.Message
	if len(message) > 60 {
		message = message[:57] + "..."
	}
	
	if f.config.ColorEnabled {
		return fmt.Sprintf("%s %s: %s", icon, severityColor(string(err.Type)), message)
	}
	return fmt.Sprintf("%s %s: %s", icon, err.Type, message)
}

// formatDetailedEnhanced creates a comprehensive multi-line error display
func (f *EnhancedErrorFormatter) formatDetailedEnhanced(err *GmanError) string {
	var result strings.Builder
	
	// Header with severity and type
	header := f.buildErrorHeader(err)
	result.WriteString(header)
	result.WriteString("\n")
	
	// Main message
	message := f.wrapText(err.Message, f.config.MaxWidth-f.config.IndentSize)
	result.WriteString(f.indent(message))
	result.WriteString("\n")
	
	// Context information
	if f.config.ShowContext && len(err.Context) > 0 {
		result.WriteString("\n")
		result.WriteString(f.formatContextEnhanced(err.Context))
	}
	
	// Technical details
	if f.config.Verbose {
		result.WriteString("\n")
		result.WriteString(f.formatTechnicalDetails(err))
	}
	
	// Suggestions
	if f.config.ShowSuggestions && len(err.Suggestions) > 0 {
		result.WriteString("\n")
		result.WriteString(f.formatSuggestionsEnhanced(err.Suggestions))
	}
	
	return result.String()
}

// formatInteractive creates an interactive error display with user prompts
func (f *EnhancedErrorFormatter) formatInteractive(err *GmanError) string {
	var result strings.Builder
	
	// Use detailed format as base
	result.WriteString(f.formatDetailedEnhanced(err))
	
	// Add interactive elements
	result.WriteString("\n")
	result.WriteString(f.colorize("🔧 Recovery options available. Run with --interactive for guided recovery.", color.FgCyan))
	result.WriteString("\n")
	
	return result.String()
}

// formatJSON creates a JSON representation of the error
func (f *EnhancedErrorFormatter) formatJSON(err *GmanError) string {
	return fmt.Sprintf(`{
  "type": "%s",
  "severity": "%s",
  "message": "%s",
  "timestamp": "%s",
  "recoverable": %t,
  "context": %s,
  "suggestions": %s
}`,
		err.Type,
		err.Severity,
		f.escapeJSON(err.Message),
		time.Now().Format(time.RFC3339),
		err.IsRecoverable(),
		f.contextToJSON(err.Context),
		f.suggestionsToJSON(err.Suggestions))
}

// formatTable creates a tabular error display
func (f *EnhancedErrorFormatter) formatTable(err *GmanError) string {
	var result strings.Builder
	
	result.WriteString("┌" + strings.Repeat("─", f.config.MaxWidth-2) + "┐\n")
	result.WriteString(fmt.Sprintf("│ %s │\n", f.centerText(f.buildErrorHeader(err), f.config.MaxWidth-4)))
	result.WriteString("├" + strings.Repeat("─", f.config.MaxWidth-2) + "┤\n")
	
	// Message rows
	messageLines := f.wrapTextToLines(err.Message, f.config.MaxWidth-6)
	for _, line := range messageLines {
		result.WriteString(fmt.Sprintf("│ %s │\n", f.padText(line, f.config.MaxWidth-4)))
	}
	
	result.WriteString("└" + strings.Repeat("─", f.config.MaxWidth-2) + "┘\n")
	
	return result.String()
}

// buildErrorHeader creates the error header with icon, severity, and type
func (f *EnhancedErrorFormatter) buildErrorHeader(err *GmanError) string {
	icon := f.getSeverityIconEnhanced(err.Severity)
	severityColor := f.getSeverityColorEnhanced(err.Severity)
	
	header := fmt.Sprintf("%s %s", icon, err.Type)
	if f.config.ShowTimestamp {
		header += fmt.Sprintf(" (%s)", time.Now().Format("15:04:05"))
	}
	
	if f.config.ColorEnabled {
		return severityColor(header)
	}
	return header
}

// formatContextEnhanced formats the error context information
func (f *EnhancedErrorFormatter) formatContextEnhanced(context map[string]string) string {
	if len(context) == 0 {
		return ""
	}
	
	var result strings.Builder
	result.WriteString(f.colorize("📍 Context:", color.FgBlue))
	result.WriteString("\n")
	
	for key, value := range context {
		line := fmt.Sprintf("%s: %s", key, value)
		result.WriteString(f.indent(line))
		result.WriteString("\n")
	}
	
	return strings.TrimRight(result.String(), "\n")
}

// formatTechnicalDetails formats technical error details for verbose mode
func (f *EnhancedErrorFormatter) formatTechnicalDetails(err *GmanError) string {
	var result strings.Builder
	result.WriteString(f.colorize("🔧 Technical Details:", color.FgCyan))
	result.WriteString("\n")
	
	details := []string{
		fmt.Sprintf("Error Type: %s", err.Type),
		fmt.Sprintf("Severity: %s", err.Severity),
		fmt.Sprintf("Recoverable: %t", err.IsRecoverable()),
		fmt.Sprintf("Timestamp: %s", time.Now().Format(time.RFC3339)),
	}
	
	for _, detail := range details {
		result.WriteString(f.indent(detail))
		result.WriteString("\n")
	}
	
	return strings.TrimRight(result.String(), "\n")
}

// formatSuggestionsEnhanced formats error suggestions
func (f *EnhancedErrorFormatter) formatSuggestionsEnhanced(suggestions []string) string {
	if len(suggestions) == 0 {
		return ""
	}
	
	var result strings.Builder
	result.WriteString(f.colorize("💡 Suggestions:", color.FgYellow))
	result.WriteString("\n")
	
	for i, suggestion := range suggestions {
		line := fmt.Sprintf("%d. %s", i+1, suggestion)
		result.WriteString(f.indent(line))
		result.WriteString("\n")
	}
	
	return strings.TrimRight(result.String(), "\n")
}

// formatRecoveryPlan formats a recovery plan for display
func (f *EnhancedErrorFormatter) formatRecoveryPlan(plan *RecoveryPlan) string {
	var result strings.Builder
	
	result.WriteString(f.colorize("🔧 Recovery Options:", color.FgCyan))
	result.WriteString("\n")
	
	// Primary action
	if plan.PrimaryAction != nil {
		result.WriteString(f.formatRecoveryAction("Primary", plan.PrimaryAction, 1))
	}
	
	// Alternative actions
	for i, action := range plan.AlternativeActions {
		if i >= 2 {
			break // Limit to first 2 alternatives
		}
		result.WriteString(f.formatRecoveryAction("Alternative", action, i+2))
	}
	
	// Impact information
	if plan.EstimatedImpact != "" {
		result.WriteString("\n")
		result.WriteString(f.indent(f.colorize("Impact: "+plan.EstimatedImpact, color.FgMagenta)))
		result.WriteString("\n")
	}
	
	return result.String()
}

// formatRecoveryAction formats a single recovery action
func (f *EnhancedErrorFormatter) formatRecoveryAction(label string, action *RecoveryAction, number int) string {
	var result strings.Builder
	
	// Action header
	header := fmt.Sprintf("%d. [%s] %s", number, action.Strategy, action.Description)
	result.WriteString(f.indent(header))
	result.WriteString("\n")
	
	// Command if available
	if action.Command != "" && !strings.HasPrefix(action.Command, "#") {
		cmdLine := f.colorize("Command: "+action.Command, color.FgYellow)
		result.WriteString(f.indent(f.indent(cmdLine)))
		result.WriteString("\n")
	}
	
	// Metadata
	metadata := []string{}
	if action.EstimatedTime != "" {
		metadata = append(metadata, "Time: "+action.EstimatedTime)
	}
	if action.SafeLevel > 0 {
		metadata = append(metadata, fmt.Sprintf("Safety: %d/5", action.SafeLevel))
	}
	if action.AutoExec {
		metadata = append(metadata, "Auto-executable")
	}
	
	if len(metadata) > 0 {
		metaLine := strings.Join(metadata, " | ")
		result.WriteString(f.indent(f.indent(f.colorize(metaLine, color.FgBlue))))
		result.WriteString("\n")
	}
	
	return result.String()
}

// Utility methods for enhanced formatting

// getSeverityIconEnhanced returns an appropriate icon for error severity
func (f *EnhancedErrorFormatter) getSeverityIconEnhanced(severity Severity) string {
	if !f.config.ShowIcons {
		return ""
	}
	
	switch severity {
	case SeverityCritical:
		return "🚨"
	case SeverityError:
		return "❌"
	case SeverityWarning:
		return "⚠️"
	case SeverityInfo:
		return "ℹ️"
	default:
		return "❓"
	}
}

// getSeverityColorEnhanced returns appropriate color function for severity
func (f *EnhancedErrorFormatter) getSeverityColorEnhanced(severity Severity) func(...interface{}) string {
	if !f.config.ColorEnabled {
		return func(a ...interface{}) string { return fmt.Sprint(a...) }
	}
	
	switch severity {
	case SeverityCritical:
		return color.New(color.FgRed, color.Bold).SprintFunc()
	case SeverityError:
		return color.New(color.FgRed).SprintFunc()
	case SeverityWarning:
		return color.New(color.FgYellow).SprintFunc()
	case SeverityInfo:
		return color.New(color.FgBlue).SprintFunc()
	default:
		return color.New(color.FgWhite).SprintFunc()
	}
}

// colorize applies color to text if colors are enabled
func (f *EnhancedErrorFormatter) colorize(text string, colorAttr color.Attribute) string {
	if !f.config.ColorEnabled {
		return text
	}
	return color.New(colorAttr).Sprint(text)
}

// indent adds indentation to text
func (f *EnhancedErrorFormatter) indent(text string) string {
	indent := strings.Repeat(" ", f.config.IndentSize)
	return indent + strings.ReplaceAll(text, "\n", "\n"+indent)
}

// wrapText wraps text to specified width
func (f *EnhancedErrorFormatter) wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}
	
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}
	
	var result strings.Builder
	currentLine := words[0]
	
	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			result.WriteString(currentLine + "\n")
			currentLine = word
		}
	}
	
	result.WriteString(currentLine)
	return result.String()
}

// wrapTextToLines wraps text and returns as slice of lines
func (f *EnhancedErrorFormatter) wrapTextToLines(text string, width int) []string {
	return strings.Split(f.wrapText(text, width), "\n")
}

// padText pads text to specified width
func (f *EnhancedErrorFormatter) padText(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}
	return text + strings.Repeat(" ", width-len(text))
}

// centerText centers text within specified width
func (f *EnhancedErrorFormatter) centerText(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}
	
	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text + strings.Repeat(" ", width-len(text)-padding)
}

// escapeJSON escapes string for JSON
func (f *EnhancedErrorFormatter) escapeJSON(text string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
		"\n", "\\n",
		"\r", "\\r",
		"\t", "\\t",
	)
	return replacer.Replace(text)
}

// contextToJSON converts context map to JSON string
func (f *EnhancedErrorFormatter) contextToJSON(context map[string]string) string {
	if len(context) == 0 {
		return "{}"
	}
	
	var pairs []string
	for k, v := range context {
		pairs = append(pairs, fmt.Sprintf("\"%s\": \"%s\"", f.escapeJSON(k), f.escapeJSON(v)))
	}
	
	return "{" + strings.Join(pairs, ", ") + "}"
}

// suggestionsToJSON converts suggestions slice to JSON string
func (f *EnhancedErrorFormatter) suggestionsToJSON(suggestions []string) string {
	if len(suggestions) == 0 {
		return "[]"
	}
	
	var quoted []string
	for _, s := range suggestions {
		quoted = append(quoted, "\""+f.escapeJSON(s)+"\"")
	}
	
	return "[" + strings.Join(quoted, ", ") + "]"
}

// ErrorSummary represents a summary of multiple errors
type ErrorSummary struct {
	TotalErrors    int                    `json:"total_errors"`
	ErrorsBySeverity map[Severity]int      `json:"errors_by_severity"`
	ErrorsByType   map[ErrorType]int      `json:"errors_by_type"`
	TimeRange      string                 `json:"time_range"`
	MostCommonType ErrorType              `json:"most_common_type"`
	HighestSeverity Severity              `json:"highest_severity"`
}

// FormatErrorSummary creates a formatted summary of multiple errors
func (f *EnhancedErrorFormatter) FormatErrorSummary(summary *ErrorSummary) string {
	var result strings.Builder
	
	header := fmt.Sprintf("📊 Error Summary (%d total)", summary.TotalErrors)
	result.WriteString(f.colorize(header, color.FgCyan))
	result.WriteString("\n")
	result.WriteString(strings.Repeat("─", len(header)))
	result.WriteString("\n\n")
	
	// Severity breakdown
	result.WriteString(f.colorize("By Severity:", color.FgBlue))
	result.WriteString("\n")
	for severity, count := range summary.ErrorsBySeverity {
		icon := f.getSeverityIconEnhanced(severity)
		line := fmt.Sprintf("%s %s: %d", icon, severity, count)
		result.WriteString(f.indent(line))
		result.WriteString("\n")
	}
	
	// Type breakdown (top 5)
	result.WriteString("\n")
	result.WriteString(f.colorize("Most Common Types:", color.FgBlue))
	result.WriteString("\n")
	
	// Sort types by frequency
	type typeCount struct {
		errorType ErrorType
		count     int
	}
	
	var typeCounts []typeCount
	for errorType, count := range summary.ErrorsByType {
		typeCounts = append(typeCounts, typeCount{errorType, count})
	}
	
	// Simple sort by count (descending)
	for i := 0; i < len(typeCounts)-1; i++ {
		for j := i + 1; j < len(typeCounts); j++ {
			if typeCounts[j].count > typeCounts[i].count {
				typeCounts[i], typeCounts[j] = typeCounts[j], typeCounts[i]
			}
		}
	}
	
	maxShow := 5
	if len(typeCounts) < maxShow {
		maxShow = len(typeCounts)
	}
	
	for i := 0; i < maxShow; i++ {
		line := fmt.Sprintf("%d. %s: %d occurrences", i+1, typeCounts[i].errorType, typeCounts[i].count)
		result.WriteString(f.indent(line))
		result.WriteString("\n")
	}
	
	// Time range
	if summary.TimeRange != "" {
		result.WriteString("\n")
		result.WriteString(f.colorize("Time Range: "+summary.TimeRange, color.FgMagenta))
		result.WriteString("\n")
	}
	
	return result.String()
}

// Global Enhanced Formatter Functions

// FormatStructured formats an error using enhanced structured display
func FormatStructured(err error, mode DisplayMode) string {
	gErr := ToGmanError(err)
	config := DefaultDisplayConfig()
	config.Mode = mode
	
	formatter := NewEnhancedErrorFormatter(config)
	return formatter.Format(gErr)
}

// FormatWithConfig formats an error with custom configuration
func FormatWithConfig(err error, config *ErrorDisplayConfig) string {
	gErr := ToGmanError(err)
	formatter := NewEnhancedErrorFormatter(config)
	return formatter.Format(gErr)
}

// FormatInteractiveError formats an error for interactive display
func FormatInteractiveError(err error) string {
	return FormatStructured(err, DisplayModeInteractive)
}

// FormatTableError formats an error in table format
func FormatTableError(err error) string {
	return FormatStructured(err, DisplayModeTable)
}

// FormatJSONError formats an error as JSON
func FormatJSONError(err error) string {
	return FormatStructured(err, DisplayModeJSON)
}