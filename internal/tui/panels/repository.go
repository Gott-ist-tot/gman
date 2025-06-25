package panels

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gman/internal/di"
	"gman/internal/git"
	"gman/internal/tui/models"
	"gman/internal/tui/styles"
	"gman/pkg/types"
)

// RepositoryPanel handles the repository list display and interaction
type RepositoryPanel struct {
	state  *models.AppState
	gitMgr *git.Manager

	// UI state
	cursor      int
	filterInput string
	filtering   bool
	repos       []models.RepoDisplayItem
	lastUpdate  time.Time
}

// NewRepositoryPanel creates a new repository panel
func NewRepositoryPanel(state *models.AppState) *RepositoryPanel {
	panel := &RepositoryPanel{
		state:      state,
		gitMgr:     di.GitManager(),
		cursor:     0,
		repos:      make([]models.RepoDisplayItem, 0),
		lastUpdate: time.Now(),
	}

	panel.refreshRepositoryList()
	return panel
}

// Init initializes the repository panel
func (r *RepositoryPanel) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Load repository status
	cmds = append(cmds, r.loadRepositoryStatus())

	// If a repository is auto-selected, trigger selection message
	if r.state.SelectedRepo != "" && len(r.repos) > 0 {
		for _, repo := range r.repos {
			if repo.Alias == r.state.SelectedRepo {
				cmds = append(cmds, models.RepositorySelectedCmd(repo.Alias, repo.Path))
				break
			}
		}
	}

	return tea.Batch(cmds...)
}

// Update handles messages for the repository panel
func (r *RepositoryPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if r.state.FocusedPanel != models.RepositoryPanel {
			return r, nil
		}

		cmd := r.handleKeyMsg(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case models.RepositoryStatusMsg:
		r.updateRepositoryStatus(msg.Alias, msg.Status)

	case models.FilterChangedMsg:
		r.state.RepositoryListState.FilterText = msg.FilterText
		r.state.RepositoryListState.FilterGroup = msg.FilterGroup
		r.refreshRepositoryList()

	case models.SortChangedMsg:
		r.state.RepositoryListState.SortBy = msg.SortBy
		r.refreshRepositoryList()

	case models.RefreshMsg:
		if msg.Force || time.Since(r.lastUpdate) > time.Minute*5 {
			cmds = append(cmds, r.loadRepositoryStatus())
		}
	}

	return r, tea.Batch(cmds...)
}

// View renders the repository panel
func (r *RepositoryPanel) View() string {
	if len(r.repos) == 0 {
		return r.renderEmptyState()
	}

	var content strings.Builder

	// Filter input if active
	if r.filtering {
		filterStyle := styles.InputFocusedStyle.Width(30)
		content.WriteString(filterStyle.Render("Filter: " + r.filterInput))
		content.WriteString("\n\n")
	}

	// Repository list
	for i, repo := range r.repos {
		line := r.renderRepositoryItem(repo, i == r.cursor)
		content.WriteString(line)
		content.WriteString("\n")
	}

	// Footer with counts and sort info
	footer := r.renderFooter()
	content.WriteString("\n")
	content.WriteString(footer)

	return content.String()
}

// handleKeyMsg handles keyboard input for the repository panel
func (r *RepositoryPanel) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	if r.filtering {
		return r.handleFilterInput(msg)
	}

	switch msg.String() {
	case "up", "k":
		if r.cursor > 0 {
			r.cursor--
		}

	case "down", "j":
		if r.cursor < len(r.repos)-1 {
			r.cursor++
		}

	case "enter":
		return r.selectCurrentRepository()

	case "/":
		r.filtering = true
		r.filterInput = ""

	case "s":
		return r.cycleSortOrder()

	case "g":
		return r.showGroupFilter()

	case "r":
		return models.RefreshCmd(true)

	case "home":
		r.cursor = 0

	case "end":
		r.cursor = len(r.repos) - 1

	case "page_up":
		r.cursor = max(0, r.cursor-10)

	case "page_down":
		r.cursor = min(len(r.repos)-1, r.cursor+10)
	}

	return nil
}

// handleFilterInput handles input when in filtering mode
func (r *RepositoryPanel) handleFilterInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		r.filtering = false
		return models.FilterChangedCmd(r.filterInput, r.state.RepositoryListState.FilterGroup)

	case "esc":
		r.filtering = false
		r.filterInput = ""

	case "backspace":
		if len(r.filterInput) > 0 {
			r.filterInput = r.filterInput[:len(r.filterInput)-1]
		}

	default:
		if len(msg.String()) == 1 {
			r.filterInput += msg.String()
		}
	}

	return nil
}

// selectCurrentRepository selects the current repository
func (r *RepositoryPanel) selectCurrentRepository() tea.Cmd {
	if r.cursor >= 0 && r.cursor < len(r.repos) {
		repo := r.repos[r.cursor]
		r.state.SelectedRepo = repo.Alias

		// Track recent usage
		r.state.ConfigManager.TrackRecentUsage(repo.Alias)

		return models.RepositorySelectedCmd(repo.Alias, repo.Path)
	}
	return nil
}

// cycleSortOrder cycles through different sort orders
func (r *RepositoryPanel) cycleSortOrder() tea.Cmd {
	current := r.state.RepositoryListState.SortBy
	next := (current + 1) % 4 // We have 4 sort types
	return models.SortChangedCmd(models.SortType(next))
}

// showGroupFilter shows group filtering options
func (r *RepositoryPanel) showGroupFilter() tea.Cmd {
	// This could open a submenu or cycle through groups
	// For now, just clear the group filter
	if r.state.RepositoryListState.FilterGroup == "" {
		// Set to first available group
		for groupName := range r.state.Groups {
			return models.FilterChangedCmd(r.state.RepositoryListState.FilterText, groupName)
		}
	} else {
		// Clear group filter
		return models.FilterChangedCmd(r.state.RepositoryListState.FilterText, "")
	}
	return nil
}

// loadRepositoryStatus loads status for all repositories
func (r *RepositoryPanel) loadRepositoryStatus() tea.Cmd {
	return tea.Batch(r.loadRepositoryStatusBatch()...)
}

// loadRepositoryStatusBatch creates commands to load status for all repos
func (r *RepositoryPanel) loadRepositoryStatusBatch() []tea.Cmd {
	var cmds []tea.Cmd

	for alias, path := range r.state.Repositories {
		cmds = append(cmds, r.loadSingleRepositoryStatus(alias, path))
	}

	return cmds
}

// loadSingleRepositoryStatus loads status for a single repository
func (r *RepositoryPanel) loadSingleRepositoryStatus(alias, path string) tea.Cmd {
	return func() tea.Msg {
		status := r.gitMgr.GetRepoStatus(alias, path)
		return models.RepositoryStatusMsg{
			Alias:  alias,
			Status: &status,
			Error:  status.Error,
		}
	}
}

// updateRepositoryStatus updates the status for a repository
func (r *RepositoryPanel) updateRepositoryStatus(alias string, status *types.RepoStatus) {
	for i, repo := range r.repos {
		if repo.Alias == alias {
			r.repos[i].Status = status
			break
		}
	}
	r.lastUpdate = time.Now()
}

// refreshRepositoryList refreshes the repository list based on current filters
func (r *RepositoryPanel) refreshRepositoryList() {
	r.repos = r.repos[:0] // Clear slice but keep capacity

	// Ensure repositories are loaded from config
	if r.state.Repositories == nil || len(r.state.Repositories) == 0 {
		// Try to reload from config manager
		if r.state.ConfigManager != nil {
			config := r.state.ConfigManager.GetConfig()
			if config != nil && config.Repositories != nil {
				r.state.Repositories = config.Repositories
			}
		}
	}

	// Get repositories from state
	for alias, path := range r.state.Repositories {
		// Apply group filter
		if r.state.RepositoryListState.FilterGroup != "" {
			groupRepos, err := r.state.ConfigManager.GetGroupRepositories(r.state.RepositoryListState.FilterGroup)
			if err != nil || groupRepos[alias] == "" {
				continue
			}
		}

		// Apply text filter
		if r.state.RepositoryListState.FilterText != "" {
			if !strings.Contains(strings.ToLower(alias), strings.ToLower(r.state.RepositoryListState.FilterText)) {
				continue
			}
		}

		// Get recent access time
		recentEntries := r.state.ConfigManager.GetRecentUsage()
		var lastAccess time.Time
		for _, entry := range recentEntries {
			if entry.Alias == alias {
				lastAccess = entry.AccessTime
				break
			}
		}

		repo := models.RepoDisplayItem{
			Alias:        alias,
			Path:         path,
			LastAccessed: lastAccess,
			IsSelected:   alias == r.state.SelectedRepo,
		}

		r.repos = append(r.repos, repo)
	}

	// Sort repositories
	r.sortRepositories()

	// Update state
	r.state.RepositoryListState.VisibleRepos = r.repos

	// Adjust cursor if needed
	if r.cursor >= len(r.repos) {
		r.cursor = max(0, len(r.repos)-1)
	}

	// Auto-select first repository if none is selected and repositories exist
	if r.state.SelectedRepo == "" && len(r.repos) > 0 {
		r.state.SelectedRepo = r.repos[0].Alias
		r.repos[0].IsSelected = true

		// Track recent usage for auto-selected repository
		r.state.ConfigManager.TrackRecentUsage(r.repos[0].Alias)
	}
}

// sortRepositories sorts the repository list based on the current sort order
func (r *RepositoryPanel) sortRepositories() {
	sort.Slice(r.repos, func(i, j int) bool {
		switch r.state.RepositoryListState.SortBy {
		case models.SortByName:
			return r.repos[i].Alias < r.repos[j].Alias
		case models.SortByLastUsed:
			return r.repos[i].LastAccessed.After(r.repos[j].LastAccessed)
		case models.SortByStatus:
			return r.getStatusPriority(r.repos[i]) < r.getStatusPriority(r.repos[j])
		case models.SortByModified:
			if r.repos[i].Status != nil && r.repos[j].Status != nil {
				return r.repos[i].Status.CommitTime.After(r.repos[j].Status.CommitTime)
			}
			return false
		default:
			return r.repos[i].Alias < r.repos[j].Alias
		}
	})
}

// getStatusPriority returns a priority value for sorting by status
func (r *RepositoryPanel) getStatusPriority(repo models.RepoDisplayItem) int {
	if repo.Status == nil {
		return 5 // Unknown status gets lowest priority
	}

	switch repo.Status.Workspace {
	case types.Dirty:
		return 1 // Dirty repos first
	case types.Clean:
		if repo.Status.SyncStatus.Behind > 0 {
			return 2
		} else if repo.Status.SyncStatus.Ahead > 0 {
			return 3
		} else {
			return 4 // Up to date
		}
	default:
		return 5
	}
}

// renderRepositoryItem renders a single repository item
func (r *RepositoryPanel) renderRepositoryItem(repo models.RepoDisplayItem, selected bool) string {
	var style lipgloss.Style
	cursor := ""

	if selected {
		style = styles.ListItemSelectedStyle
		cursor = "▶ "
	} else {
		style = styles.ListItemStyle
		cursor = "  "
	}

	// Status indicator with better icons
	statusIcon := "⚪"
	statusColor := styles.ColorTextMuted

	if repo.Status != nil {
		switch repo.Status.Workspace {
		case types.Clean:
			statusIcon = "🟢"
			statusColor = styles.ColorStatusClean
		case types.Dirty:
			statusIcon = "🔴"
			statusColor = styles.ColorStatusDirty
		}

		// Add sync status with better indicators
		if repo.Status.SyncStatus.Ahead > 0 && repo.Status.SyncStatus.Behind > 0 {
			statusIcon = "🔀" // Diverged
		} else if repo.Status.SyncStatus.Ahead > 0 {
			statusIcon = "⬆️ " // Ahead
		} else if repo.Status.SyncStatus.Behind > 0 {
			statusIcon = "⬇️ " // Behind
		}
	}

	// Repository name with better styling
	nameStyle := lipgloss.NewStyle().
		Foreground(styles.ColorTextPrimary).
		Bold(selected)
	if selected {
		nameStyle = nameStyle.Foreground(styles.ColorPrimary)
	}
	name := nameStyle.Render(repo.Alias)

	// Status with color
	statusStyle := lipgloss.NewStyle().Foreground(statusColor)
	status := statusStyle.Render(statusIcon)

	// Build the line with better spacing
	line := fmt.Sprintf("%s%s %s", cursor, status, name)

	// Add additional info with better formatting
	if repo.Status != nil && repo.Status.FilesChanged > 0 {
		filesInfo := lipgloss.NewStyle().
			Foreground(styles.ColorTextSecondary).
			Italic(true).
			Render(fmt.Sprintf(" (%d files)", repo.Status.FilesChanged))
		line += filesInfo
	}

	return style.Render(line)
}

// renderEmptyState renders the empty state when no repositories are found
func (r *RepositoryPanel) renderEmptyState() string {
	// Check if repositories exist in config but are filtered out
	totalRepos := len(r.state.Repositories)
	if totalRepos > 0 {
		// Repositories exist but are filtered
		filterInfo := ""
		if r.state.RepositoryListState.FilterText != "" {
			filterInfo += fmt.Sprintf("Filter: '%s'\n", r.state.RepositoryListState.FilterText)
		}
		if r.state.RepositoryListState.FilterGroup != "" {
			filterInfo += fmt.Sprintf("Group: '%s'\n", r.state.RepositoryListState.FilterGroup)
		}
		return styles.MutedStyle.Render(fmt.Sprintf("No repositories match current filters.\n%s\nPress 'g' to clear group filter or '/' to change text filter.\n\nTotal repositories: %d", filterInfo, totalRepos))
	}

	// No repositories configured at all
	return styles.MutedStyle.Render("No repositories configured.\n\nAdd repositories with:\n  gman add <alias> <path>\n\nOr quit (q) and run 'gman list' to see available commands.")
}

// renderFooter renders the footer with repository count and sort info
func (r *RepositoryPanel) renderFooter() string {
	// Repository count with status breakdown
	count := fmt.Sprintf("📁 %d repos", len(r.repos))
	if len(r.repos) != len(r.state.Repositories) {
		count = fmt.Sprintf("📁 %d/%d repos", len(r.repos), len(r.state.Repositories))
	}

	// Sort information with icon
	sortInfo := fmt.Sprintf("📊 %s", r.state.RepositoryListState.SortBy.String())

	// Selected repository indicator
	selectedInfo := ""
	if r.state.SelectedRepo != "" {
		selectedInfo = fmt.Sprintf("👆 %s", r.state.SelectedRepo)
	}

	// Build footer components
	components := []string{
		styles.MutedStyle.Render(count),
		styles.MutedStyle.Render(" • "),
		styles.MutedStyle.Render(sortInfo),
	}

	if selectedInfo != "" {
		components = append(components,
			styles.MutedStyle.Render(" • "),
			styles.MutedStyle.Render(selectedInfo),
		)
	}

	if r.state.RepositoryListState.FilterGroup != "" {
		components = append(components,
			styles.MutedStyle.Render(" • 🏷️  "),
			styles.MutedStyle.Render(r.state.RepositoryListState.FilterGroup),
		)
	}

	footer := lipgloss.JoinHorizontal(lipgloss.Left, components...)

	// Add keyboard shortcuts hint
	shortcuts := styles.MutedStyle.Render("↑↓ Navigate • Enter Select • / Filter • s Sort")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		footer,
		shortcuts,
	)
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
