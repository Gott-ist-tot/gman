package models

import (
	"time"

	"gman/pkg/types"

	tea "github.com/charmbracelet/bubbletea"
)

// WindowSizeMsg is sent when the terminal window is resized
type WindowSizeMsg struct {
	Width  int
	Height int
}

// RepositorySelectedMsg is sent when a repository is selected
type RepositorySelectedMsg struct {
	Alias string
	Path  string
}

// RepositoryStatusMsg is sent when repository status is updated (initial fast load)
type RepositoryStatusMsg struct {
	Alias  string
	Status *types.RepoStatus
	Error  error
}

// RepositoryStatusRefreshMsg is sent when repository status is refreshed with full fetch
type RepositoryStatusRefreshMsg struct {
	Alias  string
	Status *types.RepoStatus
	Error  error
}

// SearchResultsMsg is sent when search results are available
type SearchResultsMsg struct {
	Mode    SearchMode
	Query   string
	Results []SearchResultItem
	Error   error
}

// SearchStartedMsg is sent when a search operation begins
type SearchStartedMsg struct {
	Mode  SearchMode
	Query string
}

// SearchProgressMsg is sent periodically during search operations
type SearchProgressMsg struct {
	Mode       SearchMode
	Query      string
	Progress   int // percentage completed
	CurrentOp  string // current operation description
	Partial    []SearchResultItem // partial results
}

// SearchCancelledMsg is sent when a search is cancelled
type SearchCancelledMsg struct {
	Mode  SearchMode
	Query string
}

// PreviewContentMsg is sent when preview content is ready
type PreviewContentMsg struct {
	Content     string
	ContentType PreviewType
	FilePath    string
	CommitHash  string
	Error       error
}

// FilterChangedMsg is sent when repository filter changes
type FilterChangedMsg struct {
	FilterText  string
	FilterGroup string
}

// SortChangedMsg is sent when sort order changes
type SortChangedMsg struct {
	SortBy SortType
}

// PanelFocusMsg is sent when panel focus changes
type PanelFocusMsg struct {
	Panel PanelType
}

// RefreshMsg is sent to trigger a refresh of repository status
type RefreshMsg struct {
	Force bool
}

// StatusTickMsg is sent periodically to update status
type StatusTickMsg time.Time

// SearchModeMsg is sent when search mode changes
type SearchModeMsg struct {
	Mode SearchMode
}

// ErrorMsg is sent when an error occurs
type ErrorMsg struct {
	Error   error
	Context string // Additional context for the error
	Fatal   bool   // Whether this error should cause the app to exit
}

// CommandExecutedMsg is sent when a command is executed
type CommandExecutedMsg struct {
	Command string
	Success bool
	Output  string
	Error   error
}

// HelpToggleMsg is sent when help is toggled
type HelpToggleMsg struct{}

// ActionCompleteMsg is sent when an action completes
type ActionCompleteMsg struct {
	Result string
	Error  error
}

// HideResultMsg is sent to hide action results
type HideResultMsg struct{}

// ExitMsg is sent when the application should exit
type ExitMsg struct{}

// FzfLaunchMsg is sent when fzf should be launched
type FzfLaunchMsg struct {
	Mode  SearchMode
	Query string
}

// BackgroundTaskMsg represents background task completion
type BackgroundTaskMsg struct {
	TaskType string
	Data     interface{}
	Error    error
}

// Utility functions for creating commands

// WindowSizeCmd returns a command that sends a WindowSizeMsg
func WindowSizeCmd(width, height int) tea.Cmd {
	return func() tea.Msg {
		return WindowSizeMsg{Width: width, Height: height}
	}
}

// RepositorySelectedCmd returns a command that sends a RepositorySelectedMsg
func RepositorySelectedCmd(alias, path string) tea.Cmd {
	return func() tea.Msg {
		return RepositorySelectedMsg{Alias: alias, Path: path}
	}
}

// RefreshCmd returns a command that sends a RefreshMsg
func RefreshCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		return RefreshMsg{Force: force}
	}
}

// StatusTickCmd returns a command that sends a StatusTickMsg
func StatusTickCmd() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return StatusTickMsg(t)
	})
}

// ErrorCmd returns a command that sends an ErrorMsg
func ErrorCmd(err error) tea.Cmd {
	return ErrorCmdWithContext(err, "", false)
}

// ErrorCmdWithContext returns a command that sends an ErrorMsg with context
func ErrorCmdWithContext(err error, context string, fatal bool) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg{
			Error:   err,
			Context: context,
			Fatal:   fatal,
		}
	}
}

// ExitCmd returns a command that sends an ExitMsg
func ExitCmd() tea.Cmd {
	return func() tea.Msg {
		return ExitMsg{}
	}
}

// SearchModeCmd returns a command that changes search mode
func SearchModeCmd(mode SearchMode) tea.Cmd {
	return func() tea.Msg {
		return SearchModeMsg{Mode: mode}
	}
}

// PanelFocusCmd returns a command that changes panel focus
func PanelFocusCmd(panel PanelType) tea.Cmd {
	return func() tea.Msg {
		return PanelFocusMsg{Panel: panel}
	}
}

// FilterChangedCmd returns a command that sends a FilterChangedMsg
func FilterChangedCmd(filterText, filterGroup string) tea.Cmd {
	return func() tea.Msg {
		return FilterChangedMsg{
			FilterText:  filterText,
			FilterGroup: filterGroup,
		}
	}
}

// SortChangedCmd returns a command that sends a SortChangedMsg
func SortChangedCmd(sortBy SortType) tea.Cmd {
	return func() tea.Msg {
		return SortChangedMsg{SortBy: sortBy}
	}
}

// HelpToggleCmd returns a command that toggles help
func HelpToggleCmd() tea.Cmd {
	return func() tea.Msg {
		return HelpToggleMsg{}
	}
}
