package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gman/internal/config"
	"gman/internal/fzf"
	"gman/internal/index"
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

	case models.FzfLaunchMsg:
		// Handle fzf search launch
		return a, a.launchFzfSearch(msg.Mode, msg.Query)

	case models.SearchResultsMsg:
		// Handle search results returned from fzf
		a.state.SearchState.Results = msg.Results
		a.state.SearchState.Query = msg.Query
		if len(msg.Results) > 0 {
			a.state.SearchState.SelectedItem = 0
			// Update preview with first result
			cmds = append(cmds, a.updatePreviewFromSearch(msg.Results[0]))
		}
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

// launchFzfSearch launches an external fzf search process
func (a *App) launchFzfSearch(mode models.SearchMode, query string) tea.Cmd {
	return func() tea.Msg {
		// Create searcher to get data
		searcher, err := index.NewSearcher(a.state.ConfigManager)
		if err != nil {
			return models.ErrorMsg{Error: fmt.Errorf("failed to create searcher: %w", err)}
		}
		defer searcher.Close()

		// Get search results
		repos := a.state.Repositories
		var searchResults []index.SearchResult
		
		switch mode {
		case models.SearchFiles:
			searchResults, err = searcher.SearchFiles("", "", repos)
		case models.SearchCommits:
			searchResults, err = searcher.SearchCommits("", "", repos)
		}
		
		if err != nil {
			return models.ErrorMsg{Error: fmt.Errorf("search failed: %w", err)}
		}

		// Convert to TUI search results
		results := make([]models.SearchResultItem, len(searchResults))
		for i, result := range searchResults {
			results[i] = models.SearchResultItem{
				Type:        result.Type,
				Repository:  result.RepoAlias,
				Path:        result.Path,
				Hash:        result.Hash,
				DisplayText: result.DisplayText,
				PreviewData: result.Data,
			}
		}
		
		return models.SearchResultsMsg{
			Mode:    mode,
			Query:   query,
			Results: results,
			Error:   nil,
		}
	}
}

// buildFzfCommand constructs the fzf command for the given search mode
func (a *App) buildFzfCommand(mode models.SearchMode, query string) *exec.Cmd {
	// Check if fzf is available
	if !fzf.IsAvailable() {
		return exec.Command("echo", "fzf not available")
	}

	// Create searcher to get data
	searcher, err := index.NewSearcher(a.state.ConfigManager)
	if err != nil {
		return exec.Command("echo", fmt.Sprintf("Error creating searcher: %v", err))
	}
	defer searcher.Close()
	
	var cmd *exec.Cmd
	
	switch mode {
	case models.SearchFiles:
		// Get all files from indexed repositories
		repos := a.state.Repositories
		results, err := searcher.SearchFiles("", "", repos) // Empty query to get all files
		if err != nil {
			cmd = exec.Command("echo", fmt.Sprintf("Error: %v", err))
		} else {
			// Create input for fzf
			input := strings.Join(searcher.FormatFileSearchResults(results), "\n")
			
			// Build fzf command with file-specific options
			opts := fzf.DefaultFileOptions()
			opts.Prompt = "Files> "
			opts.Header = "Search files across repositories"
			opts.InitialQuery = query
			
			cmd = exec.Command("fzf", a.buildFzfArgs(opts)...)
			cmd.Stdin = strings.NewReader(input)
		}
		
	case models.SearchCommits:
		// Get all commits from indexed repositories
		repos := a.state.Repositories
		results, err := searcher.SearchCommits("", "", repos) // Empty query to get all commits
		if err != nil {
			cmd = exec.Command("echo", fmt.Sprintf("Error: %v", err))
		} else {
			// Create input for fzf
			input := strings.Join(searcher.FormatCommitSearchResults(results), "\n")
			
			// Build fzf command with commit-specific options
			opts := fzf.DefaultCommitOptions()
			opts.Prompt = "Commits> "
			opts.Header = "Search commits across repositories"
			opts.InitialQuery = query
			
			cmd = exec.Command("fzf", a.buildFzfArgs(opts)...)
			cmd.Stdin = strings.NewReader(input)
		}
	}
	
	if cmd == nil {
		cmd = exec.Command("echo", "No search results")
	}
	
	return cmd
}

// buildFzfArgs converts fzf options to command line arguments
func (a *App) buildFzfArgs(opts fzf.Options) []string {
	args := []string{}
	
	if opts.Prompt != "" {
		args = append(args, "--prompt", opts.Prompt)
	}
	if opts.Header != "" {
		args = append(args, "--header", opts.Header)
	}
	if opts.Height != "" {
		args = append(args, "--height", opts.Height)
	}
	if opts.Layout != "" {
		args = append(args, "--layout", opts.Layout)
	}
	if opts.Border {
		args = append(args, "--border")
	}
	if opts.InitialQuery != "" {
		args = append(args, "--query", opts.InitialQuery)
	}
	
	return args
}


// parseFzfResults parses the output from fzf and creates search results
func (a *App) parseFzfResults(mode models.SearchMode) tea.Msg {
	// Read from stdout file or use other mechanism to get fzf output
	// For now, return empty results
	// TODO: Implement proper result parsing
	
	results := []models.SearchResultItem{}
	
	return models.SearchResultsMsg{
		Mode:    mode,
		Query:   "", // TODO: Extract query from fzf
		Results: results,
		Error:   nil,
	}
}

// updatePreviewFromSearch updates the preview panel with search result content
func (a *App) updatePreviewFromSearch(result models.SearchResultItem) tea.Cmd {
	return func() tea.Msg {
		var content string
		var contentType models.PreviewType
		
		switch result.Type {
		case "file":
			// Read file content for preview
			filePath := result.Path
			if data, err := os.ReadFile(filePath); err == nil {
				content = string(data)
			} else {
				content = fmt.Sprintf("Error reading file: %v", err)
			}
			contentType = models.PreviewFile
			
		case "commit":
			// Get commit diff for preview
			content = fmt.Sprintf("Commit: %s\nRepository: %s\n\n%s", 
				result.Hash, result.Repository, result.DisplayText)
			contentType = models.PreviewCommit
		}
		
		return models.PreviewContentMsg{
			Content:     content,
			ContentType: contentType,
			FilePath:    result.Path,
			CommitHash:  result.Hash,
		}
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