package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"gman/pkg/types"
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
	configPath := m.getConfigPath()
	
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file not found, create default config
		return m.createDefaultConfig()
	}

	// Read config file directly
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	// Unmarshal YAML directly
	config := &types.Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Set defaults if not present
	if config.Settings.ParallelJobs == 0 {
		config.Settings.ParallelJobs = 5
	}
	if config.Settings.DefaultSyncMode == "" {
		config.Settings.DefaultSyncMode = "ff-only"
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

// TrackRecentUsage adds or updates a repository in the recent usage list
func (m *Manager) TrackRecentUsage(alias string) error {
	if m.config.RecentUsage == nil {
		m.config.RecentUsage = []types.RecentEntry{}
	}

	// Remove existing entry if present
	for i, entry := range m.config.RecentUsage {
		if entry.Alias == alias {
			m.config.RecentUsage = append(m.config.RecentUsage[:i], m.config.RecentUsage[i+1:]...)
			break
		}
	}

	// Add to the beginning of the list
	newEntry := types.RecentEntry{
		Alias:      alias,
		AccessTime: time.Now(),
	}
	m.config.RecentUsage = append([]types.RecentEntry{newEntry}, m.config.RecentUsage...)

	// Keep only the last 10 entries
	if len(m.config.RecentUsage) > 10 {
		m.config.RecentUsage = m.config.RecentUsage[:10]
	}

	return m.Save()
}

// GetRecentUsage returns the recent usage list
func (m *Manager) GetRecentUsage() []types.RecentEntry {
	if m.config.RecentUsage == nil {
		return []types.RecentEntry{}
	}
	return m.config.RecentUsage
}

// CreateGroup creates a new repository group
func (m *Manager) CreateGroup(name, description string, repositories []string) error {
	if m.config.Groups == nil {
		m.config.Groups = make(map[string]types.Group)
	}

	// Check if group already exists
	if _, exists := m.config.Groups[name]; exists {
		return fmt.Errorf("group '%s' already exists", name)
	}

	// Validate repositories exist
	for _, repo := range repositories {
		if _, exists := m.config.Repositories[repo]; !exists {
			return fmt.Errorf("repository '%s' not found", repo)
		}
	}

	// Create group
	group := types.Group{
		Name:         name,
		Description:  description,
		Repositories: repositories,
		CreatedAt:    time.Now(),
	}

	m.config.Groups[name] = group
	return m.Save()
}

// DeleteGroup removes a group
func (m *Manager) DeleteGroup(name string) error {
	if m.config.Groups == nil {
		return fmt.Errorf("no groups configured")
	}

	if _, exists := m.config.Groups[name]; !exists {
		return fmt.Errorf("group '%s' not found", name)
	}

	delete(m.config.Groups, name)
	return m.Save()
}

// GetGroups returns all configured groups
func (m *Manager) GetGroups() map[string]types.Group {
	if m.config.Groups == nil {
		return make(map[string]types.Group)
	}
	return m.config.Groups
}

// GetGroupRepositories returns repositories for a specific group
func (m *Manager) GetGroupRepositories(groupName string) (map[string]string, error) {
	if m.config.Groups == nil {
		return nil, fmt.Errorf("no groups configured")
	}

	group, exists := m.config.Groups[groupName]
	if !exists {
		return nil, fmt.Errorf("group '%s' not found", groupName)
	}

	result := make(map[string]string)
	for _, alias := range group.Repositories {
		if path, exists := m.config.Repositories[alias]; exists {
			result[alias] = path
		}
	}

	return result, nil
}

// AddToGroup adds repositories to an existing group
func (m *Manager) AddToGroup(groupName string, repositories []string) error {
	if m.config.Groups == nil {
		return fmt.Errorf("no groups configured")
	}

	group, exists := m.config.Groups[groupName]
	if !exists {
		return fmt.Errorf("group '%s' not found", groupName)
	}

	// Validate repositories exist
	for _, repo := range repositories {
		if _, exists := m.config.Repositories[repo]; !exists {
			return fmt.Errorf("repository '%s' not found", repo)
		}
	}

	// Add repositories (avoiding duplicates)
	for _, repo := range repositories {
		found := false
		for _, existing := range group.Repositories {
			if existing == repo {
				found = true
				break
			}
		}
		if !found {
			group.Repositories = append(group.Repositories, repo)
		}
	}

	m.config.Groups[groupName] = group
	return m.Save()
}

// RemoveFromGroup removes repositories from a group
func (m *Manager) RemoveFromGroup(groupName string, repositories []string) error {
	if m.config.Groups == nil {
		return fmt.Errorf("no groups configured")
	}

	group, exists := m.config.Groups[groupName]
	if !exists {
		return fmt.Errorf("group '%s' not found", groupName)
	}

	// Remove repositories
	for _, repo := range repositories {
		for i, existing := range group.Repositories {
			if existing == repo {
				group.Repositories = append(group.Repositories[:i], group.Repositories[i+1:]...)
				break
			}
		}
	}

	m.config.Groups[groupName] = group
	return m.Save()
}
