package tui

import (
	"github.com/charmbracelet/lipgloss"
	"gman/internal/tui/styles"
)

// KeyBinding represents a keyboard shortcut
type KeyBinding struct {
	Key         string
	Description string
	Global      bool
	Panel       string
}

// GetGlobalKeyBindings returns global keyboard shortcuts
func GetGlobalKeyBindings() []KeyBinding {
	return []KeyBinding{
		{"Tab", "Next panel", true, ""},
		{"Shift+Tab", "Previous panel", true, ""},
		{"1-4", "Jump to panel", true, ""},
		{"?/h", "Toggle help", true, ""},
		{"q/Ctrl+C", "Quit application", true, ""},
		{"r", "Refresh all", true, ""},
		{"F5", "Force refresh", true, ""},
	}
}

// GetRepositoryKeyBindings returns repository panel keyboard shortcuts
func GetRepositoryKeyBindings() []KeyBinding {
	return []KeyBinding{
		{"↑/k", "Move up", false, "Repository"},
		{"↓/j", "Move down", false, "Repository"},
		{"Enter", "Select repository", false, "Repository"},
		{"/", "Filter repositories", false, "Repository"},
		{"s", "Change sort order", false, "Repository"},
		{"g", "Filter by group", false, "Repository"},
		{"Home/End", "First/Last item", false, "Repository"},
		{"Page Up/Down", "Page scroll", false, "Repository"},
	}
}

// GetStatusKeyBindings returns status panel keyboard shortcuts
func GetStatusKeyBindings() []KeyBinding {
	return []KeyBinding{
		{"e", "Toggle extended view", false, "Status"},
		{"b", "Toggle branch info", false, "Status"},
		{"a", "Toggle auto-refresh", false, "Status"},
		{"r", "Refresh status", false, "Status"},
	}
}

// GetSearchKeyBindings returns search panel keyboard shortcuts
func GetSearchKeyBindings() []KeyBinding {
	return []KeyBinding{
		{"Enter", "Start search", false, "Search"},
		{"Tab", "Toggle search mode", false, "Search"},
		{"/", "Search files", false, "Search"},
		{"c", "Search commits", false, "Search"},
		{"↑/k", "Previous result", false, "Search"},
		{"↓/j", "Next result", false, "Search"},
		{"Esc", "Cancel search", false, "Search"},
	}
}

// GetPreviewKeyBindings returns preview panel keyboard shortcuts
func GetPreviewKeyBindings() []KeyBinding {
	return []KeyBinding{
		{"↑/k", "Scroll up", false, "Preview"},
		{"↓/j", "Scroll down", false, "Preview"},
		{"Page Up/Down", "Page scroll", false, "Preview"},
		{"Home/End", "Top/Bottom", false, "Preview"},
	}
}

// GetAllKeyBindings returns all keyboard shortcuts organized by panel
func GetAllKeyBindings() map[string][]KeyBinding {
	return map[string][]KeyBinding{
		"Global":     GetGlobalKeyBindings(),
		"Repository": GetRepositoryKeyBindings(),
		"Status":     GetStatusKeyBindings(),
		"Search":     GetSearchKeyBindings(),
		"Preview":    GetPreviewKeyBindings(),
	}
}

// RenderKeyBindings renders keyboard shortcuts in a formatted way
func RenderKeyBindings(bindings []KeyBinding, title string) string {
	var content []string
	
	// Title
	titleStyle := styles.SubHeaderStyle.Bold(true)
	content = append(content, titleStyle.Render(title))
	content = append(content, "")
	
	// Key bindings
	for _, binding := range bindings {
		keyStyle := styles.HelpKeyStyle
		descStyle := styles.HelpDescStyle
		
		line := lipgloss.JoinHorizontal(
			lipgloss.Left,
			keyStyle.Width(15).Render(binding.Key),
			descStyle.Render(binding.Description),
		)
		content = append(content, line)
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

// RenderQuickHelp renders a quick help reference
func RenderQuickHelp() string {
	quickHelp := []string{
		"Tab: Next Panel",
		"?: Help",
		"q: Quit",
		"r: Refresh",
	}
	
	style := styles.MutedStyle.Italic(true)
	return style.Render(lipgloss.JoinHorizontal(lipgloss.Left, quickHelp...))
}

// GetContextualHelp returns help text based on the current focused panel
func GetContextualHelp(panelType string) string {
	bindings := GetAllKeyBindings()
	
	switch panelType {
	case "repositories", "Repository":
		return RenderKeyBindings(bindings["Repository"], "Repository Panel")
	case "status", "Status":
		return RenderKeyBindings(bindings["Status"], "Status Panel")
	case "search", "Search":
		return RenderKeyBindings(bindings["Search"], "Search Panel")
	case "preview", "Preview":
		return RenderKeyBindings(bindings["Preview"], "Preview Panel")
	default:
		return RenderKeyBindings(bindings["Global"], "Global Shortcuts")
	}
}