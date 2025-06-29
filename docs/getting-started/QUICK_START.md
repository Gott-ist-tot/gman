# gman Quick Start Tutorial

Get up and running with gman in 5 minutes! This tutorial covers the essential features you need to start managing multiple Git repositories efficiently.

## Prerequisites

- gman installed and shell integration configured (see [Installation Guide](INSTALLATION.md))
- At least one Git repository on your system

## Step 1: Initial Setup

Run the interactive setup wizard to get started:

```bash
gman tools setup
```

This wizard will:
- Discover existing Git repositories
- Configure basic settings
- Create your first repository groups
- Show you the essential commands

**Alternative**: Manual setup if you prefer:
```bash
# Add your first repository
gman repo add /path/to/your/project my-project
# Or add current directory
gman repo add . current-project
```

## Step 2: Basic Repository Management

### Adding Repositories

```bash
# Add current directory with auto-generated alias
gman repo add

# Add specific path with custom alias  
gman repo add /path/to/backend backend-api
gman repo add /path/to/frontend frontend-app

# Create a group for related repositories
gman repo group create webdev frontend-app backend-api
```

### Viewing Your Repositories

```bash
# List all repositories
gman repo list

# Show detailed status
gman work status

# Extended view with file counts and commit times
gman work status --extended
```

**Expected Output:**
```
Alias       Branch   Workspace      Sync Status     
─────────── ──────── ─────────────── ────────────────
* backend   main     🟢 CLEAN       ✅ UP-TO-DATE   
  frontend  develop  🔴 DIRTY       ↑ 2 AHEAD      
```

## Step 3: Navigate Between Repositories

### Quick Directory Switching

```bash
# Interactive selection from all repositories
gman switch

# Switch to specific repository
gman switch backend-api

# View recently accessed repositories
gman recent
```

**Important**: The `gman switch` command actually changes your current directory - this is why shell integration is required!

## Step 4: Git Workflow Operations

### Checking Status Across Repositories

```bash
# Quick status check
gman work status

# Extended information
gman work status --extended --group webdev
```

### Synchronizing Repositories

```bash
# Sync all repositories
gman work sync

# Sync only repositories that are behind
gman work sync --only-behind

# Preview what would be synced (dry run)
gman work sync --dry-run

# Sync specific group with progress display
gman work sync --group webdev --progress
```

### Batch Operations

```bash
# Commit changes across repositories
gman work commit -m "Update dependencies" --add

# Push changes to remotes
gman work push --group webdev

# Stash uncommitted work
gman work stash save "WIP: feature development"
```

## Step 5: Enhanced Search Features

### File Search

```bash
# Find specific files across all repositories
gman tools find file config.yaml
gman tools find file package.json

# Pattern matching
gman tools find file "*.go"
gman tools find file "test_*.py"

# Search within specific group
gman tools find file "Dockerfile" --group backend
```

### Content Search

```bash
# Search for text within files
gman tools find content "TODO"
gman tools find content "FIXME"

# Use regex patterns for advanced search
gman tools find content "func.*Error"
gman tools find content "import.*axios"

# Search specific file types
gman tools find content "console.log" --type js
```

**Tip**: These search commands use `fzf` for interactive selection - use arrow keys to navigate and Enter to select!

## Step 6: Interactive Dashboard (Optional)

For a visual, interactive experience:

```bash
# Launch the TUI dashboard
gman tools dashboard
```

**Dashboard Navigation:**
- `1-5`: Switch between panels
- `Tab`/`Shift+Tab`: Navigate panels
- `/`: Start search
- `r`: Refresh status
- `q`: Quit dashboard

## Common Workflow Examples

### Daily Development Workflow

```bash
# 1. Check status of all projects
gman work status --extended

# 2. Switch to project you want to work on
gman switch frontend-app

# 3. Work on your code...

# 4. When ready, sync and commit
gman work sync --only-dirty
gman work commit -m "Add new feature" --add
gman work push
```

### Project Organization

```bash
# Create groups for different projects
gman repo group create client-alpha project1 project2
gman repo group create internal tools monitoring

# Work with specific groups
gman work status --group client-alpha
gman work sync --group internal
gman tools find file "*.md" --group client-alpha
```

### Finding Files and Content

```bash
# Find configuration files across all projects
gman tools find file config

# Search for TODOs that need attention
gman tools find content "TODO.*urgent"

# Find all test files in a group
gman tools find file "test" --group backend
```

## Status Indicators Reference

### Workspace Status
- 🟢 **CLEAN**: No uncommitted changes
- 🔴 **DIRTY**: Uncommitted changes present  
- 🟡 **STASHED**: Stashed changes available

### Sync Status
- ✅ **UP-TO-DATE**: In sync with remote
- ↑ **X AHEAD**: Local commits not pushed
- ↓ **X BEHIND**: Remote commits not pulled
- 🔄 **X↑ Y↓**: Both ahead and behind

## Essential Commands Summary

| Command | Purpose |
|---------|---------|
| `gman tools setup` | Interactive setup wizard |
| `gman repo add [path] [alias]` | Add repository |
| `gman work status [--extended]` | Check repository status |
| `gman switch [alias]` | Change to repository directory |
| `gman work sync [--group <name>]` | Synchronize with remotes |
| `gman tools find file <pattern>` | Search for files |
| `gman tools find content <pattern>` | Search file contents |
| `gman tools dashboard` | Launch interactive TUI |

## Getting Help

- **Command Help**: `gman --help` or `gman <command> --help`
- **Comprehensive Guide**: [User Guide](../user-guide/USER_GUIDE.md)
- **Issues**: [Troubleshooting Guide](../troubleshooting/TROUBLESHOOTING.md)

## Next Steps

Now that you're familiar with the basics:

1. **Explore Advanced Features**: Check out the [User Guide](../user-guide/USER_GUIDE.md) for workflows and best practices
2. **Customize Configuration**: See [Configuration Guide](../user-guide/CONFIGURATION.md) for customization options
3. **Learn Search Patterns**: Read [Search System Guide](../features/SEARCH_SYSTEM.md) for advanced search techniques
4. **Try the Dashboard**: Explore the interactive TUI with `gman tools dashboard`

Happy repository management! 🚀