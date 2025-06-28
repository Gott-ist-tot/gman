package errors

import (
	"fmt"
	"strings"

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

// ErrorDisplayConfig contains configuration for error display formatting (simplified)
type ErrorDisplayConfig struct {
	Mode           DisplayMode
	ShowSuggestions bool
	ColorEnabled   bool
	MaxWidth       int
	Verbose        bool
}

// DefaultDisplayConfig returns a sensible default configuration
func DefaultDisplayConfig() *ErrorDisplayConfig {
	return &ErrorDisplayConfig{
		Mode:           DisplayModeDetailed,
		ShowSuggestions: true,
		ColorEnabled:   true,
		MaxWidth:       80,
		Verbose:        false,
	}
}

// CompactDisplayConfig returns configuration for compact display
func CompactDisplayConfig() *ErrorDisplayConfig {
	return &ErrorDisplayConfig{
		Mode:           DisplayModeCompact,
		ShowSuggestions: false,
		ColorEnabled:   true,
		MaxWidth:       120,
		Verbose:        false,
	}
}

// ErrorFormatter provides simple error formatting
type ErrorFormatter struct {
	config *ErrorDisplayConfig
}

// NewErrorFormatter creates a new error formatter with default config
func NewErrorFormatter() *ErrorFormatter {
	return &ErrorFormatter{
		config: DefaultDisplayConfig(),
	}
}

// WithCompact sets compact mode
func (f *ErrorFormatter) WithCompact(compact bool) *ErrorFormatter {
	if compact {
		f.config = CompactDisplayConfig()
	} else {
		f.config = DefaultDisplayConfig()
	}
	return f
}

// Format formats a GmanError for display
func (f *ErrorFormatter) Format(err *GmanError) string {
	if err == nil {
		return ""
	}

	switch f.config.Mode {
	case DisplayModeCompact:
		return f.formatCompact(err)
	case DisplayModeJSON:
		return f.formatJSON(err)
	default:
		return f.formatDetailed(err)
	}
}

// formatCompact formats error in a single line
func (f *ErrorFormatter) formatCompact(err *GmanError) string {
	var parts []string
	
	// Error type and message
	if f.config.ColorEnabled {
		parts = append(parts, color.RedString("[%s]", err.Type), err.Message)
	} else {
		parts = append(parts, fmt.Sprintf("[%s]", err.Type), err.Message)
	}
	
	// Add cause if available
	if err.Cause != nil {
		parts = append(parts, fmt.Sprintf("(caused by: %s)", err.Cause.Error()))
	}
	
	return strings.Join(parts, " ")
}

// formatDetailed formats error with full details
func (f *ErrorFormatter) formatDetailed(err *GmanError) string {
	var result strings.Builder
	
	// Error header
	if f.config.ColorEnabled {
		result.WriteString(color.RedString("Error [%s]: %s\n", err.Type, err.Message))
	} else {
		result.WriteString(fmt.Sprintf("Error [%s]: %s\n", err.Type, err.Message))
	}
	
	// Show cause if available
	if err.Cause != nil {
		if f.config.ColorEnabled {
			result.WriteString(color.YellowString("Caused by: %s\n", err.Cause.Error()))
		} else {
			result.WriteString(fmt.Sprintf("Caused by: %s\n", err.Cause.Error()))
		}
	}
	
	// Show suggestions if enabled and available
	if f.config.ShowSuggestions && len(err.Suggestions) > 0 {
		result.WriteString("\nSuggestions:\n")
		for i, suggestion := range err.Suggestions {
			if f.config.ColorEnabled {
				result.WriteString(color.CyanString("  %d. %s\n", i+1, suggestion))
			} else {
				result.WriteString(fmt.Sprintf("  %d. %s\n", i+1, suggestion))
			}
		}
	}
	
	return result.String()
}

// formatJSON formats error as JSON
func (f *ErrorFormatter) formatJSON(err *GmanError) string {
	// Simple JSON formatting
	var cause string
	if err.Cause != nil {
		cause = err.Cause.Error()
	}
	
	return fmt.Sprintf(`{"type":"%s","message":"%s","cause":"%s","suggestions":[%s]}`,
		err.Type,
		strings.ReplaceAll(err.Message, `"`, `\"`),
		strings.ReplaceAll(cause, `"`, `\"`),
		f.formatSuggestionsJSON(err.Suggestions))
}

// formatSuggestionsJSON formats suggestions as JSON array
func (f *ErrorFormatter) formatSuggestionsJSON(suggestions []string) string {
	if len(suggestions) == 0 {
		return ""
	}
	
	var parts []string
	for _, suggestion := range suggestions {
		parts = append(parts, fmt.Sprintf(`"%s"`, strings.ReplaceAll(suggestion, `"`, `\"`)))
	}
	return strings.Join(parts, ",")
}

// EnhancedErrorFormatter provides enhanced formatting (simplified for compatibility)
type EnhancedErrorFormatter struct {
	*ErrorFormatter
}

// NewEnhancedErrorFormatter creates a new enhanced formatter (simplified)
func NewEnhancedErrorFormatter(config *ErrorDisplayConfig) *EnhancedErrorFormatter {
	formatter := &ErrorFormatter{config: config}
	return &EnhancedErrorFormatter{ErrorFormatter: formatter}
}

// FormatWithRecovery formats error with recovery options (simplified - no recovery)
func (f *EnhancedErrorFormatter) FormatWithRecovery(err *GmanError, plan interface{}) string {
	// Simplified - just format the error without recovery
	return f.Format(err)
}