package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"gman/internal/config"
	"gman/internal/di"
	"gman/pkg/types"
)

func TestDefaultValidationConfig(t *testing.T) {
	config := DefaultValidationConfig()
	
	if config == nil {
		t.Fatal("DefaultValidationConfig() returned nil")
	}
	
	if !config.RequireRepositories {
		t.Error("Expected RequireRepositories to be true by default")
	}
	
	if config.RequireGroups {
		t.Error("Expected RequireGroups to be false by default")
	}
	
	if !config.SkipInTesting {
		t.Error("Expected SkipInTesting to be true by default")
	}
	
	if config.Commands == nil {
		t.Error("Expected Commands map to be initialized")
	}
}

func TestCreatePersistentPreRunE(t *testing.T) {
	tests := []struct {
		name           string
		setupConfig    func(t *testing.T) *config.Manager
		validationCfg  *ValidationConfig
		expectError    bool
		skipValidation bool
		description    string
	}{
		{
			name:        "skip validation in testing",
			setupConfig: setupEmptyConfig,
			validationCfg: &ValidationConfig{
				RequireRepositories: true,
				SkipInTesting:       true,
			},
			expectError:    false,
			skipValidation: true,
			description:    "Should skip validation when GMAN_SKIP_REPO_CHECK is set",
		},
		{
			name:        "require repositories - success",
			setupConfig: setupConfigWithRepos,
			validationCfg: &ValidationConfig{
				RequireRepositories: true,
				SkipInTesting:       false,
			},
			expectError: false,
			description: "Should succeed when repositories are configured",
		},
		{
			name:        "require repositories - fail",
			setupConfig: setupEmptyConfig,
			validationCfg: &ValidationConfig{
				RequireRepositories: true,
				SkipInTesting:       false,
			},
			expectError: true,
			description: "Should fail when no repositories configured",
		},
		{
			name:        "require groups - success",
			setupConfig: setupConfigWithGroups,
			validationCfg: &ValidationConfig{
				RequireRepositories: true,
				RequireGroups:       true,
				SkipInTesting:       false,
			},
			expectError: false,
			description: "Should succeed when groups are configured",
		},
		{
			name:        "require groups - fail",
			setupConfig: setupConfigWithRepos,
			validationCfg: &ValidationConfig{
				RequireRepositories: true,
				RequireGroups:       true,
				SkipInTesting:       false,
			},
			expectError: true,
			description: "Should fail when no groups configured",
		},
		{
			name:        "command-specific validation",
			setupConfig: setupEmptyConfig,
			validationCfg: &ValidationConfig{
				RequireRepositories: true,
				SkipInTesting:       false,
				Commands: map[string]bool{
					"test-cmd": false, // This command doesn't need validation
				},
			},
			expectError: false,
			description: "Should skip validation for commands marked as not requiring it",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup configuration
			tt.setupConfig(t)
			
			// Set up DI container with test config
			di.Reset()
			
			// Set skip validation environment if needed
			if tt.skipValidation {
				os.Setenv("GMAN_SKIP_REPO_CHECK", "true")
				defer os.Unsetenv("GMAN_SKIP_REPO_CHECK")
			}
			
			// Create test command
			cmd := &cobra.Command{
				Use: "test-cmd",
			}
			
			// Create persistent pre-run function
			preRunE := CreatePersistentPreRunE(tt.validationCfg)
			
			// Execute validation
			err := preRunE(cmd, []string{})
			
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

func TestValidateGroupsExist(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func(t *testing.T) *config.Manager
		expectError bool
		description string
	}{
		{
			name:        "groups configured",
			setupConfig: setupConfigWithGroups,
			expectError: false,
			description: "Should succeed when groups are configured",
		},
		{
			name:        "no groups configured",
			setupConfig: setupConfigWithRepos,
			expectError: true,
			description: "Should fail when no groups are configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configMgr := tt.setupConfig(t)
			
			err := ValidateGroupsExist(configMgr)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none: %s", tt.description)
				} else if !strings.Contains(err.Error(), "no groups configured") {
					t.Errorf("Expected specific error message, got: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v (%s)", err, tt.description)
				}
			}
		})
	}
}

func TestCreateWorkValidation(t *testing.T) {
	config := CreateWorkValidation()
	
	if config == nil {
		t.Fatal("CreateWorkValidation() returned nil")
	}
	
	if !config.RequireRepositories {
		t.Error("Work commands should require repositories")
	}
	
	if config.RequireGroups {
		t.Error("Work commands should not require groups by default")
	}
	
	if !config.SkipInTesting {
		t.Error("Should skip validation in testing by default")
	}
}

func TestCreateToolsValidation(t *testing.T) {
	config := CreateToolsValidation()
	
	if config == nil {
		t.Fatal("CreateToolsValidation() returned nil")
	}
	
	if config.RequireRepositories {
		t.Error("Tools commands should not require repositories by default")
	}
	
	if config.RequireGroups {
		t.Error("Tools commands should not require groups")
	}
	
	// Check command-specific requirements
	if !config.Commands["find"] {
		t.Error("Expected 'find' command to require repositories")
	}
	
	if !config.Commands["task"] {
		t.Error("Expected 'task' command to require repositories")
	}
}

func TestCreateRepoValidation(t *testing.T) {
	config := CreateRepoValidation()
	
	if config == nil {
		t.Fatal("CreateRepoValidation() returned nil")
	}
	
	if config.RequireRepositories {
		t.Error("Repo commands should not require repositories (can work without existing repos)")
	}
	
	if config.RequireGroups {
		t.Error("Repo commands should not require groups")
	}
}

func TestNewValidateCommandContext(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	args := []string{"arg1", "arg2"}
	
	ctx := NewValidateCommandContext(cmd, args)
	
	if ctx == nil {
		t.Fatal("NewValidateCommandContext() returned nil")
	}
	
	if ctx.Command != cmd {
		t.Error("Command not preserved in context")
	}
	
	if len(ctx.Args) != len(args) {
		t.Error("Args not preserved in context")
	}
	
	if ctx.ConfigMgr == nil {
		t.Error("ConfigMgr should be initialized")
	}
	
	if ctx.RequireRepo {
		t.Error("RequireRepo should default to false")
	}
	
	if ctx.RequireGroup {
		t.Error("RequireGroup should default to false")
	}
}

func TestValidateCommandContext_WithRepositoryRequired(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	
	ctx := NewValidateCommandContext(cmd, []string{})
	ctx = ctx.WithRepositoryRequired()
	
	if !ctx.RequireRepo {
		t.Error("WithRepositoryRequired() should set RequireRepo to true")
	}
}

func TestValidateCommandContext_WithGroupRequired(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	
	ctx := NewValidateCommandContext(cmd, []string{})
	ctx = ctx.WithGroupRequired()
	
	if !ctx.RequireGroup {
		t.Error("WithGroupRequired() should set RequireGroup to true")
	}
}

func TestValidateCommandContext_Validate(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func(t *testing.T) *config.Manager
		requireRepo bool
		requireGroup bool
		expectError bool
		description string
	}{
		{
			name:        "no requirements",
			setupConfig: setupEmptyConfig,
			requireRepo: false,
			requireGroup: false,
			expectError: false,
			description: "Should succeed when no validation required",
		},
		{
			name:        "require repo - success",
			setupConfig: setupConfigWithRepos,
			requireRepo: true,
			requireGroup: false,
			expectError: false,
			description: "Should succeed when repositories configured",
		},
		{
			name:        "require repo - fail",
			setupConfig: setupEmptyConfig,
			requireRepo: true,
			requireGroup: false,
			expectError: true,
			description: "Should fail when no repositories configured",
		},
		{
			name:        "require group - success",
			setupConfig: setupConfigWithGroups,
			requireRepo: false,
			requireGroup: true,
			expectError: false,
			description: "Should succeed when groups configured",
		},
		{
			name:        "require group - fail",
			setupConfig: setupConfigWithRepos,
			requireRepo: false,
			requireGroup: true,
			expectError: true,
			description: "Should fail when no groups configured",
		},
		{
			name:        "require both - success",
			setupConfig: setupConfigWithGroups,
			requireRepo: true,
			requireGroup: true,
			expectError: false,
			description: "Should succeed when both repositories and groups configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configMgr := tt.setupConfig(t)
			
			cmd := &cobra.Command{Use: "test"}
			ctx := NewValidateCommandContext(cmd, []string{})
			
			// Override the config manager with our test one
			ctx.ConfigMgr = configMgr
			
			if tt.requireRepo {
				ctx = ctx.WithRepositoryRequired()
			}
			
			if tt.requireGroup {
				ctx = ctx.WithGroupRequired()
			}
			
			err := ctx.Validate()
			
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

// Helper functions for test setup

func setupEmptyConfig(t *testing.T) *config.Manager {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "empty_config.yml")
	
	configContent := `repositories: {}
groups: {}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write empty config: %v", err)
	}
	
	originalConfig := os.Getenv("GMAN_CONFIG")
	os.Setenv("GMAN_CONFIG", configPath)
	t.Cleanup(func() {
		if originalConfig != "" {
			os.Setenv("GMAN_CONFIG", originalConfig)
		} else {
			os.Unsetenv("GMAN_CONFIG")
		}
	})
	
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		t.Fatalf("Failed to load empty config: %v", err)
	}
	
	return configMgr
}

func setupConfigWithRepos(t *testing.T) *config.Manager {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "repos_config.yml")
	
	configContent := `repositories:
  test-repo1: /path/to/repo1
  test-repo2: /path/to/repo2
groups: {}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write repos config: %v", err)
	}
	
	originalConfig := os.Getenv("GMAN_CONFIG")
	os.Setenv("GMAN_CONFIG", configPath)
	t.Cleanup(func() {
		if originalConfig != "" {
			os.Setenv("GMAN_CONFIG", originalConfig)
		} else {
			os.Unsetenv("GMAN_CONFIG")
		}
	})
	
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		t.Fatalf("Failed to load repos config: %v", err)
	}
	
	return configMgr
}

func setupConfigWithGroups(t *testing.T) *config.Manager {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "groups_config.yml")
	
	// Create config manager and set up data
	configMgr := config.NewManager()
	cfg := configMgr.GetConfig()
	
	// Add repositories
	cfg.Repositories = map[string]string{
		"backend-api": "/path/to/backend-api",
		"frontend":    "/path/to/frontend",
	}
	
	// Add groups
	cfg.Groups = map[string]types.Group{
		"backend": {
			Name:         "backend",
			Description:  "Backend services",
			Repositories: []string{"backend-api"},
			CreatedAt:    time.Now(),
		},
		"all": {
			Name:         "all",
			Description:  "All repositories",
			Repositories: []string{"backend-api", "frontend"},
			CreatedAt:    time.Now(),
		},
	}
	
	// Set config path and save
	originalConfig := os.Getenv("GMAN_CONFIG")
	os.Setenv("GMAN_CONFIG", configPath)
	t.Cleanup(func() {
		if originalConfig != "" {
			os.Setenv("GMAN_CONFIG", originalConfig)
		} else {
			os.Unsetenv("GMAN_CONFIG")
		}
	})
	
	if err := configMgr.Save(); err != nil {
		t.Fatalf("Failed to save groups config: %v", err)
	}
	
	return configMgr
}