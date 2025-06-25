package panels

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"gman/internal/tui/models"
	"gman/internal/tui/styles"
	"gman/pkg/types"
)

// StatusPanel displays detailed status information for the selected repository
type StatusPanel struct {
	state *models.AppState
}

// NewStatusPanel creates a new status panel
func NewStatusPanel(state *models.AppState) *StatusPanel {
	return &StatusPanel{
		state: state,
	}
}

// Init initializes the status panel
func (s *StatusPanel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the status panel
func (s *StatusPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if s.state.FocusedPanel != models.StatusPanel {
			return s, nil
		}
		
		return s, s.handleKeyMsg(msg)

	case models.RepositorySelectedMsg:
		// Panel will be updated when status is loaded
		return s, nil

	case models.RepositoryStatusMsg:
		if msg.Alias == s.state.SelectedRepo {
			s.state.UpdateRepositoryData(msg.Alias, msg.Status)
		}
		return s, nil
	}

	return s, nil
}

// View renders the status panel
func (s *StatusPanel) View() string {
	if s.state.SelectedRepo == "" {
		return s.renderNoSelection()
	}

	repo := s.state.GetSelectedRepository()
	if repo == nil || repo.Status == nil {
		return s.renderLoading()
	}

	return s.renderStatus(repo)
}

// handleKeyMsg handles keyboard input for the status panel
func (s *StatusPanel) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "e":
		s.state.StatusState.ShowExtended = !s.state.StatusState.ShowExtended

	case "b":
		s.state.StatusState.ShowBranches = !s.state.StatusState.ShowBranches

	case "r":
		return models.RefreshCmd(true)

	case "a":
		s.state.StatusState.AutoRefresh = !s.state.StatusState.AutoRefresh
	}

	return nil
}

// renderNoSelection renders the view when no repository is selected
func (s *StatusPanel) renderNoSelection() string {
	help := "Select a repository to view its status\n\n"
	help += "Navigation:\n"
	help += "• Arrow keys: Navigate repository list\n"
	help += "• Enter: Select repository\n"
	help += "• r: Refresh status\n"
	help += "• e: Toggle extended view\n"
	help += "• b: Toggle branch information\n"
	content := styles.MutedStyle.Render(help)
	return content
}

// renderLoading renders the loading state
func (s *StatusPanel) renderLoading() string {
	repoName := s.state.SelectedRepo
	if repoName == "" {
		repoName = "repository"
	}
	
	content := fmt.Sprintf("Loading status for: %s\n\n", repoName)
	content += "Fetching:\n"
	content += "• Working directory status\n"
	content += "• Branch information\n"
	content += "• Remote sync status\n"
	content += "• Recent commits\n\n"
	content += "Please wait..."
	
	return styles.MutedStyle.Render(content)
}

// renderStatus renders the detailed status for a repository
func (s *StatusPanel) renderStatus(repo *models.RepoDisplayItem) string {
	var content strings.Builder

	// Repository header
	header := styles.HeaderStyle.Render(repo.Alias)
	content.WriteString(header)
	content.WriteString("\n")

	// Path
	pathStyle := styles.MutedStyle
	content.WriteString(pathStyle.Render(repo.Path))
	content.WriteString("\n\n")

	// Status overview
	content.WriteString(s.renderStatusOverview(repo.Status))
	content.WriteString("\n")

	// Extended information if enabled
	if s.state.StatusState.ShowExtended {
		content.WriteString(s.renderExtendedInfo(repo.Status))
	}

	// Branch information if enabled
	if s.state.StatusState.ShowBranches {
		content.WriteString(s.renderBranchInfo(repo.Status))
	}

	// Footer
	content.WriteString(s.renderStatusFooter())

	return content.String()
}

// renderStatusOverview renders the basic status overview
func (s *StatusPanel) renderStatusOverview(status *types.RepoStatus) string {
	var lines []string

	// Workspace status
	workspaceIcon := styles.GetStatusIcon(status.Workspace.String())
	workspaceStyle := styles.GetStatusStyle(status.Workspace.String())
	workspaceLine := fmt.Sprintf("%s Workspace: %s", 
		workspaceIcon, 
		workspaceStyle.Render(status.Workspace.String()))
	lines = append(lines, workspaceLine)

	// Sync status
	syncIcon := styles.GetStatusIcon(status.SyncStatus.String())
	syncStyle := styles.GetStatusStyle(status.SyncStatus.String())
	syncLine := fmt.Sprintf("%s Sync: %s", 
		syncIcon, 
		syncStyle.Render(status.SyncStatus.String()))
	lines = append(lines, syncLine)

	// Current branch
	branchLine := fmt.Sprintf("🌿 Branch: %s", 
		styles.BodyStyle.Render(status.Branch))
	lines = append(lines, branchLine)

	// File changes if any
	if status.FilesChanged > 0 {
		filesLine := fmt.Sprintf("📝 Files changed: %d", status.FilesChanged)
		lines = append(lines, styles.StatusDirtyStyle.Render(filesLine))
	}

	// Commits ahead/behind
	if status.SyncStatus.Ahead > 0 {
		aheadLine := fmt.Sprintf("↑ Ahead: %d commits", status.SyncStatus.Ahead)
		lines = append(lines, styles.StatusAheadStyle.Render(aheadLine))
	}

	if status.SyncStatus.Behind > 0 {
		behindLine := fmt.Sprintf("↓ Behind: %d commits", status.SyncStatus.Behind)
		lines = append(lines, styles.StatusBehindStyle.Render(behindLine))
	}

	return strings.Join(lines, "\n")
}

// renderExtendedInfo renders extended repository information
func (s *StatusPanel) renderExtendedInfo(status *types.RepoStatus) string {
	var lines []string

	lines = append(lines, "\n" + styles.SubHeaderStyle.Render("Extended Information:"))

	// Last commit time
	if !status.CommitTime.IsZero() {
		timeSince := time.Since(status.CommitTime)
		timeStr := formatDuration(timeSince)
		commitLine := fmt.Sprintf("⏰ Last commit: %s ago", timeStr)
		lines = append(lines, styles.BodyStyle.Render(commitLine))
	}

	// Remote information (placeholder - not in RepoStatus yet)
	// TODO: Add remote information to RepoStatus struct
	remoteLine := "🌐 Remote: origin" // Placeholder
	lines = append(lines, styles.BodyStyle.Render(remoteLine))

	// Stash information (if available)
	stashLine := "📦 Stashes: 0" // TODO: Add stash count to RepoStatus
	lines = append(lines, styles.BodyStyle.Render(stashLine))

	return strings.Join(lines, "\n")
}

// renderBranchInfo renders branch information
func (s *StatusPanel) renderBranchInfo(status *types.RepoStatus) string {
	var lines []string

	lines = append(lines, "\n" + styles.SubHeaderStyle.Render("Branch Information:"))

	// Current branch
	currentLine := fmt.Sprintf("• %s (current)", status.Branch)
	lines = append(lines, styles.StatusCleanStyle.Render(currentLine))

	// TODO: Add other branches information
	// This would require extending the RepoStatus struct or making additional Git calls

	return strings.Join(lines, "\n")
}

// renderStatusFooter renders the footer with last update time and options
func (s *StatusPanel) renderStatusFooter() string {
	var parts []string

	// Last refresh time
	if !s.state.StatusState.LastRefresh.IsZero() {
		timeSince := time.Since(s.state.StatusState.LastRefresh)
		refreshStr := fmt.Sprintf("Updated %s ago", formatDuration(timeSince))
		parts = append(parts, refreshStr)
	}

	// Auto refresh indicator
	if s.state.StatusState.AutoRefresh {
		parts = append(parts, "Auto-refresh ON")
	}

	// Keyboard shortcuts
	shortcuts := []string{
		"e: Extended",
		"b: Branches", 
		"r: Refresh",
		"a: Auto-refresh",
	}
	parts = append(parts, strings.Join(shortcuts, " • "))

	footer := strings.Join(parts, " | ")
	return "\n" + styles.MutedStyle.Render(footer)
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}