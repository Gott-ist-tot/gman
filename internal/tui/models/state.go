package models

import (
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

// NextPanel moves focus to the next panel
func (s *AppState) NextPanel() {
	s.FocusedPanel = (s.FocusedPanel + 1) % 5
}

// PrevPanel moves focus to the previous panel
func (s *AppState) PrevPanel() {
	s.FocusedPanel = (s.FocusedPanel + 4) % 5 // +4 is equivalent to -1 mod 5
}

// SetFocusedPanel sets the focused panel
func (s *AppState) SetFocusedPanel(panel PanelType) {
	s.FocusedPanel = panel
}

// ToggleHelp toggles the help display
func (s *AppState) ToggleHelp() {
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
