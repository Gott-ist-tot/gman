package panels

import (
	"testing"
	"time"

	"gman/internal/config"
	"gman/internal/tui/models"
	"gman/pkg/types"
)

// TestRepositorySortingWithNilStatus tests the nil pointer fix in sortRepositories
func TestRepositorySortingWithNilStatus(t *testing.T) {
	// Create a mock config manager
	configMgr := &config.Manager{}
	
	// Create app state
	state := &models.AppState{
		ConfigManager: configMgr,
		Repositories:  make(map[string]string),
		Groups:        make(map[string]types.Group),
		FocusedPanel:  models.RepositoryPanel,
		RepositoryListState: models.RepositoryListState{
			SortBy:       models.SortByModified,
			VisibleRepos: make([]models.RepoDisplayItem, 0),
		},
	}

	// Create repository panel
	panel := &RepositoryPanel{
		state: state,
		repos: []models.RepoDisplayItem{
			{
				Alias:        "repo-with-status",
				Path:         "/path/to/repo1",
				Status:       &types.RepoStatus{CommitTime: time.Now()},
				IsSelected:   false,
				LastAccessed: time.Now(),
			},
			{
				Alias:        "repo-without-status",
				Path:         "/path/to/repo2",
				Status:       nil, // This used to cause nil pointer dereference
				IsSelected:   false,
				LastAccessed: time.Now(),
			},
			{
				Alias:        "repo-with-older-status",
				Path:         "/path/to/repo3",
				Status:       &types.RepoStatus{CommitTime: time.Now().Add(-time.Hour)},
				IsSelected:   false,
				LastAccessed: time.Now(),
			},
		},
	}

	// Test that sorting doesn't panic with nil status
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("sortRepositories panicked with nil status: %v", r)
		}
	}()

	// This should not panic
	panel.sortRepositories()

	// Verify the sorting results
	if len(panel.repos) != 3 {
		t.Errorf("Expected 3 repositories, got %d", len(panel.repos))
	}

	// Verify that repos with status are prioritized over nil status
	// and among those with status, newer commits come first
	expectedOrder := []string{"repo-with-status", "repo-with-older-status", "repo-without-status"}
	for i, expectedAlias := range expectedOrder {
		if panel.repos[i].Alias != expectedAlias {
			t.Errorf("Expected repo at position %d to be %s, got %s", i, expectedAlias, panel.repos[i].Alias)
		}
	}
}

// TestRepositorySortingAllNilStatus tests sorting when all repositories have nil status
func TestRepositorySortingAllNilStatus(t *testing.T) {
	// Create a mock config manager
	configMgr := &config.Manager{}
	
	// Create app state
	state := &models.AppState{
		ConfigManager: configMgr,
		Repositories:  make(map[string]string),
		Groups:        make(map[string]types.Group),
		FocusedPanel:  models.RepositoryPanel,
		RepositoryListState: models.RepositoryListState{
			SortBy:       models.SortByModified,
			VisibleRepos: make([]models.RepoDisplayItem, 0),
		},
	}

	// Create repository panel with all nil status
	panel := &RepositoryPanel{
		state: state,
		repos: []models.RepoDisplayItem{
			{
				Alias:        "zzz-repo",
				Path:         "/path/to/zzz",
				Status:       nil,
				IsSelected:   false,
				LastAccessed: time.Now(),
			},
			{
				Alias:        "aaa-repo",
				Path:         "/path/to/aaa",
				Status:       nil,
				IsSelected:   false,
				LastAccessed: time.Now(),
			},
			{
				Alias:        "mmm-repo",
				Path:         "/path/to/mmm",
				Status:       nil,
				IsSelected:   false,
				LastAccessed: time.Now(),
			},
		},
	}

	// Test that sorting doesn't panic with all nil status
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("sortRepositories panicked with all nil status: %v", r)
		}
	}()

	// This should not panic and should fall back to name sorting
	panel.sortRepositories()

	// Verify alphabetical sorting by name when status is nil
	expectedOrder := []string{"aaa-repo", "mmm-repo", "zzz-repo"}
	for i, expectedAlias := range expectedOrder {
		if panel.repos[i].Alias != expectedAlias {
			t.Errorf("Expected repo at position %d to be %s, got %s", i, expectedAlias, panel.repos[i].Alias)
		}
	}
}

// TestRepositorySortingOtherModes tests that other sorting modes work correctly
func TestRepositorySortingOtherModes(t *testing.T) {
	// Create a mock config manager
	configMgr := &config.Manager{}
	
	sortModes := []models.SortType{
		models.SortByName,
		models.SortByStatus,
		models.SortByLastUsed,
	}

	for _, sortMode := range sortModes {
		t.Run(sortMode.String(), func(t *testing.T) {
			// Create app state
			state := &models.AppState{
				ConfigManager: configMgr,
				Repositories:  make(map[string]string),
				Groups:        make(map[string]types.Group),
				FocusedPanel:  models.RepositoryPanel,
				RepositoryListState: models.RepositoryListState{
					SortBy:       sortMode,
					VisibleRepos: make([]models.RepoDisplayItem, 0),
				},
			}

			// Create repository panel
			panel := &RepositoryPanel{
				state: state,
				repos: []models.RepoDisplayItem{
					{
						Alias:        "repo1",
						Path:         "/path/to/repo1",
						Status:       &types.RepoStatus{Workspace: types.Clean},
						IsSelected:   false,
						LastAccessed: time.Now(),
					},
					{
						Alias:        "repo2",
						Path:         "/path/to/repo2",
						Status:       nil,
						IsSelected:   false,
						LastAccessed: time.Now().Add(-time.Hour),
					},
				},
			}

			// Test that sorting doesn't panic with any sort mode
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("sortRepositories panicked with sort mode %s: %v", sortMode.String(), r)
				}
			}()

			// This should not panic regardless of sort mode
			panel.sortRepositories()

			// Just verify that we still have the same number of repos
			if len(panel.repos) != 2 {
				t.Errorf("Expected 2 repositories after sorting, got %d", len(panel.repos))
			}
		})
	}
}