package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gman/internal/di"
	"gman/pkg/types"
)

func TestGetManagers(t *testing.T) {
	// Reset DI container to ensure clean state
	di.Reset()

	mgrs := GetManagers()

	if mgrs == nil {
		t.Fatal("GetManagers() returned nil")
	}

	if mgrs.Config == nil {
		t.Error("GetManagers().Config is nil")
	}

	if mgrs.Git == nil {
		t.Error("GetManagers().Git is nil")
	}

	// Verify that multiple calls return different instances but are properly initialized
	mgrs2 := GetManagers()
	if mgrs2 == nil {
		t.Fatal("Second GetManagers() call returned nil")
	}

	// The struct instances should be different, but the underlying managers should be singletons
	if mgrs == mgrs2 {
		t.Error("GetManagers() should return different struct instances")
	}
}

func TestGetManagersWithConfig(t *testing.T) {
	// Create temporary config for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yml")

	// Create a valid test configuration
	configContent := `repositories:
  test-repo: /path/to/test-repo
groups:
  test-group:
    name: test-group
    description: Test group
    repositories: [test-repo]
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Set environment variable for config path
	originalConfig := os.Getenv("GMAN_CONFIG")
	os.Setenv("GMAN_CONFIG", configPath)
	defer func() {
		if originalConfig != "" {
			os.Setenv("GMAN_CONFIG", originalConfig)
		} else {
			os.Unsetenv("GMAN_CONFIG")
		}
	}()

	// Reset DI container
	di.Reset()

	mgrs, err := GetManagersWithConfig()
	if err != nil {
		t.Fatalf("GetManagersWithConfig() failed: %v", err)
	}

	if mgrs == nil {
		t.Fatal("GetManagersWithConfig() returned nil managers")
	}

	// Verify config was loaded
	cfg := mgrs.Config.GetConfig()
	if len(cfg.Repositories) == 0 {
		t.Error("Configuration was not properly loaded - no repositories found")
	}

	if _, exists := cfg.Repositories["test-repo"]; !exists {
		t.Error("Expected test-repo to be configured")
	}
}

func TestGetManagersWithConfig_InvalidConfig(t *testing.T) {
	// Create invalid config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid_config.yml")

	// Write invalid YAML
	invalidContent := `repositories:
  test-repo: /path
invalid: yaml: structure:
  - malformed
`
	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Set environment variable
	originalConfig := os.Getenv("GMAN_CONFIG")
	os.Setenv("GMAN_CONFIG", configPath)
	defer func() {
		if originalConfig != "" {
			os.Setenv("GMAN_CONFIG", originalConfig)
		} else {
			os.Unsetenv("GMAN_CONFIG")
		}
	}()

	// Reset DI container
	di.Reset()

	mgrs, err := GetManagersWithConfig()
	if err == nil {
		t.Error("Expected error for invalid configuration, but got none")
	}

	if mgrs != nil {
		t.Error("Expected nil managers for invalid configuration")
	}

	// Error message should indicate configuration loading failure
	if err != nil && !strings.Contains(err.Error(), "failed to load configuration") {
		t.Errorf("Error message should mention config loading failure, got: %v", err)
	}
}

func TestValidateRepositoriesExist(t *testing.T) {
	tests := []struct {
		name        string
		repositories map[string]string
		expectError bool
	}{
		{
			name:        "no repositories configured",
			repositories: map[string]string{},
			expectError: true,
		},
		{
			name: "repositories configured",
			repositories: map[string]string{
				"test-repo": "/path/to/repo",
			},
			expectError: false,
		},
		{
			name: "multiple repositories configured",
			repositories: map[string]string{
				"repo1": "/path/to/repo1",
				"repo2": "/path/to/repo2",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset DI container
			di.Reset()

			// Configure test repositories
			mgrs := GetManagers()
			cfg := mgrs.Config.GetConfig()
			cfg.Repositories = tt.repositories

			err := ValidateRepositoriesExist(mgrs.Config)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), "no repositories configured") {
					t.Errorf("Expected specific error message, got: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateRepositoriesWithGroups(t *testing.T) {
	tests := []struct {
		name         string
		repositories map[string]string
		groups       map[string]types.Group
		expectError  bool
		description  string
	}{
		{
			name:        "no repositories",
			repositories: map[string]string{},
			groups:      map[string]types.Group{},
			expectError: true,
			description: "Should fail when no repositories configured",
		},
		{
			name: "repositories without groups",
			repositories: map[string]string{
				"test-repo": "/path/to/repo",
			},
			groups:      map[string]types.Group{},
			expectError: false,
			description: "Should succeed with repositories but no groups (groups are optional)",
		},
		{
			name: "repositories with groups",
			repositories: map[string]string{
				"test-repo": "/path/to/repo",
			},
			groups: map[string]types.Group{
				"test-group": {
					Name:         "test-group",
					Description:  "Test group",
					Repositories: []string{"test-repo"},
				},
			},
			expectError: false,
			description: "Should succeed with both repositories and groups",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset DI container
			di.Reset()

			// Configure test data
			mgrs := GetManagers()
			cfg := mgrs.Config.GetConfig()
			cfg.Repositories = tt.repositories
			cfg.Groups = tt.groups

			err := ValidateRepositoriesWithGroups(mgrs.Config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none: %s", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v (%s)", err, tt.description)
				}
			}
		})
	}
}

func TestFormatOperationError(t *testing.T) {
	baseErr := &testError{message: "base error"}
	
	err := FormatOperationError("test operation", baseErr)
	
	if err == nil {
		t.Fatal("FormatOperationError() returned nil")
	}

	expected := "failed to test operation: base error"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}

	// Test unwrapping behavior
	if unwrapped := err.(interface{ Unwrap() error }).Unwrap(); unwrapped != baseErr {
		t.Error("FormatOperationError() should preserve error wrapping")
	}
}

func TestFormatValidationError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    string
		reason   string
		expected string
	}{
		{
			name:     "repository validation",
			field:    "repository",
			value:    "invalid-repo",
			reason:   "not found",
			expected: "invalid repository 'invalid-repo': not found",
		},
		{
			name:     "group validation",
			field:    "group",
			value:    "test-group",
			reason:   "already exists",
			expected: "invalid group 'test-group': already exists",
		},
		{
			name:     "path validation",
			field:    "path",
			value:    "/invalid/path",
			reason:   "does not exist",
			expected: "invalid path '/invalid/path': does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FormatValidationError(tt.field, tt.value, tt.reason)
			
			if err == nil {
				t.Fatal("FormatValidationError() returned nil")
			}

			if err.Error() != tt.expected {
				t.Errorf("Expected error message '%s', got '%s'", tt.expected, err.Error())
			}
		})
	}
}

func TestFormatNotFoundError(t *testing.T) {
	tests := []struct {
		name       string
		resource   string
		identifier string
		expected   string
	}{
		{
			name:       "repository not found",
			resource:   "repository",
			identifier: "test-repo",
			expected:   "repository 'test-repo' not found",
		},
		{
			name:       "group not found",
			resource:   "group",
			identifier: "test-group",
			expected:   "group 'test-group' not found",
		},
		{
			name:       "branch not found",
			resource:   "branch",
			identifier: "feature-branch",
			expected:   "branch 'feature-branch' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FormatNotFoundError(tt.resource, tt.identifier)
			
			if err == nil {
				t.Fatal("FormatNotFoundError() returned nil")
			}

			if err.Error() != tt.expected {
				t.Errorf("Expected error message '%s', got '%s'", tt.expected, err.Error())
			}
		})
	}
}

// Helper types for testing

type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}