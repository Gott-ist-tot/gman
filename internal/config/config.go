package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gman/pkg/types"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Manager handles configuration operations
type Manager struct {
	config     *types.Config
	configPath string
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{}
}

// Load loads the configuration from file
func (m *Manager) Load() error {
	// Set default values
	m.setDefaults()

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create default config
			return m.createDefaultConfig()
		}
		return fmt.Errorf("error reading config file: %w", err)
	}

	// Unmarshal config
	config := &types.Config{}
	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	m.config = config
	return nil
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *types.Config {
	return m.config
}

// AddRepository adds a new repository to the configuration
func (m *Manager) AddRepository(alias, path string) error {
	if m.config.Repositories == nil {
		m.config.Repositories = make(map[string]string)
	}

	// Expand path
	expandedPath, err := expandPath(path)
	if err != nil {
		return fmt.Errorf("error expanding path: %w", err)
	}

	// Check if path exists and is a git repository
	if !isGitRepository(expandedPath) {
		return fmt.Errorf("path %s is not a git repository", expandedPath)
	}

	m.config.Repositories[alias] = expandedPath
	return m.Save()
}

// RemoveRepository removes a repository from the configuration
func (m *Manager) RemoveRepository(alias string) error {
	if m.config.Repositories == nil {
		return fmt.Errorf("no repositories configured")
	}

	if _, exists := m.config.Repositories[alias]; !exists {
		return fmt.Errorf("repository '%s' not found", alias)
	}

	delete(m.config.Repositories, alias)
	return m.Save()
}

// Save saves the current configuration to file
func (m *Manager) Save() error {
	configPath := m.getConfigPath()

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// setDefaults sets default configuration values
func (m *Manager) setDefaults() {
	viper.SetDefault("settings.parallel_jobs", 5)
	viper.SetDefault("settings.show_last_commit", true)
	viper.SetDefault("settings.default_sync_mode", "ff-only")
}

// createDefaultConfig creates a default configuration file
func (m *Manager) createDefaultConfig() error {
	m.config = &types.Config{
		Repositories: make(map[string]string),
		Settings: types.Settings{
			ParallelJobs:    5,
			ShowLastCommit:  true,
			DefaultSyncMode: "ff-only",
		},
	}

	return m.Save()
}

// getConfigPath returns the configuration file path
func (m *Manager) getConfigPath() string {
	if m.configPath != "" {
		return m.configPath
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".config", "gman", "config.yml")
}

// expandPath expands ~ and environment variables in path
func expandPath(path string) (string, error) {
	if path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[2:])
	}

	return os.ExpandEnv(path), nil
}

// isGitRepository checks if the given path is a git repository
func isGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}