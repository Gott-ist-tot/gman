# gman Command Reference

Complete reference documentation for all gman commands, options, and usage patterns.

## Table of Contents

- [Global Options](#global-options)
- [Repository Management Commands](#repository-management-commands)
- [Git Workflow Commands](#git-workflow-commands)
- [Search and Tools Commands](#search-and-tools-commands)
- [Quick Operations](#quick-operations)
- [Utility Commands](#utility-commands)
- [Command Examples](#command-examples)
- [Exit Codes](#exit-codes)

## Global Options

These options work with any gman command:

| Option | Description |
|--------|-------------|
| `--help, -h` | Show help information |
| `--version` | Show version information |
| `--config PATH` | Use custom configuration file |
| `--verbose, -v` | Enable verbose output |
| `--quiet, -q` | Suppress non-essential output |

## Repository Management Commands

### `gman repo` (alias: `gman r`)

Repository management operations.

#### `gman repo add [PATH] [ALIAS]`

Add a repository to gman configuration.

**Arguments:**
- `PATH` (optional): Path to Git repository (default: current directory)
- `ALIAS` (optional): Custom alias for repository (default: auto-generated from directory name)

**Options:**
| Option | Description |
|--------|-------------|
| `--group GROUP` | Add repository to specified group |
| `--force` | Overwrite existing repository with same alias |

**Examples:**
```bash
# Add current directory with auto-generated alias
gman repo add

# Add specific path with auto-generated alias
gman repo add /path/to/project

# Add with custom alias
gman repo add /path/to/project my-project

# Add to specific group
gman repo add /path/to/frontend frontend-app --group webdev
```

#### `gman repo remove ALIAS`

Remove repository from gman configuration.

**Arguments:**
- `ALIAS` (required): Repository alias to remove

**Options:**
| Option | Description |
|--------|-------------|
| `--config-only` | Remove from config only, keep files on disk |
| `--force` | Skip confirmation prompt |

**Examples:**
```bash
# Remove repository (with confirmation)
gman repo remove my-project

# Remove without confirmation
gman repo remove my-project --force
```

#### `gman repo list`

List all configured repositories.

**Options:**
| Option | Description |
|--------|-------------|
| `--verbose, -v` | Show detailed information (paths, status) |
| `--group GROUP` | Show only repositories in specified group |
| `--format FORMAT` | Output format: table, json, yaml |

**Examples:**
```bash
# Basic list
gman repo list

# Detailed view
gman repo list --verbose

# JSON output
gman repo list --format json
```

#### `gman repo info ALIAS`

Show detailed information about a repository.

**Arguments:**
- `ALIAS` (required): Repository alias

**Examples:**
```bash
gman repo info my-project
```

### Repository Group Commands

#### `gman repo group create NAME [REPOS...]`

Create a new repository group.

**Arguments:**
- `NAME` (required): Group name
- `REPOS...` (optional): Repository aliases to add to group

**Options:**
| Option | Description |
|--------|-------------|
| `--description DESC` | Group description |

**Examples:**
```bash
# Create empty group
gman repo group create webdev

# Create group with repositories
gman repo group create webdev frontend-app backend-api

# Create with description
gman repo group create webdev frontend-app --description "Web development projects"
```

#### `gman repo group list [GROUP]`

List repository groups or show group details.

**Arguments:**
- `GROUP` (optional): Specific group to show details

**Examples:**
```bash
# List all groups
gman repo group list

# Show specific group details
gman repo group list webdev
```

#### `gman repo group add GROUP REPOS...`

Add repositories to a group.

**Arguments:**
- `GROUP` (required): Group name
- `REPOS...` (required): Repository aliases to add

**Examples:**
```bash
gman repo group add webdev new-frontend-project
gman repo group add tools cli-utils monitoring-scripts
```

#### `gman repo group remove GROUP REPOS...`

Remove repositories from a group.

**Arguments:**
- `GROUP` (required): Group name
- `REPOS...` (required): Repository aliases to remove

**Examples:**
```bash
gman repo group remove webdev old-project
```

#### `gman repo group delete GROUP`

Delete a repository group.

**Arguments:**
- `GROUP` (required): Group name to delete

**Options:**
| Option | Description |
|--------|-------------|
| `--force` | Skip confirmation prompt |

**Examples:**
```bash
gman repo group delete old-group
gman repo group delete temp --force
```

## Git Workflow Commands

### `gman work` (alias: `gman w`)

Git workflow operations across repositories.

#### `gman work status`

Show Git status across all repositories.

**Options:**
| Option | Description |
|--------|-------------|
| `--extended, -e` | Show extended information (file counts, commit times) |
| `--group GROUP` | Show status for specific group only |
| `--verbose, -v` | Include additional Git information |
| `--format FORMAT` | Output format: table, json, yaml |

**Examples:**
```bash
# Basic status
gman work status

# Extended view with details
gman work status --extended

# Status for specific group
gman work status --group webdev

# JSON output for scripting
gman work status --format json
```

#### `gman work sync`

Synchronize repositories with their remotes.

**Options:**
| Option | Description |
|--------|-------------|
| `--group GROUP` | Sync specific group only |
| `--only-dirty` | Sync only repositories with uncommitted changes |
| `--only-behind` | Sync only repositories behind remote |
| `--only-ahead` | Sync only repositories with unpushed commits |
| `--dry-run` | Preview sync operations without executing |
| `--progress` | Show progress bars for operations |
| `--rebase` | Use `git pull --rebase` instead of merge |
| `--autostash` | Use `git pull --autostash` |
| `--parallel JOBS` | Number of parallel operations (default: 5) |

**Examples:**
```bash
# Sync all repositories
gman work sync

# Sync with progress display
gman work sync --progress

# Sync only dirty repositories
gman work sync --only-dirty

# Preview sync operations
gman work sync --dry-run

# Sync with rebase
gman work sync --rebase

# Sync specific group
gman work sync --group webdev
```

#### `gman work commit`

Commit changes across repositories.

**Arguments:**
- `-m MESSAGE` (required): Commit message

**Options:**
| Option | Description |
|--------|-------------|
| `--add, -a` | Stage all changes before committing |
| `--group GROUP` | Commit in specific group only |
| `--dry-run` | Preview commit operations |
| `--allow-empty` | Allow empty commits |

**Examples:**
```bash
# Commit staged changes
gman work commit -m "Fix bug in authentication"

# Stage and commit all changes
gman work commit -m "Update dependencies" --add

# Commit in specific group
gman work commit -m "Frontend updates" --group webdev

# Preview commit operations
gman work commit -m "Test commit" --dry-run
```

#### `gman work push`

Push changes to remote repositories.

**Options:**
| Option | Description |
|--------|-------------|
| `--group GROUP` | Push specific group only |
| `--force` | Force push (use with caution) |
| `--set-upstream` | Set upstream branch for new branches |
| `--dry-run` | Preview push operations |

**Examples:**
```bash
# Push all repositories
gman work push

# Push specific group
gman work push --group webdev

# Force push (dangerous)
gman work push --force

# Set upstream for new branches
gman work push --set-upstream
```

#### Stash Commands

##### `gman work stash save [MESSAGE]`

Save changes to stash across repositories.

**Arguments:**
- `MESSAGE` (optional): Stash message

**Options:**
| Option | Description |
|--------|-------------|
| `--group GROUP` | Stash in specific group only |
| `--include-untracked` | Include untracked files |

**Examples:**
```bash
gman work stash save "WIP: new feature"
gman work stash save --group webdev
```

##### `gman work stash list`

List stashes across repositories.

**Options:**
| Option | Description |
|--------|-------------|
| `--group GROUP` | List stashes for specific group only |

##### `gman work stash pop`

Apply and remove latest stash across repositories.

**Options:**
| Option | Description |
|--------|-------------|
| `--group GROUP` | Pop stashes in specific group only |

##### `gman work stash clear`

Clear all stashes across repositories.

**Options:**
| Option | Description |
|--------|-------------|
| `--group GROUP` | Clear stashes in specific group only |
| `--force` | Skip confirmation prompt |

#### Branch Commands

##### `gman work branch list`

List branches across repositories.

**Options:**
| Option | Description |
|--------|-------------|
| `--remote, -r` | Include remote branches |
| `--verbose, -v` | Show detailed branch information |
| `--group GROUP` | List branches for specific group only |

##### `gman work branch create BRANCH`

Create branch across repositories.

**Arguments:**
- `BRANCH` (required): Branch name to create

**Options:**
| Option | Description |
|--------|-------------|
| `--group GROUP` | Create in specific group only |
| `--checkout` | Switch to new branch after creation |

##### `gman work branch switch BRANCH`

Switch branches across repositories.

**Arguments:**
- `BRANCH` (required): Branch name to switch to

**Options:**
| Option | Description |
|--------|-------------|
| `--group GROUP` | Switch in specific group only |
| `--create, -c` | Create branch if it doesn't exist |

##### `gman work branch clean`

Clean merged branches across repositories.

**Options:**
| Option | Description |
|--------|-------------|
| `--main BRANCH` | Specify main branch (default: auto-detect) |
| `--group GROUP` | Clean in specific group only |
| `--dry-run` | Preview operations |
| `--force` | Skip confirmation |

## Search and Tools Commands

### `gman tools` (alias: `gman t`)

Advanced utilities and search operations.

#### `gman tools find file PATTERN`

Search for files across repositories.

**Arguments:**
- `PATTERN` (required): File name pattern to search for

**Options:**
| Option | Description |
|--------|-------------|
| `--group GROUP` | Search in specific group only |
| `--ignore-case, -i` | Case-insensitive search |
| `--type TYPE` | Filter by file type (e.g., js, py, go) |
| `--exclude PATTERN` | Exclude files matching pattern |
| `--max-results N` | Limit number of results |

**Examples:**
```bash
# Find specific file
gman tools find file config.yaml

# Pattern matching
gman tools find file "*.go"
gman tools find file "test_*.py"

# Search in specific group
gman tools find file "Dockerfile" --group backend

# Case-insensitive search
gman tools find file readme --ignore-case

# Limit results
gman tools find file "*.js" --max-results 20
```

#### `gman tools find content PATTERN`

Search file contents across repositories.

**Arguments:**
- `PATTERN` (required): Text pattern to search for (supports regex)

**Options:**
| Option | Description |
|--------|-------------|
| `--group GROUP` | Search in specific group only |
| `--ignore-case, -i` | Case-insensitive search |
| `--type TYPE` | Search in specific file types |
| `--context N` | Show N lines of context around matches |
| `--max-results N` | Limit number of results |
| `--exclude PATTERN` | Exclude files matching pattern |

**Examples:**
```bash
# Basic content search
gman tools find content "TODO"
gman tools find content "FIXME"

# Regex patterns
gman tools find content "func.*Error"
gman tools find content "import.*axios"

# Search with context
gman tools find content "error" --context 3

# File type filtering
gman tools find content "console.log" --type js
gman tools find content "fmt.Println" --type go

# Group-specific search
gman tools find content "API_KEY" --group backend
```

#### `gman tools dashboard`

Launch interactive TUI dashboard.

**Options:**
| Option | Description |
|--------|-------------|
| `--theme THEME` | Color theme: dark, light |
| `--debug` | Show terminal compatibility information |
| `--force` | Force TUI mode (bypass compatibility checks) |

**Examples:**
```bash
# Launch dashboard
gman tools dashboard

# Use light theme
gman tools dashboard --theme light

# Debug terminal issues
gman tools dashboard --debug
```

#### `gman tools setup`

Interactive setup wizard for new users.

**Options:**
| Option | Description |
|--------|-------------|
| `--discover PATH` | Auto-discover repositories in path |
| `--depth N` | Repository discovery depth (default: 3) |
| `--auto-confirm` | Skip confirmation prompts |

**Examples:**
```bash
# Run setup wizard
gman tools setup

# Auto-discover in specific path
gman tools setup --discover ~/Projects

# Automated setup
gman tools setup --auto-confirm
```

## Quick Operations

### `gman switch [ALIAS]`

Switch to repository directory.

**Arguments:**
- `ALIAS` (optional): Repository alias (interactive selection if omitted)

**Options:**
| Option | Description |
|--------|-------------|
| `--fuzzy` | Enable fuzzy matching for alias |

**Examples:**
```bash
# Interactive selection
gman switch

# Switch to specific repository
gman switch my-project

# Fuzzy matching
gman switch cli  # Matches repositories containing "cli"
```

**Note:** Requires shell integration to function properly.

### `gman recent`

Show recently accessed repositories.

**Options:**
| Option | Description |
|--------|-------------|
| `--limit N` | Number of recent repositories to show (default: 10) |
| `--format FORMAT` | Output format: table, json, yaml |

**Examples:**
```bash
# Show recent repositories
gman recent

# Show last 5 repositories
gman recent --limit 5
```

## Utility Commands

### `gman completion [SHELL]`

Generate shell completion scripts.

**Arguments:**
- `SHELL` (optional): Shell type: bash, zsh, fish, powershell

**Examples:**
```bash
# Generate for current shell
gman completion

# Generate for specific shell
gman completion bash
gman completion zsh

# Install completions
gman completion bash > /etc/bash_completion.d/gman
gman completion zsh > ~/.config/zsh/completions/_gman
```

### `gman diff`

File comparison operations.

#### `gman diff file REPO BRANCH1 BRANCH2 -- FILE`

Compare file between branches in a repository.

**Arguments:**
- `REPO` (required): Repository alias
- `BRANCH1` (required): First branch to compare
- `BRANCH2` (required): Second branch to compare
- `FILE` (required): File path to compare

**Options:**
| Option | Description |
|--------|-------------|
| `--tool TOOL` | External diff tool (meld, vimdiff, etc.) |

**Examples:**
```bash
# Compare file between branches
gman diff file my-repo main develop -- src/config.js

# Use external diff tool
gman diff file my-repo main develop -- src/app.js --tool meld
```

#### `gman diff cross-repo REPO1 REPO2 -- FILE`

Compare file between different repositories.

**Arguments:**
- `REPO1` (required): First repository alias
- `REPO2` (required): Second repository alias
- `FILE` (required): File path to compare

**Options:**
| Option | Description |
|--------|-------------|
| `--tool TOOL` | External diff tool |

**Examples:**
```bash
# Compare file between repositories
gman diff cross-repo frontend backend -- package.json

# Use external tool
gman diff cross-repo repo1 repo2 -- config.yaml --tool meld
```

### `gman worktree`

Git worktree management.

#### `gman worktree add REPO PATH`

Create a new worktree.

**Arguments:**
- `REPO` (required): Repository alias
- `PATH` (required): Path for new worktree

**Options:**
| Option | Description |
|--------|-------------|
| `--branch BRANCH` | Create/checkout branch in worktree |
| `--force` | Force creation even if path exists |

**Examples:**
```bash
# Create worktree with new branch
gman worktree add my-repo /tmp/feature-branch --branch feature/new-api

# Create worktree from existing branch
gman worktree add my-repo /tmp/hotfix --branch hotfix/critical-bug
```

#### `gman worktree list REPO`

List worktrees for a repository.

**Arguments:**
- `REPO` (required): Repository alias

#### `gman worktree remove REPO PATH`

Remove a worktree.

**Arguments:**
- `REPO` (required): Repository alias
- `PATH` (required): Worktree path to remove

**Options:**
| Option | Description |
|--------|-------------|
| `--force` | Force removal even with uncommitted changes |

### Migration Commands

#### `gman migrate-di`

Analyze and migrate dependency injection usage.

**Options:**
| Option | Description |
|--------|-------------|
| `--analyze` | Analyze current DI usage |
| `--apply` | Apply automatic migration |
| `--dry-run` | Preview migration changes |

**Examples:**
```bash
# Analyze current usage
gman migrate-di --analyze

# Preview migration
gman migrate-di --dry-run

# Apply migration
gman migrate-di --apply
```

## Command Examples

### Common Workflows

#### Initial Setup
```bash
# Install and setup
./scripts/install.sh
gman tools setup

# Add repositories
gman repo add /path/to/frontend frontend-app
gman repo add /path/to/backend backend-api

# Create groups
gman repo group create webdev frontend-app backend-api
```

#### Daily Workflow
```bash
# Check status across all repositories
gman work status --extended

# Sync repositories behind remote
gman work sync --only-behind --progress

# Work on specific group
gman work status --group webdev
gman switch frontend-app

# Search across repositories
gman tools find content "TODO.*urgent"
gman tools find file "*.config.js"
```

#### Batch Operations
```bash
# Commit changes across group
gman work commit -m "Update dependencies" --add --group webdev

# Push changes
gman work push --group webdev

# Branch management
gman work branch create feature/new-api --group webdev
gman work branch switch feature/new-api --group webdev
```

#### Search and Discovery
```bash
# Find configuration files
gman tools find file "config.*" --type json

# Search for specific patterns
gman tools find content "process\.env\." --context 2

# Interactive search
gman tools dashboard
```

## Exit Codes

gman uses standard exit codes:

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Command line usage error |
| 3 | Configuration error |
| 4 | Repository not found |
| 5 | Git operation failed |
| 126 | Command not executable |
| 127 | Command not found |
| 130 | Interrupted by user (Ctrl-C) |

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GMAN_CONFIG` | Path to configuration file | `~/.config/gman/config.yml` |
| `GMAN_DEBUG` | Enable debug output | `false` |
| `GMAN_PARALLEL_JOBS` | Default parallel jobs | `5` |
| `GMAN_NO_COLOR` | Disable colored output | `false` |

## See Also

- [USER_GUIDE.md](USER_GUIDE.md) - Comprehensive user guide
- [SEARCH_GUIDE.md](SEARCH_GUIDE.md) - Enhanced search system guide
- [CONFIGURATION.md](CONFIGURATION.md) - Configuration reference
- [DEPLOYMENT.md](../DEPLOYMENT.md) - Installation and deployment
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Troubleshooting guide