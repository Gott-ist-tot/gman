package panels

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"gman/internal/tui/models"
	"gman/internal/tui/styles"
)

// PreviewPanel displays preview content for selected items
type PreviewPanel struct {
	state     *models.AppState
	scrollPos int
	maxScroll int
}

// NewPreviewPanel creates a new preview panel
func NewPreviewPanel(state *models.AppState) *PreviewPanel {
	return &PreviewPanel{
		state:     state,
		scrollPos: 0,
	}
}

// Init initializes the preview panel
func (p *PreviewPanel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the preview panel
func (p *PreviewPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if p.state.FocusedPanel != models.PreviewPanel {
			return p, nil
		}

		return p, p.handleKeyMsg(msg)

	case models.PreviewContentMsg:
		p.state.PreviewState.Content = msg.Content
		p.state.PreviewState.ContentType = msg.ContentType
		p.state.PreviewState.FilePath = msg.FilePath
		p.state.PreviewState.CommitHash = msg.CommitHash
		p.updateScrollBounds()
		return p, nil

	case models.RepositorySelectedMsg:
		// Update preview to show repository information
		p.state.PreviewState.ContentType = models.PreviewStatus
		p.updatePreviewContent()
		return p, nil
	}

	return p, nil
}

// View renders the preview panel
func (p *PreviewPanel) View() string {
	switch p.state.PreviewState.ContentType {
	case models.PreviewFile:
		return p.renderFilePreview()
	case models.PreviewCommit:
		return p.renderCommitPreview()
	case models.PreviewStatus:
		return p.renderStatusPreview()
	case models.PreviewHelp:
		return p.renderHelpPreview()
	default:
		return p.renderDefaultPreview()
	}
}

// handleKeyMsg handles keyboard input for the preview panel
func (p *PreviewPanel) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "up", "k":
		if p.scrollPos > 0 {
			p.scrollPos--
		}

	case "down", "j":
		if p.scrollPos < p.maxScroll {
			p.scrollPos++
		}

	case "page_up":
		p.scrollPos = max(0, p.scrollPos-10)

	case "page_down":
		p.scrollPos = min(p.maxScroll, p.scrollPos+10)

	case "home":
		p.scrollPos = 0

	case "end":
		p.scrollPos = p.maxScroll
	}

	return nil
}

// renderFilePreview renders a file preview
func (p *PreviewPanel) renderFilePreview() string {
	if p.state.PreviewState.Content == "" {
		return styles.MutedStyle.Render("No file selected for preview")
	}

	var content strings.Builder

	// File header
	header := styles.SubHeaderStyle.Render("File: " + p.state.PreviewState.FilePath)
	content.WriteString(header)
	content.WriteString("\n\n")

	// File content (scrollable)
	lines := strings.Split(p.state.PreviewState.Content, "\n")
	visibleLines := p.getVisibleLines(lines)

	for _, line := range visibleLines {
		content.WriteString(styles.PreviewCodeStyle.Render(line))
		content.WriteString("\n")
	}

	// Scroll indicator
	if p.maxScroll > 0 {
		scrollInfo := p.renderScrollIndicator()
		content.WriteString("\n")
		content.WriteString(scrollInfo)
	}

	return content.String()
}

// renderCommitPreview renders a commit preview
func (p *PreviewPanel) renderCommitPreview() string {
	if p.state.PreviewState.Content == "" {
		return styles.MutedStyle.Render("No commit selected for preview")
	}

	var content strings.Builder

	// Commit header
	header := styles.SubHeaderStyle.Render("Commit: " + p.state.PreviewState.CommitHash)
	content.WriteString(header)
	content.WriteString("\n\n")

	// Commit content (scrollable)
	lines := strings.Split(p.state.PreviewState.Content, "\n")
	visibleLines := p.getVisibleLines(lines)

	for _, line := range visibleLines {
		content.WriteString(styles.PreviewCodeStyle.Render(line))
		content.WriteString("\n")
	}

	// Scroll indicator
	if p.maxScroll > 0 {
		scrollInfo := p.renderScrollIndicator()
		content.WriteString("\n")
		content.WriteString(scrollInfo)
	}

	return content.String()
}

// renderStatusPreview renders a status preview for the selected repository
func (p *PreviewPanel) renderStatusPreview() string {
	if p.state.SelectedRepo == "" {
		return p.renderDefaultPreview()
	}

	repo := p.state.GetSelectedRepository()
	if repo == nil {
		return styles.MutedStyle.Render("Loading repository information...")
	}

	var content strings.Builder

	// Repository overview
	header := styles.SubHeaderStyle.Render("Repository Overview")
	content.WriteString(header)
	content.WriteString("\n\n")

	// Basic info
	info := []string{
		"Name: " + repo.Alias,
		"Path: " + repo.Path,
	}

	if repo.Status != nil {
		info = append(info,
			"Branch: "+repo.Status.Branch,
			"Status: "+repo.Status.Workspace.String(),
			"Sync: "+repo.Status.SyncStatus.String(),
		)

		if repo.Status.FilesChanged > 0 {
			info = append(info, fmt.Sprintf("Files changed: %d", repo.Status.FilesChanged))
		}
	}

	for _, line := range info {
		content.WriteString(styles.BodyStyle.Render(line))
		content.WriteString("\n")
	}

	// Recent activity (placeholder)
	content.WriteString("\n")
	content.WriteString(styles.SubHeaderStyle.Render("Recent Activity"))
	content.WriteString("\n")
	content.WriteString(styles.MutedStyle.Render("Recent commits and changes would be displayed here"))

	return content.String()
}

// renderHelpPreview renders the help preview
func (p *PreviewPanel) renderHelpPreview() string {
	help := []string{
		"Welcome to gman TUI Dashboard!",
		"",
		"This dashboard provides a unified interface",
		"for managing multiple Git repositories.",
		"",
		"Key Features:",
		"• Repository list with status indicators",
		"• Detailed status information",
		"• Integrated search (Phase 5.1 fzf)",
		"• File and commit previews",
		"",
		"Navigation:",
		"• Tab/Shift+Tab: Switch panels",
		"• 1-4: Jump to specific panel",
		"• Arrow keys: Navigate within panels",
		"• Enter: Select items",
		"• ?: Toggle help",
		"• q: Quit",
		"",
		"Select a repository to get started!",
	}

	var content strings.Builder
	for _, line := range help {
		if line == "" {
			content.WriteString("\n")
		} else if strings.HasSuffix(line, ":") {
			content.WriteString(styles.SubHeaderStyle.Render(line))
			content.WriteString("\n")
		} else if strings.HasPrefix(line, "•") {
			content.WriteString(styles.BodyStyle.Render(line))
			content.WriteString("\n")
		} else {
			content.WriteString(styles.BodyStyle.Render(line))
			content.WriteString("\n")
		}
	}

	return content.String()
}

// renderDefaultPreview renders the default preview
func (p *PreviewPanel) renderDefaultPreview() string {
	return p.renderHelpPreview()
}

// getVisibleLines returns the lines that should be visible based on scroll position
func (p *PreviewPanel) getVisibleLines(lines []string) []string {
	// Calculate visible area (assuming ~15 lines visible)
	visibleHeight := 15

	start := p.scrollPos
	end := min(len(lines), start+visibleHeight)

	if start >= len(lines) {
		return []string{}
	}

	return lines[start:end]
}

// updateScrollBounds updates the maximum scroll position
func (p *PreviewPanel) updateScrollBounds() {
	if p.state.PreviewState.Content == "" {
		p.maxScroll = 0
		return
	}

	lines := strings.Split(p.state.PreviewState.Content, "\n")
	visibleHeight := 15

	p.maxScroll = max(0, len(lines)-visibleHeight)

	// Adjust current scroll if needed
	if p.scrollPos > p.maxScroll {
		p.scrollPos = p.maxScroll
	}
}

// renderScrollIndicator renders a scroll position indicator
func (p *PreviewPanel) renderScrollIndicator() string {
	if p.maxScroll == 0 {
		return ""
	}

	percentage := int((float64(p.scrollPos) / float64(p.maxScroll)) * 100)
	indicator := fmt.Sprintf("Scroll: %d%% (%d/%d)", percentage, p.scrollPos, p.maxScroll)

	return styles.MutedStyle.Render(indicator)
}

// updatePreviewContent updates the preview content based on current selection
func (p *PreviewPanel) updatePreviewContent() {
	// This would be called when repository selection changes
	// to update the preview content accordingly
	p.updateScrollBounds()
}
