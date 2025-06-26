package di

import (
	"sync"
	"time"

	"gman/internal/config"
	"gman/internal/git"
)

// Container holds all application dependencies
type Container struct {
	configManager *config.Manager
	gitManager    *git.Manager
	gitFacade     *git.GitManager
	mu            sync.RWMutex
	initialized   bool
	// Lifecycle tracking
	initialized_at int64
	access_count   int64
}

// singleton container instance
var (
	container *Container
	once      sync.Once
)

// GetContainer returns the singleton container instance
func GetContainer() *Container {
	once.Do(func() {
		container = &Container{}
	})
	return container
}

// Reset resets the singleton container (for testing)
func Reset() {
	container = nil
	once = sync.Once{}
}

// Initialize sets up all dependencies
func (c *Container) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return nil
	}

	// Initialize config manager
	c.configManager = config.NewManager()

	// Initialize git manager
	c.gitManager = git.NewManager()

	// Initialize git facade with interfaces
	c.gitFacade = git.NewGitManager()

	c.initialized = true
	c.initialized_at = time.Now().Unix()
	return nil
}

// GetConfigManager returns the config manager instance
func (c *Container) GetConfigManager() *config.Manager {
	c.mu.RLock()
	if c.initialized {
		c.access_count++
		configMgr := c.configManager
		c.mu.RUnlock()
		return configMgr
	}
	c.mu.RUnlock()

	// Auto-initialize if not done (outside of read lock to avoid deadlock)
	c.Initialize()
	
	c.mu.RLock()
	c.access_count++
	configMgr := c.configManager
	c.mu.RUnlock()
	return configMgr
}

// GetGitManager returns the git manager instance
func (c *Container) GetGitManager() *git.Manager {
	c.mu.RLock()
	if c.initialized {
		c.access_count++
		gitMgr := c.gitManager
		c.mu.RUnlock()
		return gitMgr
	}
	c.mu.RUnlock()

	// Auto-initialize if not done (outside of read lock to avoid deadlock)
	c.Initialize()
	
	c.mu.RLock()
	c.access_count++
	gitMgr := c.gitManager
	c.mu.RUnlock()
	return gitMgr
}

// GetGitFacade returns the git facade instance
func (c *Container) GetGitFacade() *git.GitManager {
	c.mu.RLock()
	if c.initialized {
		c.access_count++
		facade := c.gitFacade
		c.mu.RUnlock()
		return facade
	}
	c.mu.RUnlock()

	// Auto-initialize if not done (outside of read lock to avoid deadlock)
	c.Initialize()
	
	c.mu.RLock()
	c.access_count++
	facade := c.gitFacade
	c.mu.RUnlock()
	return facade
}

// GetStatusReader returns a StatusReader interface
func (c *Container) GetStatusReader() git.StatusReader {
	return c.GetGitFacade().GetStatusReader()
}

// GetBranchManager returns a BranchManager interface
func (c *Container) GetBranchManager() git.BranchManager {
	return c.GetGitFacade().GetBranchManager()
}

// GetSyncManager returns a SyncManager interface
func (c *Container) GetSyncManager() git.SyncManager {
	return c.GetGitFacade().GetSyncManager()
}

// GetCommitManager returns a CommitManager interface
func (c *Container) GetCommitManager() git.CommitManager {
	return c.GetGitFacade().GetCommitManager()
}

// GetWorktreeManager returns a WorktreeManager interface
func (c *Container) GetWorktreeManager() git.WorktreeManager {
	return c.GetGitFacade().GetWorktreeManager()
}

// GetDiffProvider returns a DiffProvider interface
func (c *Container) GetDiffProvider() git.DiffProvider {
	return c.GetGitFacade().GetDiffProvider()
}

// Reset clears the container for testing purposes
func (c *Container) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.configManager = nil
	c.gitManager = nil
	c.gitFacade = nil
	c.initialized = false
	c.initialized_at = 0
	c.access_count = 0
}

// Stats returns container usage statistics
func (c *Container) Stats() (initialized bool, initializedAt int64, accessCount int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.initialized, c.initialized_at, c.access_count
}

// MustInitialize initializes the container and panics on error
func (c *Container) MustInitialize() {
	if err := c.Initialize(); err != nil {
		panic("failed to initialize DI container: " + err.Error())
	}
}

// Convenience functions for quick access

// ConfigManager returns the config manager instance
func ConfigManager() *config.Manager {
	return GetContainer().GetConfigManager()
}

// GitManager returns the git manager instance
func GitManager() *git.Manager {
	return GetContainer().GetGitManager()
}

// GitFacade returns the git facade instance
func GitFacade() *git.GitManager {
	return GetContainer().GetGitFacade()
}

// StatusReader returns a StatusReader interface
func StatusReader() git.StatusReader {
	return GetContainer().GetStatusReader()
}

// BranchManager returns a BranchManager interface
func BranchManager() git.BranchManager {
	return GetContainer().GetBranchManager()
}

// SyncManager returns a SyncManager interface
func SyncManager() git.SyncManager {
	return GetContainer().GetSyncManager()
}

// CommitManager returns a CommitManager interface
func CommitManager() git.CommitManager {
	return GetContainer().GetCommitManager()
}

// WorktreeManager returns a WorktreeManager interface
func WorktreeManager() git.WorktreeManager {
	return GetContainer().GetWorktreeManager()
}

// DiffProvider returns a DiffProvider interface
func DiffProvider() git.DiffProvider {
	return GetContainer().GetDiffProvider()
}
