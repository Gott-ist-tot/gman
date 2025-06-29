package cmd

import (
	"fmt"

	"gman/internal/config"
	"gman/internal/di"
	"gman/internal/git"
)

// Common utility functions for command implementations
// This package consolidates frequently used patterns across commands

// Managers provides commonly accessed dependency injection instances
type Managers struct {
	Config *config.Manager
	Git    *git.Manager
}

// GetManagers returns the standard set of managers used by most commands
// Replaces the repetitive pattern: configMgr := di.ConfigManager()
func GetManagers() *Managers {
	return &Managers{
		Config: di.ConfigManager(),
		Git:    di.GitManager(),
	}
}

// GetManagersWithConfig returns managers and loads configuration
// Consolidates the common pattern of getting managers and loading config
func GetManagersWithConfig() (*Managers, error) {
	mgrs := GetManagers()
	if err := mgrs.Config.Load(); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return mgrs, nil
}

// ValidateRepositoriesExist checks if any repositories are configured
// Consolidates the common validation pattern used in command groups
func ValidateRepositoriesExist(configMgr *config.Manager) error {
	cfg := configMgr.GetConfig()
	if len(cfg.Repositories) == 0 {
		return fmt.Errorf("no repositories configured. Use 'gman repo add' to add repositories")
	}
	return nil
}

// ValidateRepositoriesWithGroups checks repositories and loads groups
// Extends basic validation to include group information
func ValidateRepositoriesWithGroups(configMgr *config.Manager) error {
	if err := ValidateRepositoriesExist(configMgr); err != nil {
		return err
	}
	
	cfg := configMgr.GetConfig()
	if len(cfg.Groups) == 0 {
		// Groups are optional, just log but don't error
		// This allows commands to work without groups configured
	}
	
	return nil
}

// FormatOperationError creates standardized error messages for operations
// Consolidates the repetitive fmt.Errorf("failed to %s: %w", operation, err) pattern
func FormatOperationError(operation string, err error) error {
	return fmt.Errorf("failed to %s: %w", operation, err)
}

// FormatValidationError creates standardized validation error messages
func FormatValidationError(field, value, reason string) error {
	return fmt.Errorf("invalid %s '%s': %s", field, value, reason)
}

// FormatNotFoundError creates standardized "not found" error messages
func FormatNotFoundError(resource, identifier string) error {
	return fmt.Errorf("%s '%s' not found", resource, identifier)
}