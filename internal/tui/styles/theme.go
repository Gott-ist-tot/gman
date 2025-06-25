package styles

import (
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
				BorderStyle(lipgloss.ThickBorder())

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
)

// Theme represents a color theme for the application
type Theme struct {
	Name        string
	Primary     lipgloss.Color
	Secondary   lipgloss.Color
	Background  lipgloss.Color
	Surface     lipgloss.Color
	OnPrimary   lipgloss.Color
	OnSecondary lipgloss.Color
	OnBackground lipgloss.Color
	OnSurface   lipgloss.Color
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

// GetStatusIcon returns an icon for a status
func GetStatusIcon(status string) string {
	switch status {
	case "clean":
		return "✅"
	case "dirty":
		return "📝"
	case "ahead":
		return "↑"
	case "behind":
		return "↓"
	case "diverged":
		return "↕"
	case "error":
		return "❌"
	default:
		return "❓"
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
	style := PanelStyle
	titleStyle := PanelTitleStyle
	
	if focused {
		style = PanelFocusedStyle
		titleStyle = PanelTitleFocusedStyle
	}
	
	// Create title bar
	titleBar := titleStyle.Width(width - 2).Render(title)
	
	// Create content area
	contentHeight := height - 3 // Account for title and borders
	contentArea := style.
		Width(width).
		Height(contentHeight).
		Render(content)
	
	return lipgloss.JoinVertical(lipgloss.Left, titleBar, contentArea)
}