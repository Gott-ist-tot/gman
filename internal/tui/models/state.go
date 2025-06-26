package models

import (
	"fmt"
	"sync"
	"time"

	"gman/internal/config"
	"gman/pkg/types"
)

// AppState represents the global state of the TUI application
type AppState struct {
	mu sync.RWMutex // Protects concurrent access to state

	// Configuration
	ConfigManager *config.Manager
	Repositories  map[string]string
	Groups        map[string]types.Group

	// Current selections
	SelectedRepo     string
	SelectedGroup    string
	FocusedPanel     PanelType
	SelectedRepoData *types.RepoStatus

	// UI State
	WindowWidth  int
	WindowHeight int
	ShowHelp     bool

	// Panel states
	RepositoryListState RepositoryListState
	StatusState         StatusState
	SearchState         SearchState
	PreviewState        PreviewState

	// Background tasks
	LastStatusUpdate time.Time
	RefreshTicker    *time.Ticker

	// Error handling
	LastError     error
	ErrorVisible  bool
	ErrorMessage  string

	// Toast notifications
	ActiveToasts []ToastNotification

	// Progress indicators
	ActiveProgress map[string]ProgressIndicator
}

// PanelType represents the different panels in the dashboard
type PanelType int

const (
	RepositoryPanel PanelType = iota
	StatusPanel
	SearchPanel
	PreviewPanel
	ActionsPanel
)

func (p PanelType) String() string {
	switch p {
	case RepositoryPanel:
		return "repositories"
	case StatusPanel:
		return "status"
	case SearchPanel:
		return "search"
	case PreviewPanel:
		return "preview"
	case ActionsPanel:
		return "actions"
	default:
		return "unknown"
	}
}

// RepositoryListState holds state for the repository list panel
type RepositoryListState struct {
	FilterText    string
	FilterGroup   string
	SortBy        SortType
	ShowDirtyOnly bool
	Cursor        int
	VisibleRepos  []RepoDisplayItem
}

// StatusState holds state for the status panel
type StatusState struct {
	ShowExtended bool
	ShowBranches bool
	AutoRefresh  bool
	LastRefresh  time.Time
}

// SearchState holds state for the search panel
type SearchState struct {
	Mode         SearchMode
	Query        string
	Results      []SearchResultItem
	SelectedItem int
	IsActive     bool
	
	// Async search state
	IsSearching  bool
	Progress     int
	CurrentOp    string
	CancelFunc   func() // Function to cancel current search
}

// PreviewState holds state for the preview panel
type PreviewState struct {
	Content     string
	ContentType PreviewType
	FilePath    string
	CommitHash  string
	RepoPath    string
}

// SortType represents different sorting options
type SortType int

const (
	SortByName SortType = iota
	SortByStatus
	SortByLastUsed
	SortByModified
)

func (s SortType) String() string {
	switch s {
	case SortByName:
		return "name"
	case SortByStatus:
		return "status"
	case SortByLastUsed:
		return "last_used"
	case SortByModified:
		return "modified"
	default:
		return "name"
	}
}

// SearchMode represents different search modes
type SearchMode int

const (
	SearchFiles SearchMode = iota
	SearchCommits
)

func (s SearchMode) String() string {
	switch s {
	case SearchFiles:
		return "files"
	case SearchCommits:
		return "commits"
	default:
		return "files"
	}
}

// PreviewType represents different preview content types
type PreviewType int

const (
	PreviewFile PreviewType = iota
	PreviewCommit
	PreviewStatus
	PreviewHelp
)

// RepoDisplayItem represents a repository item for display
type RepoDisplayItem struct {
	Alias        string
	Path         string
	Status       *types.RepoStatus
	IsSelected   bool
	LastAccessed time.Time
}

// SearchResultItem represents a search result item
type SearchResultItem struct {
	Type        string
	Repository  string
	Path        string
	Hash        string
	DisplayText string
	PreviewData interface{}
}

// ToastNotification represents an active toast notification
type ToastNotification struct {
	ID        string
	Message   string
	Type      ToastType
	StartTime time.Time
	Duration  time.Duration
}

// ProgressIndicator represents an active progress indicator
type ProgressIndicator struct {
	ID            string
	Message       string
	Progress      int
	Indeterminate bool
	StartTime     time.Time
}

// NewAppState creates a new application state
func NewAppState(configMgr *config.Manager) *AppState {
	// Load configuration first
	configMgr.Load()
	config := configMgr.GetConfig()

	repos := config.Repositories
	if repos == nil {
		repos = make(map[string]string)
	}

	groups := configMgr.GetGroups()

	return &AppState{
		ConfigManager: configMgr,
		Repositories:  repos,
		Groups:        groups,
		FocusedPanel:  RepositoryPanel,
		WindowWidth:   80,
		WindowHeight:  24,
		ShowHelp:      false,

		RepositoryListState: RepositoryListState{
			SortBy:       SortByName,
			VisibleRepos: make([]RepoDisplayItem, 0),
		},
		StatusState: StatusState{
			ShowExtended: false,
			AutoRefresh:  true,
		},
		SearchState: SearchState{
			Mode: SearchFiles,
		},
		PreviewState: PreviewState{
			ContentType: PreviewHelp,
		},

		LastStatusUpdate: time.Now(),
		ActiveToasts:     make([]ToastNotification, 0),
		ActiveProgress:   make(map[string]ProgressIndicator),
	}
}

// UpdateRepositoryData updates the repository data in the state (thread-safe)
func (s *AppState) UpdateRepositoryData(alias string, status *types.RepoStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.SelectedRepoData = status

	// Update the repository in the visible list
	for i, repo := range s.RepositoryListState.VisibleRepos {
		if repo.Alias == alias {
			s.RepositoryListState.VisibleRepos[i].Status = status
			break
		}
	}
}

// GetSelectedRepository returns the currently selected repository (thread-safe)
func (s *AppState) GetSelectedRepository() *RepoDisplayItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.SelectedRepo == "" {
		return nil
	}

	for _, repo := range s.RepositoryListState.VisibleRepos {
		if repo.Alias == s.SelectedRepo {
			// Return a copy to avoid race conditions
			repoCopy := repo
			return &repoCopy
		}
	}

	return nil
}

// NextPanel moves focus to the next panel (thread-safe)
func (s *AppState) NextPanel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.FocusedPanel = (s.FocusedPanel + 1) % 5
}

// PrevPanel moves focus to the previous panel (thread-safe)
func (s *AppState) PrevPanel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.FocusedPanel = (s.FocusedPanel + 4) % 5 // +4 is equivalent to -1 mod 5
}

// SetFocusedPanel sets the focused panel (thread-safe)
func (s *AppState) SetFocusedPanel(panel PanelType) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.FocusedPanel = panel
}

// ToggleHelp toggles the help display (thread-safe)
func (s *AppState) ToggleHelp() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.ShowHelp = !s.ShowHelp
}

// UpdateWindowSize updates the window dimensions (thread-safe)
func (s *AppState) UpdateWindowSize(width, height int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.WindowWidth = width
	s.WindowHeight = height
}

// SetError sets an error in the state (thread-safe)
func (s *AppState) SetError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.LastError = err
	if err != nil {
		s.ErrorMessage = err.Error()
		s.ErrorVisible = true
	}
}

// ClearError clears any error in the state (thread-safe)
func (s *AppState) ClearError() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.LastError = nil
	s.ErrorMessage = ""
	s.ErrorVisible = false
}

// GetError returns the current error state (thread-safe)
func (s *AppState) GetError() (error, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.LastError, s.ErrorVisible
}

// SetSelectedRepo sets the selected repository (thread-safe)
func (s *AppState) SetSelectedRepo(alias string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.SelectedRepo = alias
}

// GetSelectedRepoAlias returns the selected repository alias (thread-safe)
func (s *AppState) GetSelectedRepoAlias() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.SelectedRepo
}

// UpdateSearchResults updates search state with new results (thread-safe)
func (s *AppState) UpdateSearchResults(results []SearchResultItem, query string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.SearchState.Results = results
	s.SearchState.Query = query
	if len(results) > 0 {
		s.SearchState.SelectedItem = 0
	}
}

// GetFocusedPanel returns the currently focused panel (thread-safe)
func (s *AppState) GetFocusedPanel() PanelType {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.FocusedPanel
}

// GetShowHelp returns the help display state (thread-safe)
func (s *AppState) GetShowHelp() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.ShowHelp
}

// GetWindowDimensions returns the window dimensions (thread-safe)
func (s *AppState) GetWindowDimensions() (int, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.WindowWidth, s.WindowHeight
}

// GetAutoRefresh returns the auto refresh state (thread-safe)
func (s *AppState) GetAutoRefresh() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.StatusState.AutoRefresh
}

// StartSearch starts a new search operation (thread-safe)
func (s *AppState) StartSearch(mode SearchMode, query string, cancelFunc func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Cancel any existing search
	if s.SearchState.CancelFunc != nil {
		s.SearchState.CancelFunc()
	}
	
	s.SearchState.Mode = mode
	s.SearchState.Query = query
	s.SearchState.IsSearching = true
	s.SearchState.Progress = 0
	s.SearchState.CurrentOp = "Initializing search..."
	s.SearchState.CancelFunc = cancelFunc
	s.SearchState.Results = nil
	s.SearchState.SelectedItem = 0
}

// UpdateSearchProgress updates search progress (thread-safe)
func (s *AppState) UpdateSearchProgress(progress int, currentOp string, partial []SearchResultItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.SearchState.Progress = progress
	s.SearchState.CurrentOp = currentOp
	if partial != nil {
		s.SearchState.Results = partial
	}
}

// CompleteSearch completes a search operation (thread-safe)
func (s *AppState) CompleteSearch(results []SearchResultItem, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.SearchState.IsSearching = false
	s.SearchState.Progress = 100
	s.SearchState.CancelFunc = nil
	
	if err == nil {
		s.SearchState.Results = results
		if len(results) > 0 {
			s.SearchState.SelectedItem = 0
		}
	}
}

// CancelSearch cancels the current search operation (thread-safe)
func (s *AppState) CancelSearch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.SearchState.CancelFunc != nil {
		s.SearchState.CancelFunc()
		s.SearchState.CancelFunc = nil
	}
	
	s.SearchState.IsSearching = false
	s.SearchState.Progress = 0
	s.SearchState.CurrentOp = ""
}

// GetSearchState returns the current search state (thread-safe)
func (s *AppState) GetSearchState() SearchState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	return SearchState{
		Mode:         s.SearchState.Mode,
		Query:        s.SearchState.Query,
		Results:      append([]SearchResultItem{}, s.SearchState.Results...),
		SelectedItem: s.SearchState.SelectedItem,
		IsActive:     s.SearchState.IsActive,
		IsSearching:  s.SearchState.IsSearching,
		Progress:     s.SearchState.Progress,
		CurrentOp:    s.SearchState.CurrentOp,
	}
}

// AddToast adds a new toast notification (thread-safe)
func (s *AppState) AddToast(message string, toastType ToastType, duration time.Duration) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	id := fmt.Sprintf("toast_%d", time.Now().UnixNano())
	toast := ToastNotification{
		ID:        id,
		Message:   message,
		Type:      toastType,
		StartTime: time.Now(),
		Duration:  duration,
	}
	
	s.ActiveToasts = append(s.ActiveToasts, toast)
	
	// Clean up expired toasts
	s.cleanupExpiredToasts()
	
	return id
}

// RemoveToast removes a toast notification by ID (thread-safe)
func (s *AppState) RemoveToast(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for i, toast := range s.ActiveToasts {
		if toast.ID == id {
			s.ActiveToasts = append(s.ActiveToasts[:i], s.ActiveToasts[i+1:]...)
			break
		}
	}
}

// GetActiveToasts returns active toast notifications (thread-safe)
func (s *AppState) GetActiveToasts() []ToastNotification {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Clean up expired toasts
	s.cleanupExpiredToasts()
	
	// Return a copy to avoid race conditions
	toasts := make([]ToastNotification, len(s.ActiveToasts))
	copy(toasts, s.ActiveToasts)
	return toasts
}

// cleanupExpiredToasts removes expired toasts (must be called with lock held)
func (s *AppState) cleanupExpiredToasts() {
	now := time.Now()
	validToasts := make([]ToastNotification, 0, len(s.ActiveToasts))
	
	for _, toast := range s.ActiveToasts {
		if now.Sub(toast.StartTime) < toast.Duration {
			validToasts = append(validToasts, toast)
		}
	}
	
	s.ActiveToasts = validToasts
}

// SetProgress adds or updates a progress indicator (thread-safe)
func (s *AppState) SetProgress(id, message string, progress int, indeterminate bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.ActiveProgress[id] = ProgressIndicator{
		ID:            id,
		Message:       message,
		Progress:      progress,
		Indeterminate: indeterminate,
		StartTime:     time.Now(),
	}
}

// RemoveProgress removes a progress indicator (thread-safe)
func (s *AppState) RemoveProgress(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.ActiveProgress, id)
}

// GetActiveProgress returns active progress indicators (thread-safe)
func (s *AppState) GetActiveProgress() map[string]ProgressIndicator {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	progress := make(map[string]ProgressIndicator, len(s.ActiveProgress))
	for k, v := range s.ActiveProgress {
		progress[k] = v
	}
	return progress
}
