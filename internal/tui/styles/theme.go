package styles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	// Base colors
	ColorPrimary   = lipgloss.Color("#7C3AED") // Purple
	ColorSecondary = lipgloss.Color("#10B981") // Green
	ColorAccent    = lipgloss.Color("#F59E0B") // Amber
	ColorDanger    = lipgloss.Color("#EF4444") // Red
	ColorWarning   = lipgloss.Color("#F59E0B") // Amber
	ColorInfo      = lipgloss.Color("#3B82F6") // Blue
	ColorSuccess   = lipgloss.Color("#10B981") // Green

	// Text colors
	ColorTextPrimary   = lipgloss.Color("#F9FAFB") // Light gray
	ColorTextSecondary = lipgloss.Color("#9CA3AF") // Medium gray
	ColorTextMuted     = lipgloss.Color("#6B7280") // Dark gray
	ColorTextInverse   = lipgloss.Color("#1F2937") // Very dark gray

	// Background colors
	ColorBgPrimary   = lipgloss.Color("#1F2937") // Dark gray
	ColorBgSecondary = lipgloss.Color("#374151") // Medium dark gray
	ColorBgMuted     = lipgloss.Color("#4B5563") // Light dark gray
	ColorBgActive    = lipgloss.Color("#7C3AED") // Purple
	ColorBgSelected  = lipgloss.Color("#4C1D95") // Dark purple

	// Border colors
	ColorBorderPrimary   = lipgloss.Color("#6B7280") // Medium gray
	ColorBorderSecondary = lipgloss.Color("#4B5563") // Dark gray
	ColorBorderActive    = lipgloss.Color("#7C3AED") // Purple
	ColorBorderFocused   = lipgloss.Color("#A855F7") // Light purple

	// Status colors
	ColorStatusClean  = lipgloss.Color("#10B981") // Green
	ColorStatusDirty  = lipgloss.Color("#F59E0B") // Amber
	ColorStatusAhead  = lipgloss.Color("#3B82F6") // Blue
	ColorStatusBehind = lipgloss.Color("#8B5CF6") // Light purple
	ColorStatusError  = lipgloss.Color("#EF4444") // Red
)

// Base styles
var (
	BaseStyle = lipgloss.NewStyle().
			Foreground(ColorTextPrimary).
			Background(ColorBgPrimary)

	// Panel styles
	PanelStyle = BaseStyle.Copy().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorderPrimary).
			Padding(0, 1)

	PanelFocusedStyle = PanelStyle.Copy().
				BorderForeground(ColorBorderFocused).
				BorderStyle(lipgloss.DoubleBorder())

	PanelTitleStyle = lipgloss.NewStyle().
			Foreground(ColorTextPrimary).
			Background(ColorBgSecondary).
			Bold(true).
			Padding(0, 1).
			Margin(0, 0, 1, 0)

	PanelTitleFocusedStyle = PanelTitleStyle.Copy().
				Background(ColorBgActive).
				Foreground(ColorTextInverse)

	// Text styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorTextPrimary).
			Bold(true).
			Margin(0, 0, 1, 0)

	SubHeaderStyle = lipgloss.NewStyle().
			Foreground(ColorTextSecondary).
			Bold(false).
			Margin(0, 0, 1, 0)

	BodyStyle = lipgloss.NewStyle().
			Foreground(ColorTextPrimary)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted)

	// List styles
	ListItemStyle = lipgloss.NewStyle().
			Foreground(ColorTextPrimary).
			Padding(0, 1)

	ListItemSelectedStyle = ListItemStyle.Copy().
				Background(ColorBgSelected).
				Foreground(ColorTextPrimary).
				Bold(true)

	ListItemActiveStyle = ListItemStyle.Copy().
				Background(ColorBgActive).
				Foreground(ColorTextInverse).
				Bold(true)

	// Status indicator styles
	StatusCleanStyle = lipgloss.NewStyle().
				Foreground(ColorStatusClean).
				Bold(true)

	StatusDirtyStyle = lipgloss.NewStyle().
				Foreground(ColorStatusDirty).
				Bold(true)

	StatusAheadStyle = lipgloss.NewStyle().
				Foreground(ColorStatusAhead).
				Bold(true)

	StatusBehindStyle = lipgloss.NewStyle().
				Foreground(ColorStatusBehind).
				Bold(true)

	StatusErrorStyle = lipgloss.NewStyle().
				Foreground(ColorStatusError).
				Bold(true)

	// Button styles
	ButtonStyle = lipgloss.NewStyle().
			Foreground(ColorTextInverse).
			Background(ColorBgSecondary).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorderPrimary)

	ButtonActiveStyle = ButtonStyle.Copy().
				Background(ColorBgActive).
				BorderForeground(ColorBorderActive)

	ButtonFocusedStyle = ButtonStyle.Copy().
				Background(ColorPrimary).
				BorderForeground(ColorBorderFocused)

	// Input styles
	InputStyle = lipgloss.NewStyle().
			Foreground(ColorTextPrimary).
			Background(ColorBgSecondary).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorderPrimary)

	InputFocusedStyle = InputStyle.Copy().
				BorderForeground(ColorBorderFocused)

	// Help styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorTextSecondary).
			Background(ColorBgPrimary).
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorderPrimary)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorTextSecondary)

	// Search styles
	SearchInputStyle = InputStyle.Copy().
				Width(30)

	SearchResultStyle = ListItemStyle.Copy()

	SearchResultSelectedStyle = ListItemSelectedStyle.Copy()

	// Preview styles
	PreviewStyle = BaseStyle.Copy().
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorderPrimary)

	PreviewCodeStyle = lipgloss.NewStyle().
				Foreground(ColorTextPrimary).
				Background(ColorBgSecondary).
				Padding(1)

	// Toast notification styles
	ToastSuccessStyle = lipgloss.NewStyle().
				Foreground(ColorTextInverse).
				Background(ColorSuccess).
				Padding(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorSuccess)

	ToastErrorStyle = lipgloss.NewStyle().
				Foreground(ColorTextInverse).
				Background(ColorDanger).
				Padding(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorDanger)

	ToastWarningStyle = lipgloss.NewStyle().
				Foreground(ColorTextInverse).
				Background(ColorWarning).
				Padding(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorWarning)

	ToastInfoStyle = lipgloss.NewStyle().
				Foreground(ColorTextInverse).
				Background(ColorInfo).
				Padding(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorInfo)

	// Progress indicator styles
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Background(ColorBgSecondary).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorderPrimary).
				Padding(0, 1)

	SpinnerStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)
)

// Theme represents a color theme for the application
type Theme struct {
	Name         string
	Primary      lipgloss.Color
	Secondary    lipgloss.Color
	Background   lipgloss.Color
	Surface      lipgloss.Color
	OnPrimary    lipgloss.Color
	OnSecondary  lipgloss.Color
	OnBackground lipgloss.Color
	OnSurface    lipgloss.Color
}

// Default themes
var (
	DarkTheme = Theme{
		Name:         "dark",
		Primary:      ColorPrimary,
		Secondary:    ColorSecondary,
		Background:   ColorBgPrimary,
		Surface:      ColorBgSecondary,
		OnPrimary:    ColorTextInverse,
		OnSecondary:  ColorTextInverse,
		OnBackground: ColorTextPrimary,
		OnSurface:    ColorTextPrimary,
	}

	LightTheme = Theme{
		Name:         "light",
		Primary:      lipgloss.Color("#7C3AED"),
		Secondary:    lipgloss.Color("#10B981"),
		Background:   lipgloss.Color("#F9FAFB"),
		Surface:      lipgloss.Color("#FFFFFF"),
		OnPrimary:    lipgloss.Color("#FFFFFF"),
		OnSecondary:  lipgloss.Color("#FFFFFF"),
		OnBackground: lipgloss.Color("#1F2937"),
		OnSurface:    lipgloss.Color("#1F2937"),
	}
)

// ApplyTheme applies a theme to all styles
func ApplyTheme(theme Theme) {
	// Update color variables
	ColorPrimary = theme.Primary
	ColorSecondary = theme.Secondary
	ColorBgPrimary = theme.Background
	ColorBgSecondary = theme.Surface
	ColorTextPrimary = theme.OnBackground
	ColorTextInverse = theme.OnPrimary

	// Recreate styles with new colors
	BaseStyle = BaseStyle.Foreground(ColorTextPrimary).Background(ColorBgPrimary)
	PanelStyle = PanelStyle.Foreground(ColorTextPrimary).Background(ColorBgPrimary)
	// ... apply to other styles as needed
}

// GetStatusStyle returns the appropriate style for a status
func GetStatusStyle(status string) lipgloss.Style {
	switch status {
	case "clean":
		return StatusCleanStyle
	case "dirty":
		return StatusDirtyStyle
	case "ahead":
		return StatusAheadStyle
	case "behind":
		return StatusBehindStyle
	case "error":
		return StatusErrorStyle
	default:
		return BodyStyle
	}
}

// GetStatusIcon returns an icon for a status with enhanced semantic meaning
func GetStatusIcon(status string) string {
	switch status {
	case "clean":
		return "✅"
	case "dirty", "uncommitted":
		return "📝"
	case "stashed":
		return "📦"
	case "conflicts":
		return "⚠️"
	case "ahead":
		return "↗️"
	case "behind":
		return "↙️"
	case "diverged":
		return "↕️"
	case "error":
		return "❌"
	case "synced":
		return "🔄"
	default:
		return "❓"
	}
}

// GetWorkspaceStatusIcon returns workspace-specific status icons
func GetWorkspaceStatusIcon(workspace string) string {
	switch workspace {
	case "clean":
		return "🟢"
	case "dirty":
		return "🔵"
	case "stashed":
		return "🟡"
	case "conflicts":
		return "🔴"
	default:
		return "⚪"
	}
}

// GetSyncStatusIcon returns sync-specific status icons with directional meaning
func GetSyncStatusIcon(ahead, behind int) string {
	if ahead > 0 && behind > 0 {
		return "🔀" // diverged
	} else if ahead > 0 {
		return "⬆️" // ahead
	} else if behind > 0 {
		return "⬇️" // behind
	} else {
		return "✅" // synced
	}
}

// GetActionIcon returns an icon for action items
func GetActionIcon(actionName string) string {
	switch actionName {
	case "Refresh Status":
		return "🔄"
	case "Open in Terminal":
		return "💻"
	case "Open in File Manager":
		return "📁"
	case "Sync Repository":
		return "🔽"
	case "Commit Changes":
		return "💾"
	case "Push Changes":
		return "⬆️"
	case "Stash Changes":
		return "📦"
	case "Pop Stash":
		return "📤"
	case "Switch Branch":
		return "🌿"
	case "Create Branch":
		return "🌱"
	case "Merge Branch":
		return "🧬"
	case "Create Worktree":
		return "🌳"
	case "Compare Files":
		return "🔍"
	case "View Log":
		return "📜"
	default:
		return "⚡"
	}
}

// Dimensions and layout constants
const (
	MinPanelWidth  = 20
	MinPanelHeight = 8
	PanelPadding   = 1
	PanelMargin    = 1
)

// Layout helpers
func CalculatePanelDimensions(totalWidth, totalHeight int) (int, int, int, int) {
	// Calculate dimensions for 4-panel layout
	// Top-left: Repository list, Top-right: Status detail
	// Bottom-left: Search, Bottom-right: Preview

	halfWidth := (totalWidth - PanelMargin*3) / 2
	halfHeight := (totalHeight - PanelMargin*3) / 2

	return halfWidth, halfHeight, halfWidth, halfHeight
}

// Helper function to create a bordered panel
func CreatePanel(content string, title string, width, height int, focused bool) string {
	// Ensure minimum dimensions
	if width < 10 {
		width = 10
	}
	if height < 5 {
		height = 5
	}

	style := PanelStyle
	titleStyle := PanelTitleStyle

	// Add focus indicator to title
	focusIndicator := ""
	if focused {
		style = PanelFocusedStyle
		titleStyle = PanelTitleFocusedStyle
		focusIndicator = "◆ "
	} else {
		focusIndicator = "  "
	}

	// Create title bar with focus indicator
	titleText := focusIndicator + title
	titleWidth := width - 4 // Account for borders and padding
	if titleWidth < 1 {
		titleWidth = 1
	}
	titleBar := titleStyle.Width(titleWidth).Render(titleText)

	// Create content area
	contentHeight := height - 4 // Account for title, borders, and spacing
	if contentHeight < 1 {
		contentHeight = 1
	}
	
	contentWidth := width - 4 // Account for borders and padding
	if contentWidth < 1 {
		contentWidth = 1
	}

	// Add padding to content if empty or very short
	if len(content) < 20 {
		content = content + "\n\n" // Add some breathing room
	}

	contentArea := style.
		Width(contentWidth).
		Height(contentHeight).
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, titleBar, contentArea)
}

// GetToastStyle returns the appropriate style for a toast type
func GetToastStyle(toastType interface{}) lipgloss.Style {
	// Import the ToastType from models package if needed
	switch toastType.(type) {
	case int:
		switch toastType.(int) {
		case 0: // ToastSuccess
			return ToastSuccessStyle
		case 1: // ToastError
			return ToastErrorStyle
		case 2: // ToastWarning
			return ToastWarningStyle
		case 3: // ToastInfo
			return ToastInfoStyle
		}
	}
	return ToastInfoStyle
}

// RenderToast renders a toast notification
func RenderToast(message string, toastType interface{}) string {
	style := GetToastStyle(toastType)
	icon := getToastIcon(toastType)
	return style.Render(icon + " " + message)
}

// getToastIcon returns an icon for a toast type
func getToastIcon(toastType interface{}) string {
	switch toastType.(type) {
	case int:
		switch toastType.(int) {
		case 0: // ToastSuccess
			return "✅"
		case 1: // ToastError
			return "❌"
		case 2: // ToastWarning
			return "⚠️"
		case 3: // ToastInfo
			return "ℹ️"
		}
	}
	return "ℹ️"
}

// RenderProgressBar renders a progress bar
func RenderProgressBar(progress int, message string, width int) string {
	if width <= 0 {
		width = 20
	}
	
	filled := int(float64(progress) / 100.0 * float64(width))
	if filled > width {
		filled = width
	}
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	progressText := fmt.Sprintf(" %d%% ", progress)
	
	content := fmt.Sprintf("%s [%s]%s", message, bar, progressText)
	return ProgressBarStyle.Render(content)
}

// RenderSpinner renders a spinner with message
func RenderSpinner(message string, frame int) string {
	spinChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinner := spinChars[frame%len(spinChars)]
	content := fmt.Sprintf("%s %s", spinner, message)
	return SpinnerStyle.Render(content)
}
