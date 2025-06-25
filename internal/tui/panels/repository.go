package panels

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gman/internal/git"
	"gman/internal/tui/models"
	"gman/internal/tui/styles"
	"gman/pkg/types"
)

// RepositoryPanel handles the repository list display and interaction
type RepositoryPanel struct {
	state   *models.AppState
	gitMgr  *git.Manager
	
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
		state:   state,
		gitMgr:  git.NewManager(),
		cursor:  0,
		repos:   make([]models.RepoDisplayItem, 0),
		lastUpdate: time.Now(),
	}
	
	panel.refreshRepositoryList()
	return panel
}

// Init initializes the repository panel
func (r *RepositoryPanel) Init() tea.Cmd {
	return r.loadRepositoryStatus()
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
	if selected {
		style = styles.ListItemSelectedStyle
	} else {
		style = styles.ListItemStyle
	}
	
	// Status indicator
	statusIcon := "❓"
	statusColor := styles.ColorTextMuted
	
	if repo.Status != nil {
		switch repo.Status.Workspace {
		case types.Clean:
			statusIcon = "✅"
			statusColor = styles.ColorStatusClean
		case types.Dirty:
			statusIcon = "📝"
			statusColor = styles.ColorStatusDirty
		}
		
		// Add sync status
		if repo.Status.SyncStatus.Ahead > 0 {
			statusIcon += "↑"
		}
		if repo.Status.SyncStatus.Behind > 0 {
			statusIcon += "↓"
		}
		if repo.Status.SyncStatus.Ahead > 0 && repo.Status.SyncStatus.Behind > 0 {
			statusIcon = "📝↕" // Replace with diverged icon
		}
	}
	
	// Repository name
	nameStyle := lipgloss.NewStyle().Foreground(styles.ColorTextPrimary).Bold(selected)
	name := nameStyle.Render(repo.Alias)
	
	// Status with color
	statusStyle := lipgloss.NewStyle().Foreground(statusColor)
	status := statusStyle.Render(statusIcon)
	
	// Build the line
	line := fmt.Sprintf("%s %s", status, name)
	
	// Add additional info if there's space
	if repo.Status != nil && repo.Status.FilesChanged > 0 {
		filesInfo := lipgloss.NewStyle().
			Foreground(styles.ColorTextSecondary).
			Render(fmt.Sprintf(" (%d files)", repo.Status.FilesChanged))
		line += filesInfo
	}
	
	return style.Render(line)
}

// renderEmptyState renders the empty state when no repositories are found
func (r *RepositoryPanel) renderEmptyState() string {
	return styles.MutedStyle.Render("No repositories found.\nAdd repositories with 'gman add'.")
}

// renderFooter renders the footer with repository count and sort info
func (r *RepositoryPanel) renderFooter() string {
	count := fmt.Sprintf("%d repos", len(r.repos))
	sortInfo := fmt.Sprintf("Sort: %s", r.state.RepositoryListState.SortBy.String())
	
	footer := lipgloss.JoinHorizontal(
		lipgloss.Left,
		styles.MutedStyle.Render(count),
		styles.MutedStyle.Render(" • "),
		styles.MutedStyle.Render(sortInfo),
	)
	
	if r.state.RepositoryListState.FilterGroup != "" {
		footer = lipgloss.JoinHorizontal(
			lipgloss.Left,
			footer,
			styles.MutedStyle.Render(" • Group: "),
			styles.MutedStyle.Render(r.state.RepositoryListState.FilterGroup),
		)
	}
	
	return footer
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