package repository

import (
	"fmt"

	"gman/internal/config"
)

// Filter provides repository filtering functionality
// Consolidates group filtering logic used across search commands and other operations
type Filter struct {
	configMgr *config.Manager
}

// NewFilter creates a new repository filter instance
func NewFilter(configMgr *config.Manager) *Filter {
	return &Filter{
		configMgr: configMgr,
	}
}

// FilterByGroup returns repositories filtered by group name
// Consolidates the repeated group filtering pattern from external searchers
func (f *Filter) FilterByGroup(repositories map[string]string, groupFilter string) (map[string]string, error) {
	if groupFilter == "" {
		return repositories, nil
	}

	groupRepos, err := f.configMgr.GetGroupRepositories(groupFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get group repositories: %w", err)
	}

	if len(groupRepos) == 0 {
		return nil, fmt.Errorf("group '%s' is empty or does not exist", groupFilter)
	}

	return groupRepos, nil
}

// FilterByGroupWithValidation filters repositories and validates they exist
// Extended version that ensures filtered repositories are actually configured
func (f *Filter) FilterByGroupWithValidation(repositories map[string]string, groupFilter string) (map[string]string, error) {
	filteredRepos, err := f.FilterByGroup(repositories, groupFilter)
	if err != nil {
		return nil, err
	}

	// Validate that all group repositories exist in the main repository list
	for alias, path := range filteredRepos {
		if mainPath, exists := repositories[alias]; !exists {
			return nil, fmt.Errorf("group repository '%s' not found in main repository list", alias)
		} else if mainPath != path {
			return nil, fmt.Errorf("group repository '%s' path mismatch: group has '%s', main has '%s'", alias, path, mainPath)
		}
	}

	return filteredRepos, nil
}

// GetGroupNames returns all available group names
func (f *Filter) GetGroupNames() []string {
	cfg := f.configMgr.GetConfig()
	var names []string
	for name := range cfg.Groups {
		names = append(names, name)
	}
	return names
}

// ValidateGroupExists checks if a group exists
func (f *Filter) ValidateGroupExists(groupName string) error {
	cfg := f.configMgr.GetConfig()
	if _, exists := cfg.Groups[groupName]; !exists {
		return fmt.Errorf("group '%s' does not exist", groupName)
	}
	return nil
}

// GetRepositoryCount returns the number of repositories in a group
func (f *Filter) GetRepositoryCount(groupName string) (int, error) {
	if err := f.ValidateGroupExists(groupName); err != nil {
		return 0, err
	}

	groupRepos, err := f.configMgr.GetGroupRepositories(groupName)
	if err != nil {
		return 0, fmt.Errorf("failed to get group repositories: %w", err)
	}

	return len(groupRepos), nil
}

// FilterInfo provides information about filtering results
type FilterInfo struct {
	OriginalCount int
	FilteredCount int
	GroupName     string
	Applied       bool
}

// FilterWithInfo returns filtered repositories along with filtering information
func (f *Filter) FilterWithInfo(repositories map[string]string, groupFilter string) (map[string]string, *FilterInfo, error) {
	info := &FilterInfo{
		OriginalCount: len(repositories),
		GroupName:     groupFilter,
		Applied:       groupFilter != "",
	}

	filteredRepos, err := f.FilterByGroup(repositories, groupFilter)
	if err != nil {
		return nil, info, err
	}

	info.FilteredCount = len(filteredRepos)
	return filteredRepos, info, nil
}