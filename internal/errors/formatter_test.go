package errors

import (
	"strings"
	"testing"
)

func TestEnhancedErrorFormatter_DisplayModes(t *testing.T) {
	testErr := NewRepoNotFoundError("/test/path")
	testErr.WithContext("operation", "switch")
	testErr.WithContext("alias", "test-repo")
	testErr.WithSuggestion("Check repository configuration")
	testErr.WithSuggestion("Use 'gman repo list' to verify repositories")

	testCases := []struct {
		name string
		mode DisplayMode
	}{
		{"Compact", DisplayModeCompact},
		{"Detailed", DisplayModeDetailed},
		{"Interactive", DisplayModeInteractive},
		{"JSON", DisplayModeJSON},
		{"Table", DisplayModeTable},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultDisplayConfig()
			config.Mode = tc.mode
			formatter := NewEnhancedErrorFormatter(config)

			result := formatter.Format(testErr)
			if result == "" {
				t.Error("Expected non-empty formatted output")
			}

			// Mode-specific checks
			switch tc.mode {
			case DisplayModeCompact:
				if strings.Count(result, "\n") > 1 {
					t.Error("Compact mode should produce minimal lines")
				}
			case DisplayModeJSON:
				if !strings.Contains(result, `"type"`) || !strings.Contains(result, `"severity"`) {
					t.Error("JSON mode should contain required fields")
				}
			case DisplayModeTable:
				if !strings.Contains(result, "┌") || !strings.Contains(result, "│") {
					t.Error("Table mode should contain table borders")
				}
			case DisplayModeDetailed:
				if !strings.Contains(result, "Context:") {
					t.Error("Detailed mode should contain context section")
				}
			case DisplayModeInteractive:
				if !strings.Contains(result, "Recovery options") && !strings.Contains(result, "recovery") {
					t.Error("Interactive mode should mention recovery options")
				}
			}
		})
	}
}

func TestEnhancedErrorFormatter_ConfigOptions(t *testing.T) {
	testErr := NewNetworkTimeoutError("test operation", "30s")
	testErr.WithContext("repository", "/test/repo")

	t.Run("ColorEnabled", func(t *testing.T) {
		config := DefaultDisplayConfig()
		config.ColorEnabled = true
		formatter := NewEnhancedErrorFormatter(config)

		result := formatter.Format(testErr)
		// Color codes contain escape sequences
		if !strings.Contains(result, "\033[") && !strings.Contains(result, "\x1b[") {
			t.Log("Warning: Expected color codes in output when colors enabled")
		}
	})

	t.Run("ColorDisabled", func(t *testing.T) {
		config := DefaultDisplayConfig()
		config.ColorEnabled = false
		formatter := NewEnhancedErrorFormatter(config)

		result := formatter.Format(testErr)
		if strings.Contains(result, "\033[") || strings.Contains(result, "\x1b[") {
			t.Error("Should not contain color codes when colors disabled")
		}
	})

	t.Run("VerboseMode", func(t *testing.T) {
		config := DefaultDisplayConfig()
		config.Verbose = true
		formatter := NewEnhancedErrorFormatter(config)

		result := formatter.Format(testErr)
		if !strings.Contains(result, "Technical Details:") {
			t.Error("Verbose mode should include technical details")
		}
		if !strings.Contains(result, "Error Type:") {
			t.Error("Verbose mode should include error type")
		}
	})

	t.Run("TimestampDisplay", func(t *testing.T) {
		config := DefaultDisplayConfig()
		config.ShowTimestamp = true
		formatter := NewEnhancedErrorFormatter(config)

		result := formatter.Format(testErr)
		// Should contain time format
		if !strings.Contains(result, ":") {
			t.Error("Should contain timestamp when ShowTimestamp is true")
		}
	})

	t.Run("ContextDisplay", func(t *testing.T) {
		config := DefaultDisplayConfig()
		config.ShowContext = false
		formatter := NewEnhancedErrorFormatter(config)

		result := formatter.Format(testErr)
		if strings.Contains(result, "Context:") {
			t.Error("Should not show context when ShowContext is false")
		}
	})
}

func TestEnhancedErrorFormatter_WithRecovery(t *testing.T) {
	testErr := NewMergeConflictError("/test/repo")
	engine := NewRecoveryEngine()
	plan := engine.CreateRecoveryPlan(testErr)

	config := DefaultDisplayConfig()
	formatter := NewEnhancedErrorFormatter(config)

	result := formatter.FormatWithRecovery(testErr, plan)
	
	if !strings.Contains(result, "Recovery Options:") {
		t.Error("Should contain recovery options section")
	}
	
	if !strings.Contains(result, "[MANUAL]") || !strings.Contains(result, "[USER_INPUT]") {
		t.Error("Should contain recovery strategy indicators")
	}
	
	if !strings.Contains(result, "Safety:") {
		t.Error("Should contain safety level information")
	}
}

func TestEnhancedErrorFormatter_TextWrapping(t *testing.T) {
	longMessage := "This is a very long error message that should be wrapped when the maximum width is exceeded to ensure proper display formatting and readability"
	testErr := NewConfigInvalidError(longMessage)

	config := DefaultDisplayConfig()
	config.MaxWidth = 40
	formatter := NewEnhancedErrorFormatter(config)

	result := formatter.Format(testErr)
	lines := strings.Split(result, "\n")
	
	for _, line := range lines {
		// Remove ANSI color codes for accurate length measurement
		cleanLine := removeANSICodes(line)
		if len(cleanLine) > config.MaxWidth+config.IndentSize {
			t.Errorf("Line exceeds maximum width: %d > %d\nLine: %s", 
				len(cleanLine), config.MaxWidth+config.IndentSize, cleanLine)
		}
	}
}

func TestEnhancedErrorFormatter_JSONFormat(t *testing.T) {
	testErr := NewToolNotAvailableError("test-tool", "Test tool not found")
	testErr.WithContext("operation", "file search")
	testErr.WithSuggestion("Install the missing tool")

	config := DefaultDisplayConfig()
	config.Mode = DisplayModeJSON
	formatter := NewEnhancedErrorFormatter(config)

	result := formatter.Format(testErr)

	// Basic JSON structure checks
	expectedFields := []string{
		`"type"`,
		`"severity"`,
		`"message"`,
		`"timestamp"`,
		`"recoverable"`,
		`"context"`,
		`"suggestions"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(result, field) {
			t.Errorf("JSON output missing required field: %s", field)
		}
	}

	// Should be valid JSON structure
	if !strings.HasPrefix(result, "{") || !strings.HasSuffix(result, "}") {
		t.Error("JSON output should be wrapped in braces")
	}
}

func TestEnhancedErrorFormatter_TableFormat(t *testing.T) {
	testErr := NewRepoNotFoundError("/test/path")
	testErr.WithContext("command", "switch")

	config := DefaultDisplayConfig()
	config.Mode = DisplayModeTable
	config.MaxWidth = 60
	formatter := NewEnhancedErrorFormatter(config)

	result := formatter.Format(testErr)

	// Table structure checks
	tableChars := []string{"┌", "┐", "├", "┤", "└", "┘", "│", "─"}
	for _, char := range tableChars {
		if !strings.Contains(result, char) {
			t.Errorf("Table format missing character: %s", char)
		}
	}

	// Check that table width is respected
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if len(line) > config.MaxWidth {
			t.Errorf("Table line exceeds maximum width: %d > %d", len(line), config.MaxWidth)
		}
	}
}

func TestErrorSummary_Formatting(t *testing.T) {
	summary := &ErrorSummary{
		TotalErrors: 15,
		ErrorsBySeverity: map[Severity]int{
			SeverityError:   8,
			SeverityWarning: 5,
			SeverityInfo:    2,
		},
		ErrorsByType: map[ErrorType]int{
			ErrTypeRepoNotFound:    6,
			ErrTypeNetworkTimeout:  4,
			ErrTypeMergeConflict:   3,
			ErrTypeToolNotAvailable: 2,
		},
		TimeRange:       "Last 24 hours",
		MostCommonType:  ErrTypeRepoNotFound,
		HighestSeverity: SeverityError,
	}

	config := DefaultDisplayConfig()
	formatter := NewEnhancedErrorFormatter(config)

	result := formatter.FormatErrorSummary(summary)

	// Check for required summary sections
	if !strings.Contains(result, "Error Summary") {
		t.Error("Should contain error summary header")
	}
	
	if !strings.Contains(result, "By Severity:") {
		t.Error("Should contain severity breakdown")
	}
	
	if !strings.Contains(result, "Most Common Types:") {
		t.Error("Should contain type breakdown")
	}
	
	if !strings.Contains(result, "15 total") {
		t.Error("Should contain total error count")
	}
	
	if !strings.Contains(result, "Last 24 hours") {
		t.Error("Should contain time range information")
	}
}

func TestDisplayConfigPresets(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultDisplayConfig()
		
		if config.Mode != DisplayModeDetailed {
			t.Error("Default mode should be detailed")
		}
		if !config.ShowTimestamp {
			t.Error("Default should show timestamps")
		}
		if !config.ColorEnabled {
			t.Error("Default should enable colors")
		}
	})

	t.Run("CompactConfig", func(t *testing.T) {
		config := CompactDisplayConfig()
		
		if config.Mode != DisplayModeCompact {
			t.Error("Compact config mode should be compact")
		}
		if config.ShowTimestamp {
			t.Error("Compact config should not show timestamps")
		}
		if config.ShowIcons {
			t.Error("Compact config should not show icons")
		}
	})
}

func TestGlobalFormatterFunctions(t *testing.T) {
	testErr := NewConfigInvalidError("Test error")

	t.Run("FormatStructured", func(t *testing.T) {
		result := FormatStructured(testErr, DisplayModeTable)
		if !strings.Contains(result, "┌") {
			t.Error("FormatStructured with table mode should produce table format")
		}
	})

	t.Run("FormatInteractiveError", func(t *testing.T) {
		result := FormatInteractiveError(testErr)
		if !strings.Contains(result, "Recovery options") {
			t.Error("FormatInteractiveError should mention recovery options")
		}
	})

	t.Run("FormatJSONError", func(t *testing.T) {
		result := FormatJSONError(testErr)
		if !strings.Contains(result, `"type"`) {
			t.Error("FormatJSONError should produce JSON format")
		}
	})

	t.Run("FormatTableError", func(t *testing.T) {
		result := FormatTableError(testErr)
		if !strings.Contains(result, "│") {
			t.Error("FormatTableError should produce table format")
		}
	})
}

// Helper function to remove ANSI color codes for testing
func removeANSICodes(text string) string {
	// Simple ANSI escape sequence removal for testing
	result := text
	// Remove color codes like \033[31m or \x1b[31m
	for strings.Contains(result, "\033[") {
		start := strings.Index(result, "\033[")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "m")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	for strings.Contains(result, "\x1b[") {
		start := strings.Index(result, "\x1b[")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "m")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	return result
}

func TestEnhancedErrorFormatter_EdgeCases(t *testing.T) {
	t.Run("EmptyError", func(t *testing.T) {
		testErr := &GmanError{
			Type:      "UNKNOWN",
			Severity:  SeverityInfo,
			Message:   "",
			Context:   make(map[string]string),
			Suggestions: []string{},
		}

		config := DefaultDisplayConfig()
		formatter := NewEnhancedErrorFormatter(config)

		result := formatter.Format(testErr)
		if result == "" {
			t.Error("Should handle empty error gracefully")
		}
	})

	t.Run("NilRecoveryPlan", func(t *testing.T) {
		testErr := NewRepoNotFoundError("/test")
		config := DefaultDisplayConfig()
		formatter := NewEnhancedErrorFormatter(config)

		result := formatter.FormatWithRecovery(testErr, nil)
		// Should just return the error without recovery
		if strings.Contains(result, "Recovery Options:") {
			t.Error("Should not show recovery options for nil plan")
		}
	})

	t.Run("VeryWideText", func(t *testing.T) {
		wideMessage := strings.Repeat("A", 200)
		testErr := NewConfigInvalidError(wideMessage)

		config := DefaultDisplayConfig()
		config.MaxWidth = 50
		formatter := NewEnhancedErrorFormatter(config)

		result := formatter.Format(testErr)
		lines := strings.Split(result, "\n")
		
		for _, line := range lines {
			cleanLine := removeANSICodes(line)
			if len(cleanLine) > config.MaxWidth+10 { // Allow some margin for indentation
				t.Errorf("Line too wide: %d characters", len(cleanLine))
			}
		}
	})
}