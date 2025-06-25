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
- **智能搜索和索引系統** (Phase 5.0)
- **統一 TUI 管理界面** (Phase 5.2)

這些功能足以滿足大多數多倉庫開發場景的需求。未來功能將根據用戶反饋和實際需求進行規劃。

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