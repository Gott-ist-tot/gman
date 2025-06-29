package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gman/internal/config"
	"gman/internal/di"
)

// ValidationConfig holds configuration for command validation
type ValidationConfig struct {
	RequireRepositories bool
	RequireGroups       bool
	SkipInTesting       bool
	Commands            map[string]bool // commands that need specific validation
}

// DefaultValidationConfig returns standard validation configuration
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		RequireRepositories: true,
		RequireGroups:       false,
		SkipInTesting:       true,
		Commands:            make(map[string]bool),
	}
}

// CreatePersistentPreRunE creates a standardized PersistentPreRunE function
// This consolidates the repeated validation patterns across command groups
func CreatePersistentPreRunE(config *ValidationConfig) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Call parent's PersistentPreRunE first to ensure config is loaded
		if cmd.Parent() != nil && cmd.Parent().PersistentPreRunE != nil {
			if err := cmd.Parent().PersistentPreRunE(cmd, args); err != nil {
				return err
			}
		}

		// Skip validation during testing or if explicitly disabled
		if config.SkipInTesting && os.Getenv("GMAN_SKIP_REPO_CHECK") == "true" {
			return nil
		}

		// Get configuration manager
		configMgr := di.ConfigManager()

		// Validate repositories if required
		if config.RequireRepositories {
			// Check if specific commands need repository validation
			if len(config.Commands) > 0 {
				if !config.Commands[cmd.Name()] {
					return nil // Skip validation for commands that don't need it
				}
			}

			if err := ValidateRepositoriesExist(configMgr); err != nil {
				return err
			}
		}

		// Validate groups if required
		if config.RequireGroups {
			if err := ValidateGroupsExist(configMgr); err != nil {
				return err
			}
		}

		return nil
	}
}

// ValidateGroupsExist checks if groups are configured
func ValidateGroupsExist(configMgr *config.Manager) error {
	cfg := configMgr.GetConfig()
	if len(cfg.Groups) == 0 {
		return fmt.Errorf("no groups configured. Use 'gman repo group create' to create groups")
	}
	return nil
}

// CreateWorkValidation creates validation config for work commands
func CreateWorkValidation() *ValidationConfig {
	return &ValidationConfig{
		RequireRepositories: true,
		RequireGroups:       false,
		SkipInTesting:       true,
		Commands:            make(map[string]bool), // All work commands need repos
	}
}

// CreateToolsValidation creates validation config for tools commands
func CreateToolsValidation() *ValidationConfig {
	return &ValidationConfig{
		RequireRepositories: false, // Will be checked per command
		RequireGroups:       false,
		SkipInTesting:       true,
		Commands: map[string]bool{
			"find": true,
			"task": true,
			// setup, health, init don't require repositories
		},
	}
}

// CreateRepoValidation creates validation config for repo commands
func CreateRepoValidation() *ValidationConfig {
	return &ValidationConfig{
		RequireRepositories: false, // Repo commands can work without existing repos (add, etc.)
		RequireGroups:       false,
		SkipInTesting:       true,
		Commands:            make(map[string]bool),
	}
}

// ValidateCommandContext provides context-aware validation for commands
type ValidateCommandContext struct {
	ConfigMgr   *config.Manager
	Command     *cobra.Command
	Args        []string
	RequireRepo bool
	RequireGroup bool
}

// NewValidateCommandContext creates a new validation context
func NewValidateCommandContext(cmd *cobra.Command, args []string) *ValidateCommandContext {
	return &ValidateCommandContext{
		ConfigMgr: di.ConfigManager(),
		Command:   cmd,
		Args:      args,
	}
}

// WithRepositoryRequired sets repository requirement
func (v *ValidateCommandContext) WithRepositoryRequired() *ValidateCommandContext {
	v.RequireRepo = true
	return v
}

// WithGroupRequired sets group requirement  
func (v *ValidateCommandContext) WithGroupRequired() *ValidateCommandContext {
	v.RequireGroup = true
	return v
}

// Validate performs the configured validation
func (v *ValidateCommandContext) Validate() error {
	if v.RequireRepo {
		if err := ValidateRepositoriesExist(v.ConfigMgr); err != nil {
			return err
		}
	}

	if v.RequireGroup {
		if err := ValidateGroupsExist(v.ConfigMgr); err != nil {
			return err
		}
	}

	return nil
}