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
- Each command is in its own file (add.go, list.go, status.go, recent.go, group.go, branch.go, batch.go, etc.)
- Root command in `cmd/root.go` handles global configuration and initialization
- Enhanced commands: `recent` for recently accessed repositories, `group` for repository group management
- Advanced Git workflow commands: `branch` for cross-repository branch management, batch operations (`commit`, `push`, `stash`)
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
- Cross-repository branch management: GetBranches(), CreateBranch(), SwitchBranch(), CleanMergedBranches()
- Batch Git operations: CommitChanges(), PushChanges(), StashSave(), StashPop(), StashList(), StashClear()
- Utility methods: HasUncommittedChanges(), HasUnpushedCommits(), detectMainBranch()

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

### Phase 3.1 - Git 工作流程深度整合 ✅
- **Cross-Repository Branch Management**: Complete branch operations across all repositories
  - `gman branch list [--verbose] [--remote]` - Display branch status across repositories
  - `gman branch create <name>` - Create branches in multiple repositories
  - `gman branch switch <name>` - Switch branches across repositories
  - `gman branch clean [--main <branch>]` - Clean merged branches automatically
- **Batch Git Operations**: Unified Git operations with group support and progress tracking
  - `gman commit -m "message" [--add] [--group <name>]` - Cross-repository commits
  - `gman push [--force] [--set-upstream] [--group <name>]` - Batch push operations
  - `gman stash [save|pop|list|clear] [--group <name>]` - Cross-repository stash management
- **Enhanced Group Integration**: All new commands support group filtering and dry-run modes

These enhancements significantly improve batch operation efficiency and provide advanced Git workflow management across multiple repositories.

## Future Roadmap - 未來功能規劃

### 🔧 **智能倉庫管理** (未來考慮)
- `gman discover <path>` - 自動發現並添加指定路徑下的 Git 倉庫
- `gman clone <url> [alias]` - 克隆遠程倉庫並自動添加到配置
- `gman import <config-file>` - 從其他工具或格式導入倉庫配置

### 🔧 **搜索與分析功能** (未來考慮)
- `gman search <pattern> [--type file|content]` - 跨倉庫內容和文件搜索
- `gman find <filename>` - 跨倉庫文件快速查找
- `gman analytics [--group <name>]` - 倉庫活動分析、提交統計、活躍度報告

### 🔧 **倉庫健康與維護** (未來考慮)
- `gman health [--detailed]` - 檢查倉庫狀態、大文件、潛在問題
- `gman cleanup [--aggressive]` - 自動清理和優化倉庫 (git gc, 清理分支等)
- `gman backup <destination>` - 批量備份倉庫到指定位置

### 🔧 **環境管理系統** (未來考慮)
- `gman env create <name>` - 創建環境配置 (dev/staging/prod)
- `gman env switch <name>` - 切換到指定環境的倉庫集合
- `gman env list` - 列出所有環境配置
- `gman env sync <name>` - 同步特定環境的所有倉庫

### 💡 **高級整合功能** (未來考慮)
- GitHub/GitLab API 整合：顯示 Pull Request、Merge Request 狀態
- CI/CD 管道狀態監控和顯示
- Issue tracking 系統整合
- 配置管理增強 (`export/import`)
- 項目模板系統和插件架構

### 📈 **項目狀態**

gman 現已提供完整的多倉庫管理功能，包括：
- **完整的倉庫狀態管理** (Phase 1)
- **高級批量操作和群組管理** (Phase 2)  
- **深度 Git 工作流程整合** (Phase 3.1)

這些功能足以滿足大多數多倉庫開發場景的需求。未來功能將根據用戶反饋和實際需求進行規劃。