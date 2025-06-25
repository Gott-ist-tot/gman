package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gman/internal/config"
	"gman/internal/tui/models"
	"gman/internal/tui/panels"
	"gman/internal/tui/styles"
)

// App represents the main TUI application
type App struct {
	state        *models.AppState
	repositoryPanel *panels.RepositoryPanel
	statusPanel     *panels.StatusPanel
	searchPanel     *panels.SearchPanel
	previewPanel    *panels.PreviewPanel

	// UI state
	ready  bool
	quitting bool
}

// NewApp creates a new TUI application
func NewApp(configMgr *config.Manager) *App {
	state := models.NewAppState(configMgr)
	
	return &App{
		state:           state,
		repositoryPanel: panels.NewRepositoryPanel(state),
		statusPanel:     panels.NewStatusPanel(state),
		searchPanel:     panels.NewSearchPanel(state),
		previewPanel:    panels.NewPreviewPanel(state),
		ready:           false,
		quitting:        false,
	}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		models.StatusTickCmd(),
		a.repositoryPanel.Init(),
		a.statusPanel.Init(),
		a.searchPanel.Init(),
		a.previewPanel.Init(),
	)
}

// Update handles messages and updates the application state
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.state.UpdateWindowSize(msg.Width, msg.Height)
		a.ready = true
		return a, nil

	case models.WindowSizeMsg:
		a.state.UpdateWindowSize(msg.Width, msg.Height)
		return a, nil

	case tea.KeyMsg:
		cmd := a.handleKeyMsg(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case models.ExitMsg:
		a.quitting = true
		return a, tea.Quit

	case models.StatusTickMsg:
		// Refresh repository status periodically
		cmds = append(cmds, models.StatusTickCmd())
		if a.state.StatusState.AutoRefresh {
			cmds = append(cmds, models.RefreshCmd(false))
		}

	case models.RepositorySelectedMsg:
		a.state.SelectedRepo = msg.Alias
		// Trigger status update for selected repository
		_, cmd := a.statusPanel.Update(msg)
		cmds = append(cmds, cmd)

	case models.PanelFocusMsg:
		a.state.SetFocusedPanel(msg.Panel)

	case models.HelpToggleMsg:
		a.state.ToggleHelp()

	case models.ErrorMsg:
		// Handle errors (could show in status bar or notification)
		// For now, just log them
		fmt.Printf("Error: %v\n", msg.Error)
	}

	// Update panels
	var cmd tea.Cmd

	// Update focused panel first
	switch a.state.FocusedPanel {
	case models.RepositoryPanel:
		_, cmd = a.repositoryPanel.Update(msg)
		cmds = append(cmds, cmd)
	case models.StatusPanel:
		_, cmd = a.statusPanel.Update(msg)
		cmds = append(cmds, cmd)
	case models.SearchPanel:
		_, cmd = a.searchPanel.Update(msg)
		cmds = append(cmds, cmd)
	case models.PreviewPanel:
		_, cmd = a.previewPanel.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update other panels with non-input messages
	if !isInputMessage(msg) {
		_, cmd = a.repositoryPanel.Update(msg)
		cmds = append(cmds, cmd)
		
		_, cmd = a.statusPanel.Update(msg)
		cmds = append(cmds, cmd)
		
		_, cmd = a.searchPanel.Update(msg)
		cmds = append(cmds, cmd)
		
		_, cmd = a.previewPanel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}

// View renders the application
func (a *App) View() string {
	if !a.ready {
		return "Initializing TUI Dashboard..."
	}

	if a.quitting {
		return "Goodbye!\n"
	}

	if a.state.ShowHelp {
		return a.renderHelp()
	}

	return a.renderDashboard()
}

// handleKeyMsg handles keyboard input
func (a *App) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c", "q":
		return models.ExitCmd()

	case "?", "h":
		return models.HelpToggleCmd()

	case "tab":
		a.state.NextPanel()
		return models.PanelFocusCmd(a.state.FocusedPanel)

	case "shift+tab":
		a.state.PrevPanel()
		return models.PanelFocusCmd(a.state.FocusedPanel)

	case "1":
		a.state.SetFocusedPanel(models.RepositoryPanel)
		return models.PanelFocusCmd(models.RepositoryPanel)

	case "2":
		a.state.SetFocusedPanel(models.StatusPanel)
		return models.PanelFocusCmd(models.StatusPanel)

	case "3":
		a.state.SetFocusedPanel(models.SearchPanel)
		return models.PanelFocusCmd(models.SearchPanel)

	case "4":
		a.state.SetFocusedPanel(models.PreviewPanel)
		return models.PanelFocusCmd(models.PreviewPanel)

	case "r":
		// Refresh all repository status
		return models.RefreshCmd(true)

	case "f5":
		// Force refresh
		return models.RefreshCmd(true)
	}

	return nil
}

// renderDashboard renders the main dashboard view
func (a *App) renderDashboard() string {
	width := a.state.WindowWidth
	height := a.state.WindowHeight

	// Calculate panel dimensions
	leftWidth := width / 2 - 2
	rightWidth := width - leftWidth - 3
	topHeight := height / 2 - 2
	bottomHeight := height - topHeight - 4

	// Render panels
	repoPanel := a.renderPanel(
		a.repositoryPanel.View(),
		"Repositories (1)",
		leftWidth,
		topHeight,
		a.state.FocusedPanel == models.RepositoryPanel,
	)

	statusPanel := a.renderPanel(
		a.statusPanel.View(),
		"Status (2)",
		rightWidth,
		topHeight,
		a.state.FocusedPanel == models.StatusPanel,
	)

	searchPanel := a.renderPanel(
		a.searchPanel.View(),
		"Search (3)",
		leftWidth,
		bottomHeight,
		a.state.FocusedPanel == models.SearchPanel,
	)

	previewPanel := a.renderPanel(
		a.previewPanel.View(),
		"Preview (4)",
		rightWidth,
		bottomHeight,
		a.state.FocusedPanel == models.PreviewPanel,
	)

	// Arrange panels in 2x2 grid
	topRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		repoPanel,
		" ",
		statusPanel,
	)

	bottomRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		searchPanel,
		" ",
		previewPanel,
	)

	dashboard := lipgloss.JoinVertical(
		lipgloss.Left,
		topRow,
		" ",
		bottomRow,
	)

	// Add status bar
	statusBar := a.renderStatusBar()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		dashboard,
		statusBar,
	)
}

// renderPanel renders a panel with border and title
func (a *App) renderPanel(content, title string, width, height int, focused bool) string {
	return styles.CreatePanel(content, title, width, height, focused)
}

// renderStatusBar renders the bottom status bar
func (a *App) renderStatusBar() string {
	helpText := "? Help • Tab Next Panel • q Quit • r Refresh"
	
	statusStyle := styles.BodyStyle.
		Background(styles.ColorBgSecondary).
		Foreground(styles.ColorTextSecondary).
		Padding(0, 1).
		Width(a.state.WindowWidth - 2)

	return statusStyle.Render(helpText)
}

// renderHelp renders the help screen
func (a *App) renderHelp() string {
	help := `gman TUI Dashboard - Keyboard Shortcuts

GLOBAL NAVIGATION:
  Tab / Shift+Tab     Navigate between panels
  1 / 2 / 3 / 4       Jump to specific panel
  ? / h               Toggle this help
  q / Ctrl+C          Quit application

REPOSITORY PANEL (1):
  ↑ / k               Move up
  ↓ / j               Move down
  Enter               Select repository
  / 		      Filter repositories
  s                   Change sort order
  g                   Filter by group

STATUS PANEL (2):
  e                   Toggle extended view
  b                   Toggle branch information
  r                   Refresh status

SEARCH PANEL (3):
  Enter               Start search
  Tab                 Toggle search mode (files/commits)
  /                   Search files
  c                   Search commits
  Esc                 Cancel search

PREVIEW PANEL (4):
  ↑ / k               Scroll up
  ↓ / j               Scroll down
  Page Up/Down        Page scroll
  Home / End          Top / Bottom

GLOBAL ACTIONS:
  r                   Refresh all repositories
  F5                  Force refresh
  Ctrl+R              Reload configuration

Press any key to return to dashboard...`

	helpStyle := styles.HelpStyle.
		Width(a.state.WindowWidth - 4).
		Height(a.state.WindowHeight - 4).
		Align(lipgloss.Left)

	return helpStyle.Render(help)
}

// isInputMessage returns true if the message is an input message that should
// only be handled by the focused panel
func isInputMessage(msg tea.Msg) bool {
	switch msg.(type) {
	case tea.KeyMsg:
		return true
	default:
		return false
	}
}

// Run starts the TUI application
func Run(configMgr *config.Manager) error {
	app := NewApp(configMgr)
	
	// Create the tea program
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run the program
	_, err := p.Run()
	return err
}