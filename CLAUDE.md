# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gman is a **modern, production-ready** Git repository management CLI tool built in Go. It enables developers to efficiently manage multiple Git repositories with features including status checking, interactive switching, batch operations, shell integration, advanced workflow automation, and comprehensive search capabilities.

**Current Status**: **Stable & Feature-Complete** - All core modernization objectives achieved through Optimization Blueprint v3.0.

**Latest Updates**: Command structure cleaned up, duplicate commands removed, help text refreshed, all commands verified working, TUI dashboard functional (requires proper terminal environment).

## Recent Major Enhancements (Optimization Blueprint v3.0)

### P0: Core Stability & Security ✅
- **P0.1: File Locking System** - Implemented concurrent configuration protection using github.com/gofrs/flock
- **P0.2: Git Error Handling** - Enhanced fetch failure detection with SyncError tracking
- **P0.3: Command Injection Security** - Added external tool validation and argument separation

### P1: User Experience & Consistency ✅  
- **P1.1: TUI Search Enhancement** - Added updatePreview method and improved search functionality
- **P1.2: Onboarding Wizard System** - Complete new user setup with repository discovery
- **P1.3: Command Structure Reorganization** - Intuitive command grouping with shortcuts
- **P1.4: TUI Evolution to Interactive Command Center** - Enhanced dashboard with Actions panel

### P2: Long-term Architecture ✅ (Complete)
- **P2.1a: Technical Debt Analysis** - Comprehensive code quality assessment completed
- **P2.1b: Batch Module Decomposition** - 768-line batch.go split into modular cmd/batch/ structure
- **P2.1c: Git Interface Abstraction** - 34-method Manager split into specialized interfaces
- **P2.1d: Dependency Injection Implementation** - Eliminated 93% of manual instantiations

## Optimization Blueprint v3.0 - Completion Summary

### 🎯 **Mission Accomplished**
The gman optimization blueprint has been **successfully completed**, transforming the project from a monolithic CLI tool into a modern, modular, and maintainable codebase.

### 📊 **Achievement Metrics**
- **Security Enhancements**: 100% (File locking, Git error handling, command injection protection)
- **User Experience**: 100% (TUI search, onboarding wizard, command reorganization, interactive dashboard)
- **Technical Debt Resolution**: 100% (Code analysis, modularization, interface abstraction, dependency injection)
- **Code Quality**: Transformed from 768-line monolithic files to modular, interface-driven architecture
- **Dependency Management**: 93% reduction in manual instantiation (from 82 to 6 instances)

### 🏗️ **Architectural Transformation**
**Before**: Monolithic structure with tight coupling and manual dependency management
**After**: Modular, interface-driven design with dependency injection and clear separation of concerns

```
# Structural Evolution
Old: cmd/batch.go (768 lines) → New: cmd/batch/ (4 specialized modules)
Old: git.Manager (34 methods) → New: 6 specialized interfaces
Old: Manual instantiation (82×) → New: DI container (93% reduction)
```

### 🚀 **Developer Experience Improvements**
- **Migration Tooling**: Automated DI migration with `gman migrate-di`
- **Setup Wizard**: Complete onboarding system for new users
- **Command Structure**: Intuitive grouping with shortcuts (repo/r, work/w, quick/q, tools/t)
- **Interactive Dashboard**: Enhanced TUI with real-time operations
- **Testing Coverage**: Comprehensive test suite across all layers

## New Command Structure (P1.3)

gman now features a reorganized command structure with logical grouping and intuitive shortcuts:

### Command Groups
- **`gman repo` (r)** - Repository management (add, remove, list, groups)
- **`gman work` (w)** - Git workflow operations (status, sync, commit, push, branch)
- **`gman quick` (q)** - Quick access to common operations
- **`gman tools` (t)** - Advanced utilities (dashboard, search, worktree, setup)

### Usage Examples
```bash
# Repository management
gman repo add myproject /path/to/project
gman r list                          # Using shortcut
gman repo group create webdev frontend backend

# Git workflow
gman work status --extended
gman w sync --group webdev           # Using shortcut
gman work commit -m "Fix bug" --add

# Quick access
gman quick status                    # No nested structure
gman q switch                        # Direct shortcuts

# Advanced tools
gman tools dashboard
gman t find "config.yaml"            # Using shortcut
gman tools setup discover ~/Projects
```

### Backward Compatibility
All original commands remain functional - users can adopt the new structure gradually without breaking existing workflows.

## Onboarding System (P1.2)

### Setup Wizard
- **`gman setup`** - Interactive 3-step setup wizard for new users
- **Step 1**: Repository discovery with intelligent alias generation
- **Step 2**: Basic configuration (sync mode, parallel jobs, display preferences)
- **Step 3**: Quick tutorial with personalized examples

### Repository Discovery
- **`gman setup discover [path]`** - Automatic Git repository detection
- **`--depth N`** - Control search depth (default: 3 levels)
- **`--auto-confirm`** - Skip interactive selection
- **Smart alias generation** from directory names
- **Duplicate detection** and path normalization

### New User Experience
- **`gman onboarding welcome`** - Contextual welcome message
- **First-run detection** with automatic setup wizard offering
- **Personalized guidance** using actual repository configuration
- **Progressive learning** through contextual tutorials

## Enhanced TUI Dashboard (P1.4)

### Interactive Command Center
The TUI dashboard has evolved from a monitoring interface to a comprehensive interactive command center with enhanced functionality:

### New Panel Layout
The dashboard now features a **2x3 panel layout** providing more workspace and functionality:

```
┌─ Repositories (1) ─┬─ Status (2) ──────┬─ Actions (5) ────┐
│ • Select repos     │ • Detailed status │ • Quick commands │
│ • Filter & search  │ • Branch info     │ • Git operations │
│ • Group management │ • File changes    │ • Interactive    │
└────────────────────┴───────────────────┴──────────────────┘
┌─ Search (3) ────────────────────────────┬─ Preview (4) ────┐
│ • Files & commits across repos         │ • File content   │
│ • Integrated fzf support              │ • Commit details │
│ • Real-time results                    │ • Live updates   │
└────────────────────────────────────────┴──────────────────┘
```

### Actions Panel Features
The new **Actions Panel (5)** provides:

#### Repository Operations
- **Refresh Status** - Update repository information
- **Open in Terminal** - Launch external terminal at repo location
- **Open in File Manager** - Browse repo with system file manager

#### Git Operations (Context-Aware)
- **Sync Repository** - Pull latest changes from remote
- **Commit Changes** - Interactive commit with staging (only shown when dirty)
- **Push Changes** - Push local commits to remote (only shown when ahead)
- **Stash/Pop Changes** - Manage uncommitted work (conditional display)

#### Branch Management
- **Switch Branch** - Interactive branch selection
- **Create Branch** - New branch from current commit
- **Merge Branch** - Merge operations with conflict detection

#### Advanced Operations
- **Create Worktree** - Parallel development workspaces
- **Compare Files** - Diff between branches/repos
- **View Log** - Commit history exploration

### Navigation Enhancements
- **Keyboard shortcuts**: 1-5 for direct panel access
- **Tab/Shift+Tab**: Seamless panel navigation
- **Quick actions**: Single-key operations (r=refresh, s=sync, c=commit, p=push)
- **Context-aware display**: Actions adapt to repository state

### Real-time Feedback
- **Action execution tracking** with progress indicators
- **Result display** with auto-hide after 3 seconds
- **Error handling** with user-friendly messages
- **Non-blocking operations** maintaining dashboard responsiveness

## Security Enhancements (P0.3)

### External Command Security
- **Diff tool validation** with whitelist of allowed tools
- **Command injection prevention** using `--` argument separators
- **Path validation** and character filtering
- **Safe tool execution** with proper argument handling

### Allowed Diff Tools
```
diff, meld, vimdiff, gvimdiff, kdiff3, opendiff, p4merge, 
xxdiff, tkdiff, kompare, emerge, winmerge, code, subl, 
atom, delta, difft
```

## Technical Debt Analysis (P2.1) ✅

### Refactoring Assessment Complete
Comprehensive analysis of the gman codebase identified key improvement opportunities with prioritized implementation roadmap:

### Critical Issues Identified

#### High Priority (High Impact, Medium Effort)
1. **Giant Method Smell** - `cmd/batch.go` contains 768 lines with 7 distinct operations
   - Multiple batch commands (commit, push, stash) in single file
   - Repeated patterns and flag setup across operations
   - Requires decomposition into domain-specific modules

2. **Interface Segregation Violation** - `internal/git/git.go` Manager has 34 methods
   - Single class covering diverse responsibilities (status, branches, worktrees)
   - No interface abstraction limiting testability
   - Needs domain-specific interfaces: `StatusReader`, `BranchManager`, `WorktreeManager`

3. **Repeated Instantiation Anti-Pattern**
   - 45 occurrences of `config.NewManager()` across commands
   - 21 occurrences of `git.NewManager()` across commands
   - Requires dependency injection framework

#### Medium Priority (Medium Impact, Low-Medium Effort)
4. **Function Length Issues** - Multiple functions exceed 160 lines
   - `runSync()` in sync.go: 162 lines with multiple responsibilities
   - Complex nested logic without proper decomposition
   - Requires function extraction and responsibility separation

5. **Missing Abstractions** - No interfaces for Git operations
   - Limits testability and modularity
   - Tight coupling between commands and concrete implementations

#### Low Priority (Various Impact, Low Effort)
6. **TODO Technical Debt** - 13 TODO comments in `actions.go`
7. **Modernization Gap** - Underutilized Go 1.24 features
8. **Error Handling** - Inconsistent patterns across commands
9. **Code Duplication** - Repeated repository filtering logic

### Proposed Refactoring Structure

#### Phase 1: Decomposition
```
cmd/
├── batch/
│   ├── commit.go    # Extracted from batch.go
│   ├── push.go      # Extracted from batch.go  
│   ├── stash.go     # Extracted from batch.go
│   └── common.go    # Shared batch operation logic
└── ...

internal/
├── git/
│   ├── interfaces.go    # Git operation interfaces
│   ├── status.go       # StatusReader implementation
│   ├── branch.go       # BranchManager implementation
│   ├── worktree.go     # WorktreeManager implementation
│   └── manager.go      # Facade combining interfaces
└── di/
    └── container.go    # Dependency injection setup
```

#### Phase 2: Interface Abstraction
```go
// Git operation interfaces
type StatusReader interface {
    GetRepoStatus(alias, path string) types.RepoStatus
    GetAllRepoStatus(repos map[string]string) ([]types.RepoStatus, error)
}

type BranchManager interface {
    GetBranches(path string, includeRemote bool) ([]string, error)
    CreateBranch(path, branchName string) error
    SwitchBranch(path, branchName string) error
}

type GitOperations interface {
    StatusReader
    BranchManager
    WorktreeManager
}
```

### Modernization Opportunities
- **Context Propagation**: Add cancellation support to long operations
- **Generics**: Leverage for repository operation patterns
- **Structured Errors**: Replace string-based error handling
- **Functional Options**: Modern configuration patterns

## Dependency Injection System (P2.1d)

### Overview
The gman project now uses a centralized dependency injection (DI) container to eliminate duplicate manager instantiation and improve consistency across the codebase.

### Architecture
**Container Pattern (internal/di/)**
- `container.go` - Thread-safe singleton container with lazy initialization
- `migration.go` - Analysis and automated migration tooling for DI adoption

### DI Container Features
- **Thread-safe Singleton**: Global container with sync.RWMutex protection
- **Lazy Initialization**: Auto-initialization on first access
- **Lifecycle Tracking**: Usage statistics and initialization timestamps
- **Interface Access**: Direct access to specialized Git operation interfaces

### Usage Patterns
```go
// Before: Manual instantiation (eliminated)
configMgr := config.NewManager()
gitMgr := git.NewManager()

// After: Dependency injection (consistent)
configMgr := di.ConfigManager()
gitMgr := di.GitManager()

// Interface access for specialized operations
statusReader := di.StatusReader()
branchMgr := di.BranchManager()
```

### Migration Results (Final)
- **Files Analyzed**: 67 Go files across entire codebase
- **Migration Success**: 93% reduction in manual instantiations (82 → 6)
- **Remaining Instances**: 6 (only in DI container itself and migration tools - as expected)
- **Consistency Achievement**: All 25 command files consistently use DI container
- **Developer Tooling**: Complete migration analysis and automation tools

### Developer Tools
- **`gman migrate-di`** - Analyze current DI usage and apply automated migration
- **Automatic Import Management** - goimports integration for clean code
- **Migration Verification** - Built-in analysis and reporting tools

### Benefits
- **Consistency**: Uniform dependency access across all commands
- **Testability**: Centralized mock injection for testing
- **Maintainability**: Single point of dependency configuration
- **Performance**: Reduced object creation overhead

### Implementation Status
1. **✅ Completed** (P0): Core Stability & Security - File locking, Git error handling, command injection protection
2. **✅ Completed** (P1): User Experience & Consistency - TUI enhancements, onboarding, command reorganization
3. **✅ Completed** (P2.1): Technical Debt Resolution - Analysis, decomposition, abstraction, dependency injection
4. **🔮 Future** (P2.3): Community ecosystem building - Plugin architecture (optional enhancement)

## Development Commands

### Building and Running
- `make build` - Build the gman binary
- `make run` - Build and run the binary
- `go build -o gman .` - Direct Go build command

### Testing and Quality
- `make test` or `go test ./...` - Run all tests
- `make lint` - Run golangci-lint (requires golangci-lint to be installed)
- `make fmt` - Format code with go fmt

### Development Tools
- `gman migrate-di` - Analyze and migrate dependency injection usage
- `gman migrate-di --dry-run` - Preview DI migration changes
- `gman migrate-di --apply` - Apply automatic DI migration
- `gman setup discover` - Discover and configure Git repositories
- `gman onboarding welcome` - New user guidance system

### TUI Dashboard Usage
- `gman dashboard` - Launch interactive TUI dashboard (requires proper terminal)
- `gman dashboard --debug` - Show terminal compatibility diagnostics
- `gman dashboard --force` - Bypass terminal checks (advanced users)
- `gman dashboard --theme light` - Use light color theme

#### Test Coverage Strategy
The project maintains comprehensive test coverage across multiple layers:

**Unit Tests:**
- `internal/git/git_test.go` - Git operations, diff functionality, worktree management
- `internal/interactive/selector_test.go` - Interactive selection components
- `internal/index/indexer_test.go` - Search indexing system and SQLite operations
- `pkg/types/` - Type definitions and validation

**Command Tests:**
- `cmd/diff_test.go` - File comparison commands (branch diff, cross-repo diff)
- `cmd/worktree_test.go` - Worktree lifecycle management
- `cmd/switch_test.go` - Enhanced switch functionality with worktree integration

**Integration Tests:**
- `test/integration_test.go` - Cross-package functionality
- `test/helpers.go` - Shared test utilities and mock repositories

**Test Categories:**
- **Git Integration Tests** - Real Git operations with temporary repositories
- **Command Line Tests** - End-to-end command execution and output validation
- **Interactive Component Tests** - Simulated user input and menu navigation
- **Error Handling Tests** - Edge cases, invalid inputs, and failure recovery
- **Concurrent Operation Tests** - Multi-repository parallel operations

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
- Each command is in its own file (add.go, list.go, status.go, recent.go, group.go, branch.go, batch.go, find.go, index.go, etc.)
- Root command in `cmd/root.go` handles global configuration and initialization
- Enhanced commands: `recent` for recently accessed repositories, `group` for repository group management
- Advanced Git workflow commands: `branch` for cross-repository branch management, batch operations (`commit`, `push`, `stash`), `diff` for file comparison across branches and repositories, `worktree` for Git worktree management
- **Phase 5.0 Search Commands**: `find` for fzf-powered file and commit searching, `index` for search index management
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
- **File Comparison & Diff Operations**: Advanced file comparison capabilities across branches and repositories
  - `gman diff file <repo> <branch1> <branch2> -- <file_path>` - Compare files between branches within a repository
  - `gman diff cross-repo <repo1> <repo2> -- <file_path>` - Compare files between different repositories
  - `--tool <external_tool>` support for visual diff tools (meld, vimdiff, etc.)
- **Git Worktree Management**: Native Git worktree integration for parallel development
  - `gman worktree add <repo> <path> --branch <branch>` - Create worktrees for parallel feature development
  - `gman worktree list <repo>` - Display all worktrees with branch and status information
  - `gman worktree remove <repo> <path> [--force]` - Clean removal of completed worktrees
  - **Seamless Switch Integration**: Worktrees appear as first-class targets in `gman switch` interactive menu

These enhancements significantly improve batch operation efficiency and provide advanced Git workflow management across multiple repositories.

### Phase 5.0 - 互動體驗重塑：fzf 深度整合 ✅
- **SQLite 索引系統**: 高效的全文搜索索引，支援文件和提交搜索
- **深度 fzf 整合**: 無縫的模糊搜索體驗，跨倉庫文件和提交搜索
- **智能預覽功能**: 即時的文件內容和提交差異預覽
- **索引管理命令**: 完整的索引生命週期管理 (`gman index rebuild/update/stats/clear`)
  - `gman find file [pattern] [--group <name>]` - 跨倉庫文件搜索
  - `gman find commit [pattern] [--group <name>]` - 跨倉庫提交搜索

### Phase 5.2 - TUI Dashboard：統一管理界面 ✅
- **Bubble Tea TUI 框架**: 現代化的終端用戶界面
- **四面板儀表板**: Repository/Status/Search/Preview 統一布局
- **即時狀態監控**: 實時倉庫狀態更新和視覺指示
- **鍵盤導航系統**: 完整的快捷鍵支持和 Vim 風格導航
- **主題系統**: 支援 Dark/Light 主題切換
- **無縫整合**: 與 Phase 5.1 搜索功能和所有現有 CLI 命令完美整合
- **智能終端檢測**: 增強的終端相容性檢測和診斷系統
  - `gman dashboard` - 啟動 TUI 儀表板
  - `gman dash/tui/ui` - 命令別名支持
  - `gman dashboard --debug` - 顯示終端診斷資訊
  - `gman dashboard --force` - 強制啟動 TUI 模式

## Project Status & Future Considerations

### 📈 **Current Status: Production-Ready**

gman has achieved **feature completeness** with a comprehensive multi-repository management solution:

#### ✅ **Core Features (Stable)**
- **Complete Repository Management** - Add, remove, list, group organization
- **Advanced Batch Operations** - Cross-repository commits, pushes, stash management  
- **Deep Git Workflow Integration** - Branch management, worktree support, diff tools
- **Intelligent Search System** - File and commit search with fzf integration
- **Unified TUI Dashboard** - Interactive command center with real-time operations
- **Modern Architecture** - Modular design with dependency injection and interface abstraction

#### 🎯 **Optimization Blueprint v3.0: Complete**
- **✅ P0: Core Stability & Security** - File locking, error handling, command security
- **✅ P1: User Experience & Consistency** - TUI enhancements, onboarding, command structure
- **✅ P2: Technical Debt Resolution** - Modularization, interface abstraction, dependency injection

### 🔮 **Optional Future Enhancements** (Community-Driven)

The following features represent **optional enhancements** that could be implemented based on community feedback and contribution:

#### 🧩 **Plugin Architecture** (P2.3)
- Extensible plugin system for custom commands and integrations
- API for third-party tool integration
- Community contribution framework

#### 🔧 **Advanced Integrations** (Optional)
- GitHub/GitLab API integration for PR/MR status display
- CI/CD pipeline status monitoring
- Issue tracking system integration
- Advanced analytics and reporting

#### 🌟 **Quality of Life** (Optional)
- Enhanced repository discovery and auto-configuration
- Health monitoring and maintenance automation
- Advanced backup and sync capabilities
- Environment-specific repository management

### 💡 **Development Philosophy**

**Current Focus**: **Maintenance and Stability**
- gman is feature-complete for its core mission
- Focus on bug fixes, performance optimization, and documentation
- Community contributions welcome for optional enhancements

**Design Principles Achieved**:
- ✅ **Modular Architecture** - Clean separation of concerns
- ✅ **Interface-Driven Design** - Testable and extensible
- ✅ **User Experience First** - Intuitive commands and workflows
- ✅ **Developer-Friendly** - Comprehensive tooling and documentation

## TUI Dashboard Requirements

### Terminal Compatibility
The interactive TUI dashboard requires a compatible terminal environment:

**✅ Supported Environments:**
- Standard terminal emulators (Terminal.app, iTerm2, GNOME Terminal, etc.)
- SSH sessions with proper TTY allocation (`ssh -t`)
- tmux/screen sessions
- VS Code integrated terminal (when properly configured)

**❌ Unsupported Environments:**
- CI/CD pipelines and automated scripts
- Output redirection (`gman dashboard > file`)
- Non-interactive shells
- Environments without /dev/tty access

**🔧 Troubleshooting:**
```bash
# Check terminal compatibility
gman dashboard --debug

# Force TUI mode (bypass checks)
gman dashboard --force

# Use CLI commands instead
gman status --extended    # Alternative to TUI status
gman switch              # Interactive repository switching
```

## 常見問題與解決方案 (Common Issues and Solutions)

### 🚨 gman switch 無法切換目錄問題

**問題描述**: 執行 `gman switch <repo>` 後看到 `GMAN_CD:/path/to/repo` 輸出，但當前目錄沒有改變。

**技術原理**: 
- Go 程序作為子進程運行，受到操作系統進程隔離機制限制
- 子進程無法修改父 shell 的工作目錄狀態（這是安全設計）
- `os.Chdir()` 只影響 Go 程序本身，不影響調用它的 shell

**解決方案**: 必須安裝 shell 包裝函數來處理 `GMAN_CD:` 輸出

**診斷步驟**:
1. **檢查 gman 是否在 PATH 中**:
   ```bash
   which gman  # 應該顯示 gman 二進制文件路徑
   ```

2. **檢查 shell 函數是否已加載**:
   ```bash
   type gman   # 應該顯示 "gman is a function"
   ```

3. **檢查 shell 配置**:
   ```bash
   grep -n "gman" ~/.zshrc  # 檢查配置是否存在
   ```

**完整配置示例** (添加到 `~/.zshrc` 或 `~/.bashrc`):
```bash
# gman Git Repository Manager - Shell Integration
export PATH="/path/to/gman/directory:$PATH"

# gman wrapper function for directory switching
gman() {
    local output
    local exit_code

    # Call the actual gman binary and capture both output and exit code
    output=$(command gman "$@" 2>&1)
    exit_code=$?

    # Check if this is a directory change request
    if [[ "$output" == GMAN_CD:* ]]; then
        local target_dir="${output#GMAN_CD:}"
        if [ -d "$target_dir" ]; then
            cd "$target_dir"
            echo "Switched to: $target_dir"
        else
            echo "Error: Directory not found: $target_dir" >&2
            return 1
        fi
    else
        # For all other commands, just print the output
        echo "$output"
    fi

    return $exit_code
}

# Enable gman completion if available
if command -v gman &> /dev/null; then
    eval "$(gman completion zsh)"  # 或 bash
fi
```

**測試驗證**:
```bash
# 重新加載配置
source ~/.zshrc

# 測試功能
gman switch <repo-alias>
pwd  # 應該顯示切換後的目錄路徑
```

### 🔧 測試和開發問題

**測試環境配置**:
- 使用 `GMAN_CONFIG` 環境變量指定測試配置文件
- 創建臨時 Git 倉庫進行測試

**調試技巧**:
- 使用 `gman --config /path/to/test/config.yml` 指定配置
- 檢查 `~/.config/gman/config.yml` 文件內容
- 使用 `gman list` 確認倉庫配置正確