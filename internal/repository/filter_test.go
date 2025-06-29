package repository

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gman/internal/config"
	"gman/internal/di"
	"gman/pkg/types"
)

func TestNewFilter(t *testing.T) {
	configMgr := setupTestConfig(t)
	
	filter := NewFilter(configMgr)
	
	if filter == nil {
		t.Fatal("NewFilter() returned nil")
	}
	
	if filter.configMgr != configMgr {
		t.Error("NewFilter() did not preserve config manager reference")
	}
}

func TestFilter_FilterByGroup(t *testing.T) {
	tests := []struct {
		name           string
		repositories   map[string]string
		groupFilter    string
		expectedRepos  map[string]string
		expectError    bool
		description    string
	}{
		{
			name: "no group filter",
			repositories: map[string]string{
				"repo1": "/path/to/repo1",
				"repo2": "/path/to/repo2",
			},
			groupFilter: "",
			expectedRepos: map[string]string{
				"repo1": "/path/to/repo1", 
				"repo2": "/path/to/repo2",
			},
			expectError: false,
			description: "Should return all repositories when no group filter specified",
		},
		{
			name: "valid group filter",
			repositories: map[string]string{
				"backend-api":  "/path/to/backend-api",
				"backend-db":   "/path/to/backend-db", 
				"frontend-web": "/path/to/frontend-web",
			},
			groupFilter: "backend",
			expectedRepos: map[string]string{
				"backend-api": "/path/to/backend-api",
				"backend-db":  "/path/to/backend-db",
			},
			expectError: false,
			description: "Should return only repositories in the specified group",
		},
		{
			name: "nonexistent group filter",
			repositories: map[string]string{
				"repo1": "/path/to/repo1",
			},
			groupFilter:   "nonexistent",
			expectedRepos: map[string]string{}, // empty map for backward compatibility
			expectError:   false,
			description:   "Should return empty map for nonexistent group (backward compatibility)",
		},
		{
			name: "empty group",
			repositories: map[string]string{
				"repo1": "/path/to/repo1",
			},
			groupFilter:   "empty-group",
			expectedRepos: map[string]string{}, // empty map for empty groups
			expectError:   false,
			description:   "Should return empty map for empty groups",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configMgr := setupTestConfigWithGroups(t, tt.repositories)
			filter := NewFilter(configMgr)
			
			result, err := filter.FilterByGroup(tt.repositories, tt.groupFilter)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none: %s", tt.description)
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v (%s)", err, tt.description)
				return
			}
			
			if len(result) != len(tt.expectedRepos) {
				t.Errorf("Expected %d repositories, got %d (%s)", len(tt.expectedRepos), len(result), tt.description)
				return
			}
			
			for alias, expectedPath := range tt.expectedRepos {
				if resultPath, exists := result[alias]; !exists {
					t.Errorf("Expected repository '%s' not found in result (%s)", alias, tt.description)
				} else if resultPath != expectedPath {
					t.Errorf("Repository '%s' path mismatch: expected '%s', got '%s' (%s)", alias, expectedPath, resultPath, tt.description)
				}
			}
		})
	}
}

func TestFilter_FilterByGroupStrict(t *testing.T) {
	tests := []struct {
		name        string
		groupFilter string
		expectError bool
		errorText   string
		description string
	}{
		{
			name:        "valid group",
			groupFilter: "backend",
			expectError: false,
			description: "Should succeed for valid group with repositories",
		},
		{
			name:        "nonexistent group",
			groupFilter: "nonexistent",
			expectError: true,
			errorText:   "group 'nonexistent' is empty or does not exist",
			description: "Should error for nonexistent group",
		},
		{
			name:        "empty group",
			groupFilter: "empty-group",
			expectError: true,
			errorText:   "group 'empty-group' is empty or does not exist",
			description: "Should error for empty group",
		},
		{
			name:        "no filter",
			groupFilter: "",
			expectError: false,
			description: "Should succeed with no filter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repositories := map[string]string{
				"backend-api": "/path/to/backend-api",
				"backend-db":  "/path/to/backend-db",
				"frontend":    "/path/to/frontend",
			}
			
			configMgr := setupTestConfigWithGroups(t, repositories)
			filter := NewFilter(configMgr)
			
			result, err := filter.FilterByGroupStrict(repositories, tt.groupFilter)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none: %s", tt.description)
				} else if tt.errorText != "" && !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error to contain '%s', got '%s' (%s)", tt.errorText, err.Error(), tt.description)
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v (%s)", err, tt.description)
				return
			}
			
			if tt.groupFilter == "" {
				// Should return all repositories
				if len(result) != len(repositories) {
					t.Errorf("Expected all repositories when no filter, got %d of %d (%s)", len(result), len(repositories), tt.description)
				}
			} else {
				// Should return only group repositories
				if len(result) == 0 {
					t.Errorf("Expected group repositories, got empty result (%s)", tt.description)
				}
			}
		})
	}
}

func TestFilter_FilterByGroupWithValidation(t *testing.T) {
	tests := []struct {
		name         string
		repositories map[string]string
		groupRepos   map[string]string
		groupFilter  string
		expectError  bool
		errorText    string
		description  string
	}{
		{
			name: "valid group with matching repositories",
			repositories: map[string]string{
				"backend-api": "/path/to/backend-api",
				"backend-db":  "/path/to/backend-db",
				"frontend":    "/path/to/frontend",
			},
			groupFilter: "backend",
			expectError: false,
			description: "Should succeed when group repositories exist in main list",
		},
		{
			name: "group repository not in main list",
			repositories: map[string]string{
				"frontend": "/path/to/frontend",
			},
			groupFilter: "backend",
			expectError: true,
			errorText:   "group repository 'backend-api' not found in main repository list",
			description: "Should error when group repository not in main list",
		},
		{
			name: "path mismatch between group and main",
			repositories: map[string]string{
				"backend-api": "/different/path/to/backend-api",
			},
			groupFilter: "backend",
			expectError: true,
			errorText:   "path mismatch",
			description: "Should error when paths don't match between group and main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configMgr := setupTestConfigWithGroups(t, tt.repositories)
			filter := NewFilter(configMgr)
			
			result, err := filter.FilterByGroupWithValidation(tt.repositories, tt.groupFilter)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none: %s", tt.description)
				} else if tt.errorText != "" && !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error to contain '%s', got '%s' (%s)", tt.errorText, err.Error(), tt.description)
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v (%s)", err, tt.description)
				return
			}
			
			if result == nil {
				t.Errorf("Expected non-nil result (%s)", tt.description)
			}
		})
	}
}

func TestFilter_GetGroupNames(t *testing.T) {
	repositories := map[string]string{
		"backend-api":  "/path/to/backend-api",
		"frontend-web": "/path/to/frontend-web",
	}
	
	configMgr := setupTestConfigWithGroups(t, repositories)
	filter := NewFilter(configMgr)
	
	names := filter.GetGroupNames()
	
	expectedGroups := []string{"backend", "frontend", "empty-group"}
	if len(names) != len(expectedGroups) {
		t.Errorf("Expected %d group names, got %d", len(expectedGroups), len(names))
	}
	
	// Check that all expected groups are present
	nameSet := make(map[string]bool)
	for _, name := range names {
		nameSet[name] = true
	}
	
	for _, expected := range expectedGroups {
		if !nameSet[expected] {
			t.Errorf("Expected group '%s' not found in result", expected)
		}
	}
}

func TestFilter_ValidateGroupExists(t *testing.T) {
	repositories := map[string]string{
		"backend-api": "/path/to/backend-api",
	}
	
	configMgr := setupTestConfigWithGroups(t, repositories)
	filter := NewFilter(configMgr)
	
	// Test existing group
	err := filter.ValidateGroupExists("backend")
	if err != nil {
		t.Errorf("Unexpected error for existing group: %v", err)
	}
	
	// Test nonexistent group
	err = filter.ValidateGroupExists("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent group")
	} else if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected 'does not exist' in error message, got: %v", err)
	}
}

func TestFilter_GetRepositoryCount(t *testing.T) {
	repositories := map[string]string{
		"backend-api": "/path/to/backend-api",
		"backend-db":  "/path/to/backend-db",
		"frontend":    "/path/to/frontend",
	}
	
	configMgr := setupTestConfigWithGroups(t, repositories)
	filter := NewFilter(configMgr)
	
	// Test valid group
	count, err := filter.GetRepositoryCount("backend")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2 for backend group, got %d", count)
	}
	
	// Test empty group
	count, err = filter.GetRepositoryCount("empty-group")
	if err != nil {
		t.Errorf("Unexpected error for empty group: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0 for empty group, got %d", count)
	}
	
	// Test nonexistent group
	_, err = filter.GetRepositoryCount("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent group")
	}
}

func TestFilter_FilterWithInfo(t *testing.T) {
	repositories := map[string]string{
		"backend-api":  "/path/to/backend-api",
		"backend-db":   "/path/to/backend-db",
		"frontend-web": "/path/to/frontend-web",
	}
	
	configMgr := setupTestConfigWithGroups(t, repositories)
	filter := NewFilter(configMgr)
	
	// Test with group filter
	result, info, err := filter.FilterWithInfo(repositories, "backend")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if info.OriginalCount != 3 {
		t.Errorf("Expected original count 3, got %d", info.OriginalCount)
	}
	
	if info.FilteredCount != 2 {
		t.Errorf("Expected filtered count 2, got %d", info.FilteredCount)
	}
	
	if info.GroupName != "backend" {
		t.Errorf("Expected group name 'backend', got '%s'", info.GroupName)
	}
	
	if !info.Applied {
		t.Error("Expected Applied to be true when group filter specified")
	}
	
	if len(result) != 2 {
		t.Errorf("Expected 2 repositories in result, got %d", len(result))
	}
	
	// Test without group filter
	result, info, err = filter.FilterWithInfo(repositories, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if info.Applied {
		t.Error("Expected Applied to be false when no group filter specified")
	}
	
	if info.OriginalCount != info.FilteredCount {
		t.Error("Expected original and filtered counts to be equal when no filter applied")
	}
}

// Helper functions for test setup

func setupTestConfig(t *testing.T) *config.Manager {
	// Reset DI container
	di.Reset()
	
	// Create temporary config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yml")
	
	configContent := `repositories:
  test-repo: /path/to/test-repo
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	
	// Set environment
	os.Setenv("GMAN_CONFIG", configPath)
	t.Cleanup(func() {
		os.Unsetenv("GMAN_CONFIG")
	})
	
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}
	
	return configMgr
}

func setupTestConfigWithGroups(t *testing.T, repositories map[string]string) *config.Manager {
	// Reset DI container
	di.Reset()
	
	// Create test configuration with groups
	configMgr := config.NewManager()
	cfg := configMgr.GetConfig()
	cfg.Repositories = repositories
	
	// Add test groups
	cfg.Groups = map[string]types.Group{
		"backend": {
			Name:         "backend",
			Description:  "Backend services",
			Repositories: []string{"backend-api", "backend-db"},
			CreatedAt:    time.Now(),
		},
		"frontend": {
			Name:         "frontend",
			Description:  "Frontend applications",
			Repositories: []string{"frontend-web"},
			CreatedAt:    time.Now(),
		},
		"empty-group": {
			Name:         "empty-group",
			Description:  "Empty test group",
			Repositories: []string{},
			CreatedAt:    time.Now(),
		},
	}
	
	// Save configuration
	if err := configMgr.Save(); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}
	
	return configMgr
}