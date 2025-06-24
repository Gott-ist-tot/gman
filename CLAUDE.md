# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gman is a Git repository management CLI tool built in Go. It allows developers to manage multiple Git repositories efficiently with features like status checking, quick switching, batch operations, and shell integration.

## Development Commands

### Building and Running
- `make build` - Build the gman binary
- `make run` - Build and run the binary
- `go build -o gman .` - Direct Go build command

### Testing and Quality
- `make test` or `go test ./...` - Run all tests
- `make lint` - Run golangci-lint (requires golangci-lint to be installed)
- `make fmt` - Format code with go fmt

### Cross-platform Building
- `make build-all` - Build for multiple platforms (Linux, macOS, Windows)
- `make clean` - Clean build artifacts

### Installation
- `make install` - Install binary to /usr/local/bin (requires sudo)
- `./scripts/install.sh` - Installation script for end users

## Architecture

### Core Components

**Command Structure (cmd/)**
- Uses Cobra CLI framework for command handling
- Each command is in its own file (add.go, list.go, status.go, recent.go, group.go, etc.)
- Root command in `cmd/root.go` handles global configuration and initialization
- Enhanced commands: `recent` for recently accessed repositories, `group` for repository group management
- Extended sync command with conditional options, dry-run mode, and progress display

**Configuration Management (internal/config/)**
- `config.go` handles YAML-based configuration at `~/.config/gman/config.yml`
- Manages repository mappings, user settings, recent usage tracking, and repository groups
- Direct YAML parsing for full feature support (recent usage, groups, extended settings)
- Methods: TrackRecentUsage(), CreateGroup(), DeleteGroup(), GetGroupRepositories(), AddToGroup(), RemoveFromGroup()

**Interactive Package (internal/interactive/)**
- `selector.go` provides interactive repository selection
- Fuzzy matching capabilities for repository aliases
- User-friendly numbered selection interface

**Git Operations (internal/git/)**
- Handles all Git-specific operations like status checking and sync
- Supports concurrent operations across multiple repositories with semaphore-based concurrency control
- Extended status information: file change counts, commit timestamps
- Repository filtering for conditional sync operations
- New methods: getFilesChangedCount(), getLastCommitTime()

**Progress Tracking (internal/progress/)**
- `progress.go` provides real-time progress tracking for concurrent operations
- MultiBar system for tracking multiple repository operations simultaneously
- Individual progress bars with ETA calculations and duration formatting
- OperationStatus tracking for pending, running, completed, and failed operations

**Display Layer (internal/display/)**
- Manages colorized table output with flexible column layouts
- Formats repository status with visual indicators (🟢🔴🟡 for workspace status, ✅↑↓🔄 for sync status)
- Extended display mode showing file changes and commit times
- NewExtendedStatusDisplayer() for detailed information display

**Shared Types (pkg/types/)**
- Core data structures: Config, RepoStatus, WorkspaceStatus, SyncStatus, RecentEntry, Group
- Extended RepoStatus with FilesChanged and CommitTime fields
- RecentEntry for tracking repository access history with timestamps
- Group type for organizing repositories with metadata (name, description, creation time)
- Enums with String() methods for display formatting

### Key Features Implementation

**Shell Integration**: The tool outputs special `GMAN_CD:` prefix for directory changes that the shell wrapper function processes to actually change directories.

**Interactive Commands**: 
- `gman switch` without arguments shows an interactive selection menu
- Fuzzy matching allows partial repository name matching (e.g., `gman switch cli` matches `cli-tool`)
- Recent usage tracking automatically records and displays recently accessed repositories

**Enhanced Status Display**:
- Standard mode: `gman status` shows basic repository information
- Extended mode: `gman status --extended` includes file change counts and commit timestamps
- Human-readable time formatting (19m, 2h, 3d, etc.)

**Recent Usage Tracking**:
- `gman recent` command shows recently accessed repositories with timestamps
- Automatic tracking when switching between repositories
- Maintains last 10 accessed repositories with access times

**Batch Operations & Conditional Sync**:
- Conditional sync flags: `--only-dirty`, `--only-behind`, `--only-ahead` for selective repository synchronization
- `--dry-run` mode for previewing sync operations without execution
- `--progress` flag for real-time progress tracking with concurrent operation status
- Repository filtering logic based on workspace and sync status

**Repository Grouping**:
- `gman group create <name> <repos...>` - Create repository groups with optional descriptions
- `gman group list` - Display all configured groups with metadata and repository counts
- `gman group delete <name>` - Remove repository groups
- `gman group add/remove <name> <repos...>` - Modify group membership
- `gman sync --group <name>` - Sync only repositories in specific groups
- YAML configuration storage with timestamps and descriptions

**Concurrent Operations**: Uses Go's concurrency with semaphore-based parallelism for repository operations (default 5 parallel jobs).

**Configuration**: YAML-based with defaults, supports repository aliases, custom sync modes (ff-only, rebase, autostash), recent usage history, and repository groups.

## Dependencies

Key external libraries:
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management  
- `github.com/fatih/color` - Colorized output
- `github.com/olekukonko/tablewriter` - Table formatting
- `gopkg.in/yaml.v3` - YAML parsing

## Testing Strategy

The project uses Go's built-in testing framework. Run tests before submitting changes to ensure functionality remains intact across the concurrent Git operations and configuration management.

## Recent Enhancements

All Phase 1 and Phase 2 features have been successfully implemented:

### Phase 1 - Interactive Experience ✅
- **Interactive Repository Selection**: Enhanced `gman switch` command with fuzzy matching and interactive selection
- **Recent Usage Tracking**: Automatic tracking of repository access with `gman recent` command
- **Enhanced Status Display**: Extended status mode with file change counts and commit timestamps

### Phase 2 - Batch Operations ✅  
- **Conditional Sync Options**: Implemented selective synchronization flags (`--only-dirty`, `--only-behind`, `--only-ahead`)
- **Dry-Run Mode**: Added preview functionality with `--dry-run` flag for safe operation verification
- **Progress Display**: Real-time progress tracking with `--progress` flag and MultiBar system
- **Repository Grouping**: Complete group management system with create, list, delete, add, and remove operations

These enhancements significantly improve batch operation efficiency and provide better control over multi-repository workflows.