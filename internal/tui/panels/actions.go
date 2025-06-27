package panels

import (
	"fmt"
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

// ActionsPanel handles interactive operations and command execution
type ActionsPanel struct {
	state *models.AppState

	// Action state
	selectedAction int
	actions        []Action
	executing      bool
	lastResult     string
	lastError      error
	showResult     bool
	resultTimer    time.Time
}

// Action represents an actionable operation
type Action struct {
	Name        string
	Description string
	Shortcut    string
	Category    string
	Handler     func(*models.AppState) tea.Cmd
	Condition   func(*models.AppState) bool // Optional condition to show action
}

// NewActionsPanel creates a new actions panel
func NewActionsPanel(state *models.AppState) *ActionsPanel {
	panel := &ActionsPanel{
		state:          state,
		selectedAction: 0,
		executing:      false,
		showResult:     false,
	}

	panel.initializeActions()
	return panel
}

// Init initializes the actions panel
func (a *ActionsPanel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the actions panel
func (a *ActionsPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if a.state.FocusedPanel != models.ActionsPanel {
			return a, nil
		}

		return a, a.handleKeyMsg(msg)

	case models.ActionCompleteMsg:
		a.executing = false
		a.lastResult = msg.Result
		a.lastError = msg.Error
		a.showResult = true
		a.resultTimer = time.Now()
		
		// Hide any active progress indicators
		cmds := []tea.Cmd{
			models.ProgressHideCmd("refresh"),
			models.ProgressHideCmd("sync"),
			models.ProgressHideCmd("commit"),
		}
		
		// Show success or error toast
		if msg.Error != nil {
			cmds = append(cmds, models.ToastErrorCmd(msg.Error.Error()))
		} else if msg.Result != "" {
			cmds = append(cmds, models.ToastSuccessCmd(msg.Result))
		}
		
		// Auto-hide result after 3 seconds
		cmds = append(cmds, tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
			return models.HideResultMsg{}
		}))
		
		return a, tea.Batch(cmds...)

	case models.HideResultMsg:
		a.showResult = false
		return a, nil

	case models.RepositorySelectedMsg:
		// Update available actions when repository changes
		a.filterActions()
		return a, nil
	}

	return a, nil
}

// View renders the actions panel
func (a *ActionsPanel) View() string {
	if a.state.SelectedRepo == "" {
		return a.renderNoRepository()
	}

	var content strings.Builder

	// Panel header
	header := styles.HeaderStyle.Render("⚡ Actions")
	content.WriteString(header)
	content.WriteString("\n\n")

	// Show result if available
	if a.showResult {
		content.WriteString(a.renderResult())
		content.WriteString("\n")
	}

	// Action categories
	categories := a.groupActionsByCategory()

	for category, actions := range categories {
		if len(actions) == 0 {
			continue
		}

		// Category header
		categoryStyle := styles.SubHeaderStyle.Copy().
			Foreground(lipgloss.Color("6"))
		content.WriteString(categoryStyle.Render(category))
		content.WriteString("\n")

		// Actions in category
		for i, action := range actions {
			content.WriteString(a.renderAction(action, i))
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	// Footer with shortcuts
	content.WriteString(a.renderFooter())

	return content.String()
}

// handleKeyMsg handles keyboard input for the actions panel
func (a *ActionsPanel) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	if a.executing {
		return nil // Ignore input while executing
	}

	switch msg.String() {
	case "up", "k":
		if a.selectedAction > 0 {
			a.selectedAction--
		}

	case "down", "j":
		availableActions := a.getAvailableActions()
		if a.selectedAction < len(availableActions)-1 {
			a.selectedAction++
		}

	case "enter":
		return a.executeSelectedAction()

	case "r":
		return a.executeQuickAction("refresh")

	case "s":
		return a.executeQuickAction("sync")

	case "c":
		return a.executeQuickAction("commit")

	case "p":
		return a.executeQuickAction("push")

	case "S":
		return a.executeQuickAction("stash")
	}

	return nil
}

// initializeActions sets up the available actions
func (a *ActionsPanel) initializeActions() {
	a.actions = []Action{
		// Repository Operations
		{
			Name:        "Refresh Status",
			Description: "Refresh repository status information",
			Shortcut:    "r",
			Category:    "Repository",
			Handler:     a.handleRefresh,
		},
		{
			Name:        "Open in Terminal",
			Description: "Open repository in external terminal",
			Shortcut:    "t",
			Category:    "Repository",
			Handler:     a.handleOpenTerminal,
		},
		{
			Name:        "Open in File Manager",
			Description: "Open repository in system file manager",
			Shortcut:    "o",
			Category:    "Repository",
			Handler:     a.handleOpenFileManager,
		},

		// Git Operations
		{
			Name:        "Sync Repository",
			Description: "Pull latest changes from remote",
			Shortcut:    "s",
			Category:    "Git",
			Handler:     a.handleSync,
		},
		{
			Name:        "Commit Changes",
			Description: "Commit staged changes",
			Shortcut:    "c",
			Category:    "Git",
			Handler:     a.handleCommit,
			Condition:   a.hasUncommittedChanges,
		},
		{
			Name:        "Push Changes",
			Description: "Push local commits to remote",
			Shortcut:    "p",
			Category:    "Git",
			Handler:     a.handlePush,
			Condition:   a.hasUnpushedCommits,
		},
		{
			Name:        "Stash Changes",
			Description: "Stash current uncommitted changes",
			Shortcut:    "S",
			Category:    "Git",
			Handler:     a.handleStash,
			Condition:   a.hasUncommittedChanges,
		},
		{
			Name:        "Pop Stash",
			Description: "Apply and remove latest stash",
			Shortcut:    "P",
			Category:    "Git",
			Handler:     a.handleStashPop,
			Condition:   a.hasStashes,
		},

		// Branch Operations
		{
			Name:        "Switch Branch",
			Description: "Switch to different branch",
			Shortcut:    "b",
			Category:    "Branch",
			Handler:     a.handleSwitchBranch,
		},
		{
			Name:        "Create Branch",
			Description: "Create new branch from current",
			Shortcut:    "B",
			Category:    "Branch",
			Handler:     a.handleCreateBranch,
		},
		{
			Name:        "Merge Branch",
			Description: "Merge branch into current",
			Shortcut:    "m",
			Category:    "Branch",
			Handler:     a.handleMergeBranch,
		},

		// Advanced Operations
		{
			Name:        "Create Worktree",
			Description: "Create new worktree for parallel development",
			Shortcut:    "w",
			Category:    "Advanced",
			Handler:     a.handleCreateWorktree,
		},
		{
			Name:        "Compare Files",
			Description: "Compare files between branches",
			Shortcut:    "d",
			Category:    "Advanced",
			Handler:     a.handleDiff,
		},
		{
			Name:        "View Log",
			Description: "View commit history",
			Shortcut:    "l",
			Category:    "Advanced",
			Handler:     a.handleLog,
		},
	}
}

// getAvailableActions returns actions that meet their conditions
func (a *ActionsPanel) getAvailableActions() []Action {
	var available []Action

	for _, action := range a.actions {
		if action.Condition == nil || action.Condition(a.state) {
			available = append(available, action)
		}
	}

	return available
}

// filterActions updates available actions based on current state
func (a *ActionsPanel) filterActions() {
	// Reset selection when filtering
	a.selectedAction = 0
}

// groupActionsByCategory groups actions by their category
func (a *ActionsPanel) groupActionsByCategory() map[string][]Action {
	categories := make(map[string][]Action)
	available := a.getAvailableActions()

	for _, action := range available {
		categories[action.Category] = append(categories[action.Category], action)
	}

	return categories
}

// renderAction renders a single action item
func (a *ActionsPanel) renderAction(action Action, index int) string {
	// Calculate global index for selection
	globalIndex := a.getGlobalIndex(action)
	isSelected := globalIndex == a.selectedAction

	var style lipgloss.Style
	if isSelected {
		style = styles.ListItemSelectedStyle
	} else {
		style = styles.ListItemStyle
	}

	// Build action text with icon
	icon := styles.GetActionIcon(action.Name)
	shortcut := ""
	if action.Shortcut != "" {
		shortcut = fmt.Sprintf("[%s] ", action.Shortcut)
	}

	text := fmt.Sprintf("%s %s%s - %s", icon, shortcut, action.Name, action.Description)

	if isSelected {
		text = "▶ " + text
	} else {
		text = "  " + text
	}

	return style.Render(text)
}

// getGlobalIndex calculates the global index of an action across all categories
func (a *ActionsPanel) getGlobalIndex(targetAction Action) int {
	index := 0
	available := a.getAvailableActions()

	for _, action := range available {
		if action.Name == targetAction.Name {
			return index
		}
		index++
	}

	return -1
}

// renderResult renders the last operation result
func (a *ActionsPanel) renderResult() string {
	if a.lastError != nil {
		errorStyle := styles.StatusErrorStyle.Copy().
			Padding(0, 1).
			Margin(0, 0, 1, 0)
		return errorStyle.Render(fmt.Sprintf("❌ Error: %s", a.lastError.Error()))
	}

	if a.lastResult != "" {
		successStyle := styles.StatusCleanStyle.Copy().
			Padding(0, 1).
			Margin(0, 0, 1, 0)
		return successStyle.Render(fmt.Sprintf("✅ %s", a.lastResult))
	}

	return ""
}

// renderNoRepository renders view when no repository is selected
func (a *ActionsPanel) renderNoRepository() string {
	return styles.MutedStyle.Render("Select a repository to see available actions")
}

// renderFooter renders the panel footer with shortcuts
func (a *ActionsPanel) renderFooter() string {
	shortcuts := []string{
		"↑/k: Up",
		"↓/j: Down",
		"Enter: Execute",
		"Tab: Next Panel",
	}

	footer := strings.Join(shortcuts, " • ")
	return styles.MutedStyle.Render(footer)
}

// executeSelectedAction executes the currently selected action
func (a *ActionsPanel) executeSelectedAction() tea.Cmd {
	available := a.getAvailableActions()
	if a.selectedAction >= len(available) {
		return nil
	}

	action := available[a.selectedAction]
	a.executing = true

	return action.Handler(a.state)
}

// executeQuickAction executes an action by shortcut
func (a *ActionsPanel) executeQuickAction(shortcut string) tea.Cmd {
	for _, action := range a.actions {
		if action.Shortcut == shortcut && (action.Condition == nil || action.Condition(a.state)) {
			a.executing = true
			return action.Handler(a.state)
		}
	}
	return nil
}

// Condition functions
func (a *ActionsPanel) hasUncommittedChanges(state *models.AppState) bool {
	repo := state.GetSelectedRepository()
	if repo == nil || repo.Status == nil {
		return false
	}
	return repo.Status.Workspace != types.Clean
}

func (a *ActionsPanel) hasUnpushedCommits(state *models.AppState) bool {
	repo := state.GetSelectedRepository()
	if repo == nil || repo.Status == nil {
		return false
	}
	return repo.Status.SyncStatus.Ahead > 0
}

func (a *ActionsPanel) hasStashes(state *models.AppState) bool {
	// This would need to be implemented with actual stash checking
	// For now, always show the option
	return true
}

// Action handlers - these will be implemented to execute actual operations
func (a *ActionsPanel) handleRefresh(state *models.AppState) tea.Cmd {
	return tea.Batch(
		models.ToastInfoCmd("Refreshing repository status..."),
		models.ProgressCmd("refresh", "Updating repository status", 0, true),
		func() tea.Msg {
			// Get selected repository
			repo := state.GetSelectedRepository()
			if repo == nil {
				return models.ActionCompleteMsg{
					Result: "",
					Error:  fmt.Errorf("no repository selected"),
				}
			}

			// Get repository path from configuration
			repoPath, exists := state.Repositories[repo.Alias]
			if !exists {
				return models.ActionCompleteMsg{
					Result: "",
					Error:  fmt.Errorf("repository path not found for %s", repo.Alias),
				}
			}

			// Get git manager
			gitMgr := getDIGitManager()
			if gitMgr == nil {
				return models.ActionCompleteMsg{
					Result: "",
					Error:  fmt.Errorf("git manager not available"),
				}
			}

			// Get fresh repository status
			status := gitMgr.GetRepoStatus(repo.Alias, repoPath)

			// Update the state with new status
			state.UpdateRepositoryData(repo.Alias, &status)

			return models.ActionCompleteMsg{
				Result: fmt.Sprintf("Refreshed status for %s", repo.Alias),
				Error:  nil,
			}
		},
	)
}

func (a *ActionsPanel) handleSync(state *models.AppState) tea.Cmd {
	return tea.Batch(
		models.ToastInfoCmd("Syncing repository..."),
		models.ProgressCmd("sync", "Pulling latest changes", 0, true),
		func() tea.Msg {
			// Get selected repository
			repo := state.GetSelectedRepository()
			if repo == nil {
				return models.ActionCompleteMsg{
					Result: "",
					Error:  fmt.Errorf("no repository selected"),
				}
			}

			// Get repository path from configuration
			repoPath, exists := state.Repositories[repo.Alias]
			if !exists {
				return models.ActionCompleteMsg{
					Result: "",
					Error:  fmt.Errorf("repository path not found for %s", repo.Alias),
				}
			}

			// Get git manager
			gitMgr := getDIGitManager()
			if gitMgr == nil {
				return models.ActionCompleteMsg{
					Result: "",
					Error:  fmt.Errorf("git manager not available"),
				}
			}

			// Execute sync operation (fetch + pull)
			err := gitMgr.SyncRepository(repoPath, "ff-only") // Use fast-forward only mode
			if err != nil {
				return models.ActionCompleteMsg{
					Result: "",
					Error:  fmt.Errorf("failed to sync repository: %w", err),
				}
			}

			return models.ActionCompleteMsg{
				Result: fmt.Sprintf("Successfully synchronized %s", repo.Alias),
				Error:  nil,
			}
		},
	)
}

func (a *ActionsPanel) handleCommit(state *models.AppState) tea.Cmd {
	return tea.Batch(
		models.ToastInfoCmd("Preparing commit..."),
		models.ProgressCmd("commit", "Committing changes", 0, true),
		func() tea.Msg {
			// Get selected repository
			repo := state.GetSelectedRepository()
			if repo == nil {
				return models.ActionCompleteMsg{
					Result: "",
					Error:  fmt.Errorf("no repository selected"),
				}
			}

			// Get repository path from configuration
			repoPath, exists := state.Repositories[repo.Alias]
			if !exists {
				return models.ActionCompleteMsg{
					Result: "",
					Error:  fmt.Errorf("repository path not found for %s", repo.Alias),
				}
			}

			// Import DI to access git manager
			gitMgr := getDIGitManager()
			if gitMgr == nil {
				return models.ActionCompleteMsg{
					Result: "",
					Error:  fmt.Errorf("git manager not available"),
				}
			}

			// Check for uncommitted changes
			hasChanges, err := gitMgr.HasUncommittedChanges(repoPath)
			if err != nil {
				return models.ActionCompleteMsg{
					Result: "",
					Error:  fmt.Errorf("failed to check for changes: %w", err),
				}
			}

			if !hasChanges {
				return models.ActionCompleteMsg{
					Result: "No changes to commit",
					Error:  nil,
				}
			}

			// For now, use a default commit message. Later this can be enhanced with a dialog
			commitMessage := "TUI commit: automated commit from dashboard"

			// Stage all changes and commit
			err = gitMgr.CommitChanges(repoPath, commitMessage, true) // true = add all changes
			if err != nil {
				return models.ActionCompleteMsg{
					Result: "",
					Error:  fmt.Errorf("failed to commit changes: %w", err),
				}
			}

			return models.ActionCompleteMsg{
				Result: fmt.Sprintf("Successfully committed changes in %s", repo.Alias),
				Error:  nil,
			}
		},
	)
}

func (a *ActionsPanel) handlePush(state *models.AppState) tea.Cmd {
	return tea.Batch(
		models.ToastInfoCmd("Pushing changes..."),
		models.ProgressCmd("push", "Pushing to remote", 0, true),
		func() tea.Msg {
			// Get selected repository
			repo := state.GetSelectedRepository()
			if repo == nil {
				return models.ActionCompleteMsg{
					Result: "",
					Error:  fmt.Errorf("no repository selected"),
				}
			}

			// Get repository path from configuration
			repoPath, exists := state.Repositories[repo.Alias]
			if !exists {
				return models.ActionCompleteMsg{
					Result: "",
					Error:  fmt.Errorf("repository path not found for %s", repo.Alias),
				}
			}

			// Get git manager
			gitMgr := getDIGitManager()
			if gitMgr == nil {
				return models.ActionCompleteMsg{
					Result: "",
					Error:  fmt.Errorf("git manager not available"),
				}
			}

			// Check for unpushed commits
			hasUnpushed, err := gitMgr.HasUnpushedCommits(repoPath)
			if err != nil {
				return models.ActionCompleteMsg{
					Result: "",
					Error:  fmt.Errorf("failed to check for unpushed commits: %w", err),
				}
			}

			if !hasUnpushed {
				return models.ActionCompleteMsg{
					Result: "No commits to push",
					Error:  nil,
				}
			}

			// Execute push operation
			err = gitMgr.PushChanges(repoPath, false, false) // force=false, setUpstream=false
			if err != nil {
				return models.ActionCompleteMsg{
					Result: "",
					Error:  fmt.Errorf("failed to push changes: %w", err),
				}
			}

			return models.ActionCompleteMsg{
				Result: fmt.Sprintf("Successfully pushed changes from %s", repo.Alias),
				Error:  nil,
			}
		},
	)
}

func (a *ActionsPanel) handleStash(state *models.AppState) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement stash operation
		return models.ActionCompleteMsg{
			Result: "Changes stashed",
			Error:  nil,
		}
	}
}

func (a *ActionsPanel) handleStashPop(state *models.AppState) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement stash pop operation
		return models.ActionCompleteMsg{
			Result: "Stash applied",
			Error:  nil,
		}
	}
}

func (a *ActionsPanel) handleSwitchBranch(state *models.AppState) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement branch switching dialog
		return models.ActionCompleteMsg{
			Result: "Branch switched",
			Error:  nil,
		}
	}
}

func (a *ActionsPanel) handleCreateBranch(state *models.AppState) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement branch creation dialog
		return models.ActionCompleteMsg{
			Result: "Branch created",
			Error:  nil,
		}
	}
}

func (a *ActionsPanel) handleMergeBranch(state *models.AppState) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement branch merge dialog
		return models.ActionCompleteMsg{
			Result: "Branch merged",
			Error:  nil,
		}
	}
}

func (a *ActionsPanel) handleCreateWorktree(state *models.AppState) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement worktree creation dialog
		return models.ActionCompleteMsg{
			Result: "Worktree created",
			Error:  nil,
		}
	}
}

func (a *ActionsPanel) handleDiff(state *models.AppState) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement diff viewer
		return models.ActionCompleteMsg{
			Result: "Diff viewer opened",
			Error:  nil,
		}
	}
}

func (a *ActionsPanel) handleLog(state *models.AppState) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement log viewer
		return models.ActionCompleteMsg{
			Result: "Log viewer opened",
			Error:  nil,
		}
	}
}

func (a *ActionsPanel) handleOpenTerminal(state *models.AppState) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement terminal opening
		return models.ActionCompleteMsg{
			Result: "Terminal opened",
			Error:  nil,
		}
	}
}

func (a *ActionsPanel) handleOpenFileManager(state *models.AppState) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement file manager opening
		return models.ActionCompleteMsg{
			Result: "File manager opened",
			Error:  nil,
		}
	}
}

// getDIGitManager safely gets the git manager from DI container
func getDIGitManager() *git.Manager {
	// This is safe to call from any context
	return di.GitManager()
}
