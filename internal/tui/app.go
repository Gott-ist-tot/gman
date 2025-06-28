package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gman/internal/config"
	"gman/internal/external"
	"gman/internal/fzf"
	"gman/internal/tui/models"
	"gman/internal/tui/panels"
	"gman/internal/tui/styles"
)

// App represents the main TUI application
type App struct {
	state           *models.AppState
	repositoryPanel *panels.RepositoryPanel
	statusPanel     *panels.StatusPanel
	searchPanel     *panels.SearchPanel
	previewPanel    *panels.PreviewPanel
	actionsPanel    *panels.ActionsPanel

	// UI state
	ready    bool
	quitting bool
	
	// Tea program for sending async messages
	teaProgram *tea.Program
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
		actionsPanel:    panels.NewActionsPanel(state),
		ready:           false,
		quitting:        false,
	}
}

// SetTeaProgram sets the tea program reference for async messaging
func (a *App) SetTeaProgram(p *tea.Program) {
	a.teaProgram = p
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		models.StatusTickCmd(),
		a.repositoryPanel.Init(),
		a.statusPanel.Init(),
		a.searchPanel.Init(),
		a.previewPanel.Init(),
		a.actionsPanel.Init(),
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
		if a.state.GetAutoRefresh() {
			cmds = append(cmds, models.RefreshCmd(false))
		}

	case models.RepositorySelectedMsg:
		a.state.SetSelectedRepo(msg.Alias)
		// Trigger status update for selected repository
		_, cmd := a.statusPanel.Update(msg)
		cmds = append(cmds, cmd)

	case models.PanelFocusMsg:
		a.state.SetFocusedPanel(msg.Panel)

	case models.HelpToggleMsg:
		a.state.ToggleHelp()

	case models.ErrorMsg:
		// Handle errors with improved context and state management
		a.state.SetError(msg.Error)
		
		// Log error with context for debugging
		if msg.Context != "" {
			fmt.Printf("Error in %s: %v\n", msg.Context, msg.Error)
		} else {
			fmt.Printf("Error: %v\n", msg.Error)
		}
		
		// If this is a fatal error, initiate app exit
		if msg.Fatal {
			cmds = append(cmds, models.ExitCmd())
		}

	case models.FzfLaunchMsg:
		// Handle fzf search launch
		return a, a.launchFzfSearch(msg.Mode, msg.Query)

	case models.SearchResultsMsg:
		// Handle search results returned from fzf
		if msg.Error != nil {
			a.state.CompleteSearch(nil, msg.Error)
		} else {
			a.state.CompleteSearch(msg.Results, nil)
			if len(msg.Results) > 0 {
				// Update preview with first result
				cmds = append(cmds, a.updatePreviewFromSearch(msg.Results[0]))
			}
		}

	case models.SearchStartedMsg:
		// Search has started - the state was already updated in launchFzfSearch
		// No additional action needed here

	case models.SearchProgressMsg:
		// Update search progress
		a.state.UpdateSearchProgress(msg.Progress, msg.CurrentOp, msg.Partial)

	case models.SearchCancelledMsg:
		// Search was cancelled
		a.state.CancelSearch()

	case models.ToastMsg:
		// Add toast notification
		id := a.state.AddToast(msg.Message, msg.Type, msg.Duration)
		
		// Set up auto-hide timer
		cmds = append(cmds, tea.Tick(msg.Duration, func(t time.Time) tea.Msg {
			return models.ToastHideMsg{ID: id}
		}))

	case models.ToastHideMsg:
		// Remove toast notification
		a.state.RemoveToast(msg.ID)

	case models.ProgressMsg:
		// Update progress indicator
		a.state.SetProgress(msg.ID, msg.Message, msg.Progress, msg.Indeterminate)

	case models.ProgressHideMsg:
		// Remove progress indicator
		a.state.RemoveProgress(msg.ID)
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
	case models.ActionsPanel:
		_, cmd = a.actionsPanel.Update(msg)
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

		_, cmd = a.actionsPanel.Update(msg)
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

	case "5":
		a.state.SetFocusedPanel(models.ActionsPanel)
		return models.PanelFocusCmd(models.ActionsPanel)

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

	// Ensure minimum dimensions
	if width < 80 {
		width = 80
	}
	if height < 20 {
		height = 20
	}

	// Calculate panel dimensions for 2x3 layout with proper proportions
	// Top row: Repository (25%) | Status (35%) | Actions (40%)
	spacing := 2 // space between panels
	totalSpacing := spacing * 2 // 2 spaces between 3 panels
	
	repoWidth := (width * 25) / 100
	statusWidth := (width * 35) / 100
	actionsWidth := width - repoWidth - statusWidth - totalSpacing
	
	// Bottom row: Search (60%) | Preview (40%)
	searchWidth := (width * 60) / 100
	previewWidth := width - searchWidth - spacing
	
	// Heights
	topHeight := (height * 55) / 100 // Give more space to top row
	bottomHeight := height - topHeight - 3 // Account for status bar
	
	// Debug: Add layout information to help diagnosis (can be removed later)
	if width < 120 || height < 30 {
		// For smaller terminals, adjust proportions
		repoWidth = (width * 30) / 100
		statusWidth = (width * 40) / 100
		actionsWidth = width - repoWidth - statusWidth - totalSpacing
	}

	// Render panels with corrected dimensions
	repoPanel := a.renderPanel(
		a.repositoryPanel.View(),
		"Repositories (1)",
		repoWidth,
		topHeight,
		a.state.FocusedPanel == models.RepositoryPanel,
	)

	statusPanel := a.renderPanel(
		a.statusPanel.View(),
		"Status (2)",
		statusWidth,
		topHeight,
		a.state.FocusedPanel == models.StatusPanel,
	)

	actionsPanel := a.renderPanel(
		a.actionsPanel.View(),
		"Actions (5)",
		actionsWidth,
		topHeight,
		a.state.FocusedPanel == models.ActionsPanel,
	)

	searchPanel := a.renderPanel(
		a.searchPanel.View(),
		"Search (3)",
		searchWidth,
		bottomHeight,
		a.state.FocusedPanel == models.SearchPanel,
	)

	previewPanel := a.renderPanel(
		a.previewPanel.View(),
		"Preview (4)",
		previewWidth,
		bottomHeight,
		a.state.FocusedPanel == models.PreviewPanel,
	)

	// Arrange panels in 2x3 grid with proper spacing
	spacer := strings.Repeat(" ", spacing)
	
	topRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		repoPanel,
		spacer,
		statusPanel,
		spacer,
		actionsPanel,
	)

	bottomRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		searchPanel,
		spacer,
		previewPanel,
	)

	dashboard := lipgloss.JoinVertical(
		lipgloss.Left,
		topRow,
		" ",
		bottomRow,
	)

	// Add status bar with layout debug info
	statusBar := a.renderStatusBar()
	
	// Debug layout info (can be removed later)
	debugInfo := fmt.Sprintf("Layout: W=%d H=%d | Repo=%d Status=%d Actions=%d | Search=%d Preview=%d", 
		width, height, repoWidth, statusWidth, actionsWidth, searchWidth, previewWidth)
	debugBar := styles.MutedStyle.Render(debugInfo)

	// Add toasts overlay
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		dashboard,
		debugBar,
		statusBar,
	)

	// Render toasts on top-right corner
	content = a.renderWithToasts(content)

	// Render progress indicators on bottom
	content = a.renderWithProgress(content)

	return content
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
  1 / 2 / 3 / 4 / 5   Jump to specific panel
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

ACTIONS PANEL (5):
  ↑ / k               Move up action list
  ↓ / j               Move down action list
  Enter               Execute selected action
  r                   Quick refresh
  s                   Quick sync
  c                   Quick commit
  p                   Quick push
  S                   Quick stash

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

// renderWithToasts renders content with toast notifications overlaid
func (a *App) renderWithToasts(content string) string {
	toasts := a.state.GetActiveToasts()
	if len(toasts) == 0 {
		return content
	}

	// Calculate position for toasts (top-right corner)
	windowWidth := a.state.WindowWidth
	toastWidth := 40
	toastX := windowWidth - toastWidth - 2

	// Render each toast
	var toastViews []string
	for _, toast := range toasts {
		toastView := styles.RenderToast(toast.Message, int(toast.Type))
		toastViews = append(toastViews, toastView)
	}

	// Join toasts vertically
	toastStack := lipgloss.JoinVertical(lipgloss.Left, toastViews...)

	// Position toasts at top-right
	toastOverlay := lipgloss.NewStyle().
		MarginTop(1).
		MarginLeft(toastX).
		Render(toastStack)

	return lipgloss.JoinVertical(lipgloss.Left, toastOverlay, content)
}

// renderWithProgress renders content with progress indicators
func (a *App) renderWithProgress(content string) string {
	activeProgress := a.state.GetActiveProgress()
	if len(activeProgress) == 0 {
		return content
	}

	// Render progress indicators
	var progressViews []string
	for _, progress := range activeProgress {
		var progressView string
		if progress.Indeterminate {
			// Use spinner for indeterminate progress
			frame := int(time.Since(progress.StartTime).Milliseconds() / 100)
			progressView = styles.RenderSpinner(progress.Message, frame)
		} else {
			// Use progress bar for determinate progress
			progressView = styles.RenderProgressBar(progress.Progress, progress.Message, 30)
		}
		progressViews = append(progressViews, progressView)
	}

	// Join progress indicators
	progressStack := lipgloss.JoinVertical(lipgloss.Left, progressViews...)

	// Add progress at bottom
	return lipgloss.JoinVertical(lipgloss.Left, content, progressStack)
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

// launchFzfSearch launches an asynchronous search operation
func (a *App) launchFzfSearch(mode models.SearchMode, query string) tea.Cmd {
	return func() tea.Msg {
		// Create cancellable context
		ctx, cancel := context.WithCancel(context.Background())
		
		// Start the search in the app state
		a.state.StartSearch(mode, query, cancel)
		
		// Launch the actual search asynchronously
		go a.performAsyncSearch(ctx, mode, query)
		
		// Return immediately to not block the UI
		return models.SearchStartedMsg{
			Mode:  mode,
			Query: query,
		}
	}
}

// performAsyncSearch performs the search operation asynchronously
func (a *App) performAsyncSearch(ctx context.Context, mode models.SearchMode, query string) {
	// Get repositories from state safely
	repos := a.state.Repositories
	
	// Send progress update
	a.sendSearchProgress(mode, query, 10, "Starting search...", nil)
	
	// Check for cancellation
	if ctx.Err() != nil {
		a.sendSearchCancelled(mode, query)
		return
	}

	switch mode {
	case models.SearchFiles:
		// Use new SmartSearcher for file search
		a.performFileSearch(ctx, query, repos)
	case models.SearchCommits:
		// Keep using index searcher for commits (different infrastructure)
		a.performCommitSearch(ctx, query, repos)
	}
}

// performFileSearch performs file search using the new SmartSearcher
func (a *App) performFileSearch(ctx context.Context, query string, repos map[string]string) {
	a.sendSearchProgress(models.SearchFiles, query, 30, "Searching files with SmartSearcher...", nil)
	
	// Create SmartSearcher with verbose=false for TUI
	smartSearcher := external.NewSmartSearcher(false)
	
	// Get current group filter from app state if available
	groupFilter := "" // TODO: Implement group filter in TUI state
	
	// Perform the search
	fileResults, err := smartSearcher.SearchFiles(query, repos, groupFilter)
	if err != nil {
		a.sendSearchError(models.SearchFiles, query, fmt.Errorf("file search failed: %w", err))
		return
	}
	
	// Check for cancellation after search
	if ctx.Err() != nil {
		a.sendSearchCancelled(models.SearchFiles, query)
		return
	}
	
	// Convert file results to SearchResultItem
	a.sendSearchProgress(models.SearchFiles, query, 90, "Processing file results...", nil)
	
	results := make([]models.SearchResultItem, len(fileResults))
	for i, result := range fileResults {
		// Check for cancellation during conversion
		if ctx.Err() != nil {
			a.sendSearchCancelled(models.SearchFiles, query)
			return
		}
		
		results[i] = models.SearchResultItem{
			Type:        "file",
			Repository:  result.RepoAlias,
			Path:        result.FullPath,
			Hash:        "", // Not applicable for files
			DisplayText: result.DisplayText,
			PreviewData: result.RelativePath, // Use relative path for preview
		}
	}
	
	// Send final results
	a.sendSearchComplete(models.SearchFiles, query, results)
}

// performCommitSearch performs commit search using real-time git log approach
func (a *App) performCommitSearch(ctx context.Context, query string, repos map[string]string) {
	a.sendSearchProgress(models.SearchCommits, query, 20, "Initializing commit search...", nil)
	
	// Check for cancellation
	if ctx.Err() != nil {
		a.sendSearchCancelled(models.SearchCommits, query)
		return
	}

	// Collect commits from all repositories using git log (same approach as CLI)
	var results []models.SearchResultItem
	repoCount := 0
	totalRepos := len(repos)

	for alias, path := range repos {
		// Update progress
		progress := 20 + (repoCount*60)/totalRepos
		a.sendSearchProgress(models.SearchCommits, query, progress, fmt.Sprintf("Searching commits in %s...", alias), nil)
		
		// Check for cancellation between repositories
		if ctx.Err() != nil {
			a.sendSearchCancelled(models.SearchCommits, query)
			return
		}

		// Build git log command - same as CLI implementation
		gitArgs := []string{"log", "--oneline", "--all", "--decorate", "--color=always", "-n", "100"}
		if query != "" {
			gitArgs = append(gitArgs, fmt.Sprintf("--grep=%s", query))
		}

		// Execute git log
		gitCmd := exec.Command("git", gitArgs...)
		gitCmd.Dir = path
		output, err := gitCmd.Output()
		if err != nil {
			// Skip repositories that don't have commits or have errors (same as CLI)
			repoCount++
			continue
		}

		if len(output) > 0 {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, line := range lines {
				if line != "" {
					// Parse commit hash from git log output
					parts := strings.Fields(line)
					var hash string
					if len(parts) > 0 {
						// Remove ANSI color codes to get clean hash
						hash = strings.Fields(strings.ReplaceAll(parts[0], "\x1b[m", ""))[0]
					}

					// Create search result item
					result := models.SearchResultItem{
						Type:        "commit",
						Repository:  alias,
						Path:        path,
						Hash:        hash,
						DisplayText: fmt.Sprintf("[%s] %s", alias, line),
						PreviewData: map[string]string{
							"hash":       hash,
							"repository": alias,
							"path":       path,
						},
					}
					results = append(results, result)
				}
			}
		}
		repoCount++
	}

	// Final progress update
	a.sendSearchProgress(models.SearchCommits, query, 90, "Processing commit results...", nil)
	
	// Check for cancellation after search
	if ctx.Err() != nil {
		a.sendSearchCancelled(models.SearchCommits, query)
		return
	}

	// Send final results
	a.sendSearchComplete(models.SearchCommits, query, results)
}

// searchWithProgress wraps search operations with progress updates - temporarily disabled
func (a *App) searchWithProgress(ctx context.Context, searchFunc func() ([]models.SearchResultItem, error), mode models.SearchMode, query string) ([]models.SearchResultItem, error) {
	// TODO: Update during TUI refactoring to use new git log approach
	return nil, fmt.Errorf("search functionality temporarily disabled during refactoring")
}

// Helper functions to send search messages asynchronously
func (a *App) sendSearchProgress(mode models.SearchMode, query string, progress int, currentOp string, partial []models.SearchResultItem) {
	// Send message via tea program (this is safe from goroutines)
	go func() {
		if a.teaProgram != nil {
			a.teaProgram.Send(models.SearchProgressMsg{
				Mode:      mode,
				Query:     query,
				Progress:  progress,
				CurrentOp: currentOp,
				Partial:   partial,
			})
		}
	}()
}

func (a *App) sendSearchComplete(mode models.SearchMode, query string, results []models.SearchResultItem) {
	go func() {
		if a.teaProgram != nil {
			a.teaProgram.Send(models.SearchResultsMsg{
				Mode:    mode,
				Query:   query,
				Results: results,
			})
		}
	}()
}

func (a *App) sendSearchError(mode models.SearchMode, query string, err error) {
	go func() {
		if a.teaProgram != nil {
			a.teaProgram.Send(models.SearchResultsMsg{
				Mode:    mode,
				Query:   query,
				Results: nil,
				Error:   err,
			})
		}
	}()
}

func (a *App) sendSearchCancelled(mode models.SearchMode, query string) {
	go func() {
		if a.teaProgram != nil {
			a.teaProgram.Send(models.SearchCancelledMsg{
				Mode:  mode,
				Query: query,
			})
		}
	}()
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// buildFzfCommand constructs the fzf command for the given search mode
func (a *App) buildFzfCommand(mode models.SearchMode, query string) *exec.Cmd {
	// TODO: Update TUI search to use new real-time approach during refactoring
	return exec.Command("echo", "TUI search temporarily disabled - use 'gman tools find' commands instead")
	
	/* Original implementation - commented out during refactoring
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
	*/

	/* Original implementation - commented out during refactoring
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
	*/
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

	// Create the tea program with progressive feature degradation
	options := []tea.ProgramOption{}
	
	// Always set basic I/O
	options = append(options,
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stdout),
	)
	
	// Progressively add advanced features based on terminal capabilities
	if supportsAltScreen() {
		options = append(options, tea.WithAltScreen())
	}
	
	if supportsMouseEvents() {
		options = append(options, tea.WithMouseCellMotion())
	}
	
	// Additional progressive features could be added here
	// For example: 256 color support, true color support, etc.

	p := tea.NewProgram(app, options...)
	
	// Set tea program reference for async messaging
	app.SetTeaProgram(p)

	// Run the program
	_, err := p.Run()
	return err
}

// canUseAdvancedTUI checks if advanced TUI features are available
func canUseAdvancedTUI() bool {
	// Check if stdout is a TTY
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}

	// Check TERM environment for basic compatibility
	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		return false
	}
	
	// Check for terminals that are known to not support advanced features
	incompatibleTerms := []string{"linux", "vt100", "vt102", "cons25"}
	for _, incompatible := range incompatibleTerms {
		if term == incompatible {
			return false
		}
	}

	// Check if we can access /dev/tty (Unix-like systems)
	if tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		tty.Close()
		return true
	}
	
	// For terminals that typically support advanced features
	supportedTerms := []string{
		"xterm", "xterm-256color", "xterm-color", 
		"screen", "screen-256color",
		"tmux", "tmux-256color",
		"iterm", "iterm2",
		"alacritty", "kitty", "wezterm",
		"gnome", "gnome-terminal",
		"konsole", "terminator",
	}
	
	for _, supported := range supportedTerms {
		if strings.HasPrefix(term, supported) {
			return true
		}
	}
	
	// Check for color support as a proxy for advanced features
	if strings.Contains(term, "256color") || strings.Contains(term, "color") {
		return true
	}
	
	// Conservative default: enable if we have a reasonable terminal
	// This replaces the hardcoded false return
	return len(term) > 0 && !strings.Contains(term, "dumb")
}

// supportsMouseEvents checks if the terminal supports mouse events
func supportsMouseEvents() bool {
	term := os.Getenv("TERM")
	
	// Mouse support is generally available in modern terminals
	mouseCapableTerms := []string{
		"xterm", "screen", "tmux", "iterm", "alacritty", 
		"kitty", "wezterm", "gnome", "konsole", "terminator",
	}
	
	for _, capable := range mouseCapableTerms {
		if strings.HasPrefix(term, capable) {
			return true
		}
	}
	
	return canUseAdvancedTUI() // Fallback to advanced TUI check
}

// supportsAltScreen checks if the terminal supports alternate screen buffer
func supportsAltScreen() bool {
	term := os.Getenv("TERM")
	
	// Most modern terminals support alternate screen
	altScreenCapableTerms := []string{
		"xterm", "screen", "tmux", "iterm", "alacritty",
		"kitty", "wezterm", "gnome", "konsole",
	}
	
	for _, capable := range altScreenCapableTerms {
		if strings.HasPrefix(term, capable) {
			return true
		}
	}
	
	// Check for specific terminal features
	if strings.Contains(term, "256color") {
		return true
	}
	
	return false
}

// supports256Colors checks if the terminal supports 256 colors
func supports256Colors() bool {
	term := os.Getenv("TERM")
	return strings.Contains(term, "256color") || 
		   strings.HasPrefix(term, "alacritty") ||
		   strings.HasPrefix(term, "kitty") ||
		   strings.HasPrefix(term, "iterm")
}
