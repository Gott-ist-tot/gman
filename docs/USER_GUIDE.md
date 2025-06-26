# gman User Guide

A comprehensive guide to using gman (Git Repository Manager) for efficient multi-repository development workflows.

## Table of Contents

- [Getting Started](#getting-started)
- [Basic Concepts](#basic-concepts)
- [Command Structure](#command-structure)
- [Repository Management](#repository-management)
- [Git Workflow Operations](#git-workflow-operations)
- [Enhanced Search System](#enhanced-search-system)
- [Interactive Dashboard](#interactive-dashboard)
- [Repository Groups](#repository-groups)
- [Configuration](#configuration)
- [Advanced Features](#advanced-features)
- [Tips & Best Practices](#tips--best-practices)

## Getting Started

### First Time Setup

1. **Install gman** (see [DEPLOYMENT.md](../DEPLOYMENT.md) for detailed instructions):
   ```bash
   git clone <repository-url>
   cd gman
   go build -o gman .
   ./scripts/install.sh
   ```

2. **Run the setup wizard**:
   ```bash
   gman tools setup
   ```
   This interactive wizard will:
   - Discover existing Git repositories on your system
   - Configure basic settings
   - Create your first repository groups
   - Set up shell integration

3. **Verify installation**:
   ```bash
   gman --version
   gman repo list
   ```

### Alternative Manual Setup

If you prefer manual configuration:

```bash
# Add your first repository
gman repo add /path/to/project my-project

# Check status
gman work status

# Test directory switching
gman switch my-project
```

## Basic Concepts

### Repositories
Individual Git repositories managed by gman. Each repository has:
- **Alias**: Short, memorable name for the repository
- **Path**: Full filesystem path to the repository
- **Status**: Current Git status (clean, dirty, stashed)
- **Sync Status**: Relationship with remote (up-to-date, ahead, behind)

### Groups
Collections of related repositories for batch operations:
- **Named collections**: E.g., "webdev", "tools", "personal"
- **Batch operations**: Run commands on all repositories in a group
- **Organized workflow**: Focus on specific project areas

### Shell Integration
Critical component that enables directory switching:
- **Required for `gman switch`**: Without it, switching won't change directories
- **Automatic installation**: Handled by installation scripts
- **Manual setup**: Source `~/.config/gman/shell-integration.sh`

## Command Structure

gman uses a modern, organized command structure with logical grouping and shortcuts:

### Command Groups

| Group | Shortcut | Purpose | Examples |
|-------|----------|---------|----------|
| `repo` | `r` | Repository management | `gman r add`, `gman r list` |
| `work` | `w` | Git workflow operations | `gman w status`, `gman w sync` |
| `tools` | `t` | Advanced utilities | `gman t find`, `gman t dashboard` |

### Common Patterns

```bash
# Full command
gman repo add /path/to/project my-project

# Using shortcut
gman r add /path/to/project my-project

# Help for any command
gman repo --help
gman work sync --help
```

## Repository Management

### Adding Repositories

```bash
# Add current directory with auto-generated alias
gman repo add

# Add specific path with auto-generated alias
gman repo add /path/to/project

# Add with custom alias
gman repo add /path/to/project my-project

# Add and assign to group immediately
gman repo add /path/to/frontend frontend-app
gman repo group add webdev frontend-app
```

### Listing Repositories

```bash
# Basic list
gman repo list

# Detailed view with paths and status
gman repo list --verbose

# List repositories in specific group
gman repo group list webdev
```

### Removing Repositories

```bash
# Remove by alias
gman repo remove my-project

# Remove from configuration only (keep files)
gman repo remove my-project --config-only
```

### Repository Information

```bash
# Show detailed information
gman repo info my-project

# Show recent activity
gman recent

# Switch to recently used repository
gman switch  # Interactive selection from recent
```

## Git Workflow Operations

### Checking Status

```bash
# Basic status across all repositories
gman work status

# Extended view with file counts and commit times
gman work status --extended

# Status for specific group
gman work status --group webdev

# Status with additional Git information
gman work status --verbose
```

**Status Indicators:**
- 🟢 **CLEAN**: No uncommitted changes
- 🔴 **DIRTY**: Uncommitted changes present
- 🟡 **STASHED**: Stashed changes available
- ✅ **UP-TO-DATE**: In sync with remote
- ↑ **X AHEAD**: Local commits not pushed
- ↓ **X BEHIND**: Remote commits not pulled
- 🔄 **X↑ Y↓**: Both ahead and behind

### Synchronizing Repositories

```bash
# Sync all repositories (default: fast-forward only)
gman work sync

# Sync specific group
gman work sync --group webdev

# Conditional sync options
gman work sync --only-dirty      # Only repos with changes
gman work sync --only-behind     # Only repos behind remote
gman work sync --only-ahead      # Only repos with unpushed commits

# Preview sync without executing
gman work sync --dry-run

# Sync with progress display
gman work sync --progress

# Alternative sync modes
gman work sync --rebase          # Use git pull --rebase
gman work sync --autostash       # Use git pull --autostash
```

### Batch Git Operations

```bash
# Commit changes across repositories
gman work commit -m "Update dependencies"
gman work commit -m "Fix bug" --add  # Stage all changes first

# Commit in specific group
gman work commit -m "Frontend updates" --group webdev

# Push changes
gman work push
gman work push --group webdev
gman work push --force  # Force push (use carefully)

# Stash operations
gman work stash save "WIP: feature development"
gman work stash list
gman work stash pop
gman work stash clear
```

### Branch Management

```bash
# List branches across repositories
gman work branch list
gman work branch list --remote  # Include remote branches

# Create branch in all repositories
gman work branch create feature/new-api

# Switch branches across repositories
gman work branch switch main
gman work branch switch feature/new-api

# Clean merged branches
gman work branch clean
gman work branch clean --main develop  # Specify main branch
```

## Enhanced Search System

gman's Phase 2 search system provides powerful discovery capabilities across all repositories.

### File Search

Fast file discovery using `fd`:

```bash
# Find files by name
gman tools find file config.yaml
gman tools find file package.json

# Pattern matching
gman tools find file "*.go"
gman tools find file "test_*.py"

# Search in specific group
gman tools find file "Dockerfile" --group backend

# Case-insensitive search
gman tools find file readme --ignore-case
```

**File Search Features:**
- Lightning-fast performance
- Respects .gitignore patterns
- Excludes build directories automatically
- Supports glob patterns and regex
- Interactive selection with fzf

### Content Search

Powerful text search using `ripgrep`:

```bash
# Basic content search
gman tools find content "TODO"
gman tools find content "FIXME"

# Regex patterns
gman tools find content "func.*Error"
gman tools find content "import.*axios"

# Case-sensitive search
gman tools find content "API_KEY" --case-sensitive

# Search in specific file types
gman tools find content "console.log" --type js
gman tools find content "fmt.Println" --type go

# Search with context
gman tools find content "error" --context 3  # 3 lines before/after
```

**Content Search Features:**
- Ultra-fast regex search
- Context line display
- File type filtering
- Respects .gitignore
- Multi-repository results

### Search Tips

```bash
# Combine with other tools
gman tools find file "*.md" | head -10
gman tools find content "TODO" | grep -v test

# Use quotes for complex patterns
gman tools find content "func.*\(.*error\)"

# Search specific groups for focused results
gman tools find file "config" --group backend
```

## Interactive Dashboard

The TUI dashboard provides a unified interface for repository management and search.

### Launching the Dashboard

```bash
# Launch dashboard
gman tools dashboard

# Alternative commands
gman tools tui
gman tools ui
gman dash
```

### Dashboard Layout

```
┌─ Repositories (1) ─┬─ Status (2) ──────┬─ Actions (5) ────┐
│ • Select repos     │ • Detailed status │ • Quick commands │
│ • Filter & search  │ • Branch info     │ • Git operations │
│ • Group management │ • File changes    │ • Interactive    │
└────────────────────┴───────────────────┴──────────────────┘
┌─ Search (3) ────────────────────────────┬─ Preview (4) ────┐
│ • Files & content across repos         │ • File content   │
│ • Integrated fzf support              │ • Commit details │
│ • Real-time results                    │ • Live updates   │
└────────────────────────────────────────┴──────────────────┘
```

### Navigation

| Key | Action |
|-----|--------|
| `1-5` | Switch to numbered panel |
| `Tab`, `Shift+Tab` | Navigate between panels |
| `j`, `k` | Move up/down in lists |
| `Enter` | Select/execute item |
| `/` | Start search |
| `Esc` | Cancel/go back |
| `q` | Quit dashboard |

### Quick Actions

| Key | Action |
|-----|--------|
| `r` | Refresh status |
| `s` | Sync repository |
| `c` | Commit changes |
| `p` | Push changes |
| `o` | Open in file manager |
| `t` | Open terminal |

### Dashboard Features

- **Real-time status**: Live repository status updates
- **Integrated search**: File and content search with preview
- **Context-aware actions**: Actions adapt to repository state
- **Group filtering**: Focus on specific repository groups
- **Keyboard shortcuts**: Efficient navigation without mouse

## Repository Groups

Groups organize repositories for batch operations and logical organization.

### Creating Groups

```bash
# Create group with repositories
gman repo group create webdev frontend-app backend-api

# Create empty group and add repositories later
gman repo group create tools
gman repo group add tools cli-utils monitoring-scripts

# Create with description
gman repo group create webdev frontend-app backend-api --description "Web development projects"
```

### Managing Groups

```bash
# List all groups
gman repo group list

# Show group details
gman repo group info webdev

# Add repositories to group
gman repo group add webdev new-frontend-project

# Remove repositories from group
gman repo group remove webdev old-project

# Delete group (repositories remain in config)
gman repo group delete old-group
```

### Using Groups in Operations

```bash
# Status for specific group
gman work status --group webdev

# Sync group repositories
gman work sync --group webdev

# Commit across group
gman work commit -m "Update dependencies" --group webdev

# Search within group
gman tools find file "*.json" --group webdev
```

### Group Best Practices

- **Logical organization**: Group by project, technology, or workflow
- **Consistent naming**: Use clear, descriptive group names
- **Size consideration**: Keep groups manageable (5-15 repositories)
- **Regular maintenance**: Update groups as projects evolve

Examples:
- `webdev`: Frontend and backend web projects
- `tools`: CLI tools and utilities
- `personal`: Personal projects and experiments
- `work`: Professional work repositories
- `client-xyz`: Client-specific projects

## Configuration

gman uses YAML configuration stored at `~/.config/gman/config.yml`.

### Configuration Structure

```yaml
repositories:
  backend-api: /home/user/projects/backend
  frontend-app: /home/user/projects/frontend
  tools-cli: /home/user/tools/gman

groups:
  webdev:
    name: "Web Development"
    description: "Frontend and backend projects"
    repositories: ["frontend-app", "backend-api"]
    created: "2024-01-15T10:30:00Z"

settings:
  parallel_jobs: 5
  show_last_commit: true
  default_sync_mode: "ff-only"
  
recent:
  - alias: "backend-api"
    last_accessed: "2024-01-15T14:30:00Z"
  - alias: "frontend-app"
    last_accessed: "2024-01-15T14:25:00Z"
```

### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `parallel_jobs` | 5 | Number of concurrent operations |
| `show_last_commit` | true | Show last commit in status |
| `default_sync_mode` | "ff-only" | Default sync mode |

### Sync Modes

- **`ff-only`**: Only fast-forward merges (safest)
- **`rebase`**: Use `git pull --rebase`
- **`autostash`**: Use `git pull --autostash`

### Manual Configuration

```bash
# Edit configuration directly
$EDITOR ~/.config/gman/config.yml

# Validate configuration
gman repo list --validate

# Reset configuration
rm ~/.config/gman/config.yml
gman tools setup  # Re-run setup wizard
```

## Advanced Features

### Directory Switching

```bash
# Interactive repository selection
gman switch

# Switch to specific repository
gman switch my-project

# Switch with fuzzy matching
gman switch cli  # Matches repositories containing "cli"

# Recent repositories (fastest access)
gman recent
gman switch  # Shows recent repositories first
```

### Diff and Comparison

```bash
# Compare files between branches
gman diff file my-repo main develop -- src/config.js

# Cross-repository file comparison
gman diff cross-repo repo1 repo2 -- package.json

# Use external diff tools
gman diff file my-repo main develop -- src/app.js --tool meld
```

### Worktree Management

```bash
# Create worktree for parallel development
gman worktree add my-repo /tmp/feature-branch --branch feature/new-api

# List worktrees
gman worktree list my-repo

# Remove completed worktree
gman worktree remove my-repo /tmp/feature-branch
```

### Migration and Maintenance

```bash
# Migrate to dependency injection pattern
gman migrate-di --analyze
gman migrate-di --apply

# Verify installation
./scripts/verify-setup.sh

# Update external dependencies
./scripts/setup-dependencies.sh
```

## Tips & Best Practices

### Repository Organization

1. **Use consistent aliases**: Prefer kebab-case for repository aliases
2. **Organize by purpose**: Group related repositories together
3. **Keep paths organized**: Use consistent directory structures
4. **Regular cleanup**: Remove obsolete repositories and groups

### Workflow Optimization

1. **Use groups extensively**: Organize repositories by project or technology
2. **Leverage recent access**: `gman switch` shows recent repositories first
3. **Conditional sync**: Use `--only-dirty` or `--only-behind` for efficiency
4. **Dry-run operations**: Preview changes with `--dry-run`

### Search Efficiency

1. **Learn regex patterns**: Powerful content search with regular expressions
2. **Use groups for scope**: Search within relevant repository groups
3. **Combine tools**: Pipe search results to other command-line tools
4. **Interactive selection**: Let fzf help with file/content selection

### Configuration Management

1. **Backup configuration**: Keep `~/.config/gman/config.yml` in version control
2. **Environment-specific configs**: Use different configs for work/personal
3. **Regular maintenance**: Clean up old repositories and groups
4. **Share group configurations**: Export/import group definitions across teams

### Performance Tips

1. **Adjust parallel jobs**: Increase `parallel_jobs` for faster operations
2. **Use external dependencies**: Install fd, rg, fzf for best performance
3. **Regular Git maintenance**: Keep repositories clean and optimized
4. **Monitor disk space**: Large repositories can slow operations

### Troubleshooting

1. **Check dependencies**: Run `./scripts/setup-dependencies.sh --verify-only`
2. **Verify shell integration**: Ensure `gman switch` changes directories
3. **Monitor performance**: Use `--progress` flag for long operations
4. **Check terminal compatibility**: Use `gman tools dashboard --debug`

For detailed troubleshooting, see [TROUBLESHOOTING.md](TROUBLESHOOTING.md) and [DEPLOYMENT.md](../DEPLOYMENT.md).

## Getting Help

- **Command help**: `gman --help`, `gman <command> --help`
- **Installation issues**: See [DEPLOYMENT.md](../DEPLOYMENT.md)
- **Search problems**: See [SEARCH_GUIDE.md](SEARCH_GUIDE.md)
- **Configuration**: See [CONFIGURATION.md](CONFIGURATION.md)
- **Command reference**: See [COMMAND_REFERENCE.md](COMMAND_REFERENCE.md)

Happy multi-repository management with gman! 🚀