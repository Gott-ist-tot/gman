# gman Interactive TUI Dashboard

The TUI (Terminal User Interface) dashboard provides a unified, interactive interface for repository management, search, and Git operations. Built with modern terminal UI principles, it offers a comprehensive command center for multi-repository workflows.

## Overview

The dashboard transforms gman from a command-line tool into an interactive application, providing:

- **Real-time repository monitoring** with live status updates
- **Integrated search capabilities** with instant preview
- **Context-aware operations** that adapt to repository state
- **Keyboard-driven navigation** for efficient workflow
- **Visual Git operations** with progress feedback

## Launching the Dashboard

```bash
# Standard launch
gman tools dashboard

# Alternative commands
gman tools tui
gman tools ui
gman dash

# With specific options
gman tools dashboard --theme light
gman tools dashboard --debug    # Show terminal diagnostics
gman tools dashboard --force    # Bypass terminal checks
```

## Dashboard Layout

### Panel Structure (2x3 Layout)

```
┌─ Repositories (1) ─┬─ Status (2) ──────┬─ Actions (5) ────┐
│ • Repository list  │ • Detailed status │ • Quick commands │
│ • Group filtering  │ • Branch info     │ • Git operations │
│ • Active selection │ • File changes    │ • Context actions│
└────────────────────┴───────────────────┴──────────────────┘
┌─ Search (3) ────────────────────────────┬─ Preview (4) ────┐
│ • File search across repos             │ • File content   │
│ • Content search with regex            │ • Commit details │
│ • Interactive fzf integration          │ • Live updates   │
└────────────────────────────────────────┴──────────────────┘
```

### Panel Functions

#### 1. Repositories Panel
- **Repository List**: All configured repositories with status indicators
- **Group Filtering**: Filter repositories by group membership
- **Selection**: Current active repository with visual highlighting
- **Quick Info**: Branch name and basic status for each repository

#### 2. Status Panel
- **Detailed Status**: Comprehensive repository information
- **Branch Information**: Current branch, upstream tracking
- **File Changes**: Modified, staged, and untracked files
- **Sync Status**: Ahead/behind commit counts

#### 3. Search Panel
- **File Search**: Real-time file discovery across all repositories
- **Content Search**: Regex-powered text search within files
- **Interactive Results**: fzf integration for result navigation
- **Group Scope**: Search within specific repository groups

#### 4. Preview Panel
- **File Content**: Syntax-highlighted file preview
- **Commit Details**: Full commit information and diff
- **Live Updates**: Content updates as selection changes
- **Smart Detection**: Automatic content type recognition

#### 5. Actions Panel
- **Repository Operations**: Refresh, sync, open in terminal/file manager
- **Git Operations**: Commit, push, stash management (context-aware)
- **Branch Management**: Switch, create, merge operations
- **Advanced Features**: Worktree management, file comparison

## Navigation

### Keyboard Shortcuts

#### Panel Navigation
| Key | Action |
|-----|--------|
| `1-5` | Switch to numbered panel directly |
| `Tab` | Next panel |
| `Shift+Tab` | Previous panel |
| `h/j/k/l` | Vim-style panel navigation |

#### List Navigation
| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Move up/down in lists |
| `Page Up/Down` | Page through long lists |
| `Home/End` | Jump to start/end of list |
| `Enter` | Select item/execute action |

#### Search and Filter
| Key | Action |
|-----|--------|
| `/` | Start search in current panel |
| `Ctrl+F` | Global search mode |
| `Esc` | Cancel search/go back |
| `Ctrl+G` | Toggle group filter |

#### Quick Actions
| Key | Action |
|-----|--------|
| `r` | Refresh repository status |
| `s` | Sync selected repository |
| `c` | Commit changes (if dirty) |
| `p` | Push changes (if ahead) |
| `o` | Open in file manager |
| `t` | Open terminal at repository |

#### General
| Key | Action |
|-----|--------|
| `q` | Quit dashboard |
| `?` | Show help/keyboard shortcuts |
| `Ctrl+C` | Force quit |

## Features

### Real-Time Status Monitoring

The dashboard continuously monitors repository state and updates displays:

- **Git Status Changes**: Automatic detection of file modifications
- **Branch Updates**: Real-time branch and remote tracking information
- **Visual Indicators**: Color-coded status for quick assessment
- **Background Refresh**: Non-blocking status updates every 30 seconds

### Context-Aware Actions

Actions panel adapts based on repository state:

#### Clean Repository
- Refresh Status
- Open in Terminal/File Manager
- Switch Branch
- Create Worktree

#### Dirty Repository (Uncommitted Changes)
- Refresh Status
- **Commit Changes** (highlighted)
- Stash Changes
- View Diff

#### Ahead Repository (Unpushed Commits)
- Refresh Status
- **Push Changes** (highlighted)
- View Log
- Rebase Interactive

#### Behind Repository (Remote Updates)
- Refresh Status
- **Sync Repository** (highlighted)
- Force Pull
- View Incoming Changes

### Integrated Search System

#### File Search
- **Real-time Discovery**: Uses `fd` for instant file search
- **Pattern Matching**: Support for glob patterns and regex
- **Cross-Repository**: Search across all configured repositories
- **Group Filtering**: Limit search to specific repository groups

#### Content Search
- **Regex Support**: Full regex pattern matching with `ripgrep`
- **Context Display**: Show surrounding lines for better understanding
- **File Type Filtering**: Search specific file types
- **Exclude Patterns**: Respect .gitignore and custom excludes

#### Interactive Selection
- **fzf Integration**: Smooth fuzzy finding interface
- **Preview Pane**: Live preview of file content or search results
- **Multi-Select**: Select multiple files for batch operations
- **History**: Remember recent searches

### Visual Git Operations

#### Commit Workflow
1. **Detection**: Dashboard shows repositories with uncommitted changes
2. **Staging**: Interactive file selection for staging
3. **Message**: Built-in commit message editor
4. **Execution**: Real-time commit with progress display
5. **Feedback**: Immediate status update and confirmation

#### Sync Operations
1. **Analysis**: Check remote status and conflicts
2. **Preview**: Show what will be pulled/pushed
3. **Execution**: Progress display for fetch/merge operations
4. **Resolution**: Guided conflict resolution if needed
5. **Completion**: Updated status with sync results

## Terminal Compatibility

### Supported Environments
- **Terminal Emulators**: iTerm2, Terminal.app, GNOME Terminal, Konsole
- **Terminal Multiplexers**: tmux, screen with proper TERM settings
- **SSH Sessions**: With TTY allocation (`ssh -t user@host`)
- **IDE Terminals**: VS Code, JetBrains when properly configured

### Requirements
- **TTY Support**: Must be connected to a terminal device
- **Color Support**: ANSI color codes (256-color recommended)
- **Minimum Size**: 80x24 characters (120x30 recommended)
- **Unicode Support**: For visual indicators and icons

### Unsupported Environments
- **CI/CD Pipelines**: Automated environments without TTY
- **Output Redirection**: When stdout is redirected to file
- **Non-Interactive Shells**: Scripts and automation contexts
- **Limited Terminals**: TERM=dumb or very basic terminals

### Troubleshooting Terminal Issues

```bash
# Check terminal compatibility
gman tools dashboard --debug

# Common issues and solutions:

# Issue: "stdout is not connected to a TTY"
# Solution: Use proper SSH connection
ssh -t user@host

# Issue: Dashboard doesn't display correctly
# Solution: Check TERM environment variable
echo $TERM  # Should be xterm-256color or similar

# Issue: Colors not working
# Solution: Force color support
export FORCE_COLOR=1
gman tools dashboard

# Issue: Size too small
# Solution: Resize terminal window to at least 80x24
```

## Configuration

### Theme Options

```bash
# Dark theme (default)
gman tools dashboard --theme dark

# Light theme
gman tools dashboard --theme light
```

### Keyboard Customization

Create `~/.config/gman/dashboard.yml` for custom shortcuts:

```yaml
keybindings:
  quit: "q"
  refresh: "r"
  search: "/"
  help: "?"
  
panels:
  navigation: "12345"
  tab_next: "Tab"
  tab_prev: "Shift+Tab"
  
actions:
  sync: "s"
  commit: "c"
  push: "p"
  open_terminal: "t"
  open_file_manager: "o"
```

### Display Settings

```yaml
display:
  auto_refresh: 30s
  show_hidden_files: false
  preview_lines: 20
  max_commits: 50
  
search:
  default_tool: "fd"
  preview_enabled: true
  context_lines: 3
```

## Tips and Best Practices

### Efficient Workflow
1. **Use Number Keys**: Jump directly to panels with 1-5
2. **Learn Quick Actions**: Memorize `r`, `s`, `c`, `p` for common operations
3. **Group Filtering**: Use `Ctrl+G` to focus on specific project groups
4. **Search Shortcuts**: Use `/` for quick panel searches

### Performance
1. **Monitor Large Repositories**: Dashboard may be slower with very large repos
2. **Limit Auto-Refresh**: Adjust auto-refresh interval for better performance
3. **Use Groups**: Filter to relevant repositories to reduce overhead
4. **Terminal Size**: Larger terminals provide better experience

### Integration
1. **External Tools**: Dashboard integrates with configured diff tools
2. **Shell Commands**: Use `t` key to open terminal at repository location
3. **File Manager**: Use `o` key to browse repository in file manager
4. **Search Results**: Selected files can be opened in configured editor

## Advanced Features

### Worktree Management
- **Create Worktrees**: Direct worktree creation from dashboard
- **Switch Context**: Navigate between main repo and worktrees
- **Status Tracking**: Monitor worktree status independently

### Branch Operations
- **Visual Branch List**: See all branches with tracking information
- **Interactive Switching**: Switch branches with preview
- **Merge Management**: Guided merge operations with conflict detection

### File Comparison
- **Cross-Repository Diff**: Compare files between different repositories
- **Branch Comparison**: Compare files between branches
- **External Tools**: Integration with visual diff tools

## Limitations

### Current Limitations
- **Single User**: Designed for single-user repository management
- **Git Only**: Limited to Git repositories (no SVN, Mercurial support)
- **Terminal Bound**: Requires compatible terminal environment
- **Resource Usage**: Can be memory intensive with many large repositories

### Future Enhancements
- **Multi-User Support**: Shared repository configurations
- **Plugin System**: Extensible functionality
- **Remote Repositories**: Support for remote repository management
- **Performance Optimization**: Improved handling of large repository sets

For more information about terminal compatibility and troubleshooting, see [Terminal Compatibility Guide](../troubleshooting/TERMINAL_COMPATIBILITY.md).