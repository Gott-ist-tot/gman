# gman Configuration Guide

Complete guide to configuring gman for optimal multi-repository management workflows.

## Table of Contents

- [Configuration Overview](#configuration-overview)
- [Configuration File Structure](#configuration-file-structure)
- [Repository Configuration](#repository-configuration)
- [Groups Configuration](#groups-configuration)
- [Settings Configuration](#settings-configuration)
- [Recent Usage Tracking](#recent-usage-tracking)
- [Environment Variables](#environment-variables)
- [Advanced Configuration](#advanced-configuration)
- [Configuration Management](#configuration-management)
- [Troubleshooting Configuration](#troubleshooting-configuration)

## Configuration Overview

gman uses a YAML-based configuration system that provides flexible, human-readable configuration management.

### Configuration Locations

| File | Purpose | Priority |
|------|---------|----------|
| `~/.config/gman/config.yml` | Main configuration | Default |
| Custom path via `--config` | Override configuration | High |
| Environment variables | Runtime overrides | Highest |

### Configuration Creation

Configuration is created automatically when:
1. Running the setup wizard: `gman tools setup`
2. Adding your first repository: `gman repo add`
3. Manual creation with proper YAML structure

## Configuration File Structure

### Complete Configuration Example

```yaml
# gman Configuration File
# Location: ~/.config/gman/config.yml

# Repository mappings: alias -> absolute path
repositories:
  backend-api: /home/user/projects/backend
  frontend-app: /home/user/projects/frontend
  infrastructure: /home/user/projects/infra
  cli-tools: /home/user/tools/gman
  personal-site: /home/user/personal/website

# Repository groups for batch operations
groups:
  webdev:
    name: "Web Development"
    description: "Frontend and backend web projects"
    repositories: ["frontend-app", "backend-api"]
    created: "2024-01-15T10:30:00Z"
  
  tools:
    name: "Development Tools"
    description: "CLI tools and utilities"
    repositories: ["cli-tools"]
    created: "2024-01-16T09:15:00Z"
  
  personal:
    name: "Personal Projects"
    description: "Personal projects and experiments"
    repositories: ["personal-site"]
    created: "2024-01-17T14:20:00Z"

# Global settings
settings:
  parallel_jobs: 5
  show_last_commit: true
  default_sync_mode: "ff-only"
  sync_timeout: 300
  max_recent_repositories: 10
  color_output: true
  confirm_destructive_operations: true

# Recent repository access tracking
recent:
  - alias: "backend-api"
    last_accessed: "2024-01-20T14:30:00Z"
  - alias: "frontend-app"
    last_accessed: "2024-01-20T14:25:00Z"
  - alias: "cli-tools"
    last_accessed: "2024-01-20T10:15:00Z"

# Optional: Advanced configuration sections
advanced:
  git:
    fetch_timeout: 60
    clone_timeout: 300
    max_retries: 3
  
  search:
    file_search_timeout: 10
    content_search_timeout: 30
    max_search_results: 1000
  
  ui:
    dashboard_theme: "dark"
    terminal_compatibility_check: true
    animation_enabled: true
```

### Minimal Configuration Example

```yaml
# Minimal working configuration
repositories:
  my-project: /path/to/my/project

settings:
  parallel_jobs: 3
```

## Repository Configuration

### Adding Repositories

#### Automatic Addition (Recommended)
```bash
# Add with auto-generated alias
gman repo add /path/to/project

# Add with custom alias
gman repo add /path/to/project my-project

# Add current directory
gman repo add . current-project
```

#### Manual Configuration
```yaml
repositories:
  # Simple path mapping
  project-alias: /absolute/path/to/repository
  
  # Various project types
  frontend: /home/user/web/frontend
  backend: /home/user/web/backend
  mobile: /home/user/mobile/app
  docs: /home/user/documentation
  
  # Tools and utilities
  scripts: /home/user/scripts
  configs: /home/user/dotfiles
```

### Repository Validation

gman validates repositories on startup:

```bash
# Validate configuration
gman repo list --validate

# Check specific repository
gman repo info my-project
```

**Validation Checks:**
- Path exists and is accessible
- Directory contains `.git` folder
- Repository is not corrupted
- Permissions are sufficient

### Repository Path Guidelines

#### Best Practices
```yaml
repositories:
  # Good: Descriptive aliases
  ecommerce-frontend: /projects/ecommerce/frontend
  ecommerce-backend: /projects/ecommerce/backend
  user-service: /services/user-management
  
  # Avoid: Generic aliases
  # project1: /some/path
  # temp: /tmp/repo
```

#### Path Requirements
- **Absolute paths only**: Relative paths are not supported
- **Git repositories**: Must contain `.git` directory
- **Accessible locations**: gman must have read/write permissions
- **Stable paths**: Avoid temporary or mounted locations

## Groups Configuration

### Group Structure

```yaml
groups:
  group-name:
    name: "Human Readable Name"
    description: "Detailed description of the group"
    repositories: ["repo1", "repo2", "repo3"]
    created: "2024-01-15T10:30:00Z"
    # Optional metadata
    metadata:
      owner: "team-name"
      environment: "development"
      priority: "high"
```

### Creating Groups

#### Command Line Creation
```bash
# Create group with repositories
gman repo group create webdev frontend-app backend-api

# Create with description
gman repo group create tools cli-utils --description "Development tools"

# Create empty group
gman repo group create experiments
```

#### Manual Configuration
```yaml
groups:
  # Development environments
  development:
    name: "Development Environment"
    description: "All development repositories"
    repositories: ["api-dev", "frontend-dev", "database-dev"]
    created: "2024-01-15T10:30:00Z"
  
  # Production environments
  production:
    name: "Production Environment"
    description: "Production repositories requiring careful handling"
    repositories: ["api-prod", "frontend-prod"]
    created: "2024-01-15T11:00:00Z"
  
  # By technology
  javascript:
    name: "JavaScript Projects"
    description: "All JavaScript/Node.js projects"
    repositories: ["frontend-app", "api-server", "cli-tools"]
    created: "2024-01-16T09:00:00Z"
  
  python:
    name: "Python Projects"
    description: "Python applications and scripts"
    repositories: ["data-pipeline", "ml-models", "automation"]
    created: "2024-01-16T09:30:00Z"
```

### Group Management

#### Adding/Removing Repositories
```bash
# Add repositories to group
gman repo group add webdev new-project

# Remove repositories from group
gman repo group remove webdev old-project

# Replace group contents
gman repo group create webdev new-repo1 new-repo2 --force
```

#### Group Operations
```bash
# List group repositories
gman repo group list webdev

# Group-specific operations
gman work status --group webdev
gman work sync --group webdev
gman tools find file "*.js" --group webdev
```

### Group Best Practices

#### Organizational Strategies

**By Project:**
```yaml
groups:
  ecommerce:
    repositories: ["ecommerce-frontend", "ecommerce-backend", "ecommerce-mobile"]
  analytics:
    repositories: ["data-pipeline", "reporting-api", "dashboard"]
```

**By Environment:**
```yaml
groups:
  development:
    repositories: ["api-dev", "frontend-dev"]
  staging:
    repositories: ["api-staging", "frontend-staging"]
  production:
    repositories: ["api-prod", "frontend-prod"]
```

**By Technology:**
```yaml
groups:
  frontend:
    repositories: ["react-app", "vue-dashboard", "angular-admin"]
  backend:
    repositories: ["node-api", "python-service", "go-microservice"]
```

**By Team:**
```yaml
groups:
  frontend-team:
    repositories: ["web-app", "mobile-app", "design-system"]
  backend-team:
    repositories: ["api-gateway", "user-service", "payment-service"]
  devops:
    repositories: ["infrastructure", "deployment-scripts", "monitoring"]
```

## Settings Configuration

### Core Settings

```yaml
settings:
  # Concurrency control
  parallel_jobs: 5                    # Number of concurrent operations
  
  # Display preferences
  show_last_commit: true              # Show commit info in status
  color_output: true                  # Enable colored output
  
  # Git operation defaults
  default_sync_mode: "ff-only"        # Sync mode: ff-only, rebase, autostash
  sync_timeout: 300                   # Timeout for sync operations (seconds)
  
  # Recent tracking
  max_recent_repositories: 10         # Number of recent repos to track
  
  # Safety settings
  confirm_destructive_operations: true # Confirm dangerous operations
```

### Setting Descriptions

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `parallel_jobs` | integer | 5 | Concurrent operations (1-20) |
| `show_last_commit` | boolean | true | Display commit info in status |
| `color_output` | boolean | true | Enable ANSI color codes |
| `default_sync_mode` | string | "ff-only" | Default Git sync strategy |
| `sync_timeout` | integer | 300 | Sync operation timeout (seconds) |
| `max_recent_repositories` | integer | 10 | Recent repositories to track |
| `confirm_destructive_operations` | boolean | true | Confirm dangerous operations |

### Sync Modes

```yaml
settings:
  # Safe mode (default): Only fast-forward merges
  default_sync_mode: "ff-only"
  
  # Rebase mode: Use git pull --rebase
  default_sync_mode: "rebase"
  
  # Auto-stash mode: Use git pull --autostash
  default_sync_mode: "autostash"
```

**Sync Mode Comparison:**

| Mode | Safety | Merge Behavior | Use Case |
|------|--------|----------------|----------|
| `ff-only` | High | Fast-forward only | Safe default |
| `rebase` | Medium | Rebase local commits | Clean history |
| `autostash` | Low | Auto-stash changes | Convenience |

### Performance Settings

```yaml
settings:
  # Optimize for your system
  parallel_jobs: 8                    # Match CPU cores
  sync_timeout: 600                   # Longer timeout for slow networks
  
  # Reduce output for automation
  color_output: false                 # Disable colors in scripts
  show_last_commit: false            # Reduce status output
```

## Recent Usage Tracking

### Recent Configuration

```yaml
recent:
  - alias: "backend-api"
    last_accessed: "2024-01-20T14:30:00Z"
  - alias: "frontend-app"
    last_accessed: "2024-01-20T14:25:00Z"
  - alias: "infrastructure"
    last_accessed: "2024-01-19T16:10:00Z"

settings:
  max_recent_repositories: 10         # Adjust tracking size
```

### Recent Usage Features

- **Automatic tracking**: Updated when using `gman switch`
- **Ordered by access**: Most recent repositories first
- **Configurable size**: Control with `max_recent_repositories`
- **Interactive selection**: `gman switch` shows recent first

### Managing Recent History

```bash
# View recent repositories
gman recent

# Clear recent history (manual config edit required)
# Remove 'recent:' section from config.yml

# Disable recent tracking
# Set max_recent_repositories: 0
```

## Environment Variables

### Runtime Configuration

```bash
# Configuration file location
export GMAN_CONFIG="/custom/path/to/config.yml"

# Override settings
export GMAN_PARALLEL_JOBS=8
export GMAN_DEFAULT_SYNC_MODE="rebase"
export GMAN_DEBUG=true

# Disable features
export GMAN_NO_COLOR=true
export GMAN_DISABLE_RECENT=true
```

### Environment Variable Priority

1. **Command line options**: `gman --config /path/config.yml`
2. **Environment variables**: `GMAN_CONFIG=/path/config.yml`
3. **Default location**: `~/.config/gman/config.yml`

### Available Environment Variables

| Variable | Purpose | Example |
|----------|---------|---------|
| `GMAN_CONFIG` | Configuration file path | `/custom/gman/config.yml` |
| `GMAN_PARALLEL_JOBS` | Override parallel jobs | `8` |
| `GMAN_DEFAULT_SYNC_MODE` | Override sync mode | `rebase` |
| `GMAN_DEBUG` | Enable debug output | `true` |
| `GMAN_NO_COLOR` | Disable colors | `true` |
| `GMAN_DISABLE_RECENT` | Disable recent tracking | `true` |

## Advanced Configuration

### Git-Specific Settings

```yaml
advanced:
  git:
    # Operation timeouts
    fetch_timeout: 60                 # Git fetch timeout
    clone_timeout: 300                # Git clone timeout
    push_timeout: 120                 # Git push timeout
    
    # Retry behavior
    max_retries: 3                    # Max retry attempts
    retry_delay: 5                    # Delay between retries (seconds)
    
    # Default options
    fetch_options: ["--prune"]        # Additional git fetch options
    pull_options: ["--ff-only"]       # Additional git pull options
```

### Search Configuration

```yaml
advanced:
  search:
    # Search timeouts
    file_search_timeout: 10           # File search timeout
    content_search_timeout: 30        # Content search timeout
    
    # Result limits
    max_search_results: 1000          # Maximum search results
    default_context_lines: 3          # Default context for content search
    
    # External tool paths (if not in PATH)
    fd_path: "/usr/local/bin/fd"
    rg_path: "/usr/local/bin/rg"
    fzf_path: "/usr/local/bin/fzf"
```

### UI Configuration

```yaml
advanced:
  ui:
    # Dashboard settings
    dashboard_theme: "dark"           # Theme: dark, light
    terminal_compatibility_check: true # Check terminal capabilities
    animation_enabled: true           # Enable UI animations
    
    # Display preferences
    table_style: "rounded"            # Table style: rounded, ascii, unicode
    timestamp_format: "relative"     # Timestamp format: relative, absolute
    
    # Interactive settings
    fuzzy_matching_threshold: 0.6     # Fuzzy match sensitivity
```

### Custom Command Aliases

```yaml
# Not yet implemented - future feature
command_aliases:
  st: "work status"
  sync-all: "work sync --progress"
  find-todos: "tools find content 'TODO.*' --context 2"
  quick-commit: "work commit --add"
```

## Configuration Management

### Backup and Restore

#### Backup Configuration
```bash
# Backup current configuration
cp ~/.config/gman/config.yml ~/.config/gman/config.yml.backup

# Backup with timestamp
cp ~/.config/gman/config.yml ~/.config/gman/config.yml.$(date +%Y%m%d)

# Version control (recommended)
cd ~/.config/gman
git init
git add config.yml
git commit -m "Initial gman configuration"
```

#### Restore Configuration
```bash
# Restore from backup
cp ~/.config/gman/config.yml.backup ~/.config/gman/config.yml

# Reset to defaults (requires re-setup)
rm ~/.config/gman/config.yml
gman tools setup
```

### Configuration Validation

```bash
# Validate current configuration
gman repo list --validate

# Test configuration with custom file
gman --config /path/to/test-config.yml repo list

# Check for configuration issues
gman tools setup --check-config
```

### Configuration Sharing

#### Team Configuration Template
```yaml
# team-template.yml - Shared team configuration template
repositories:
  # Team members customize these paths
  project-frontend: /customize/path/to/frontend
  project-backend: /customize/path/to/backend
  project-mobile: /customize/path/to/mobile

groups:
  # Shared group structure
  development:
    name: "Development Environment"
    repositories: ["project-frontend", "project-backend"]
  
  mobile:
    name: "Mobile Development"
    repositories: ["project-mobile"]

settings:
  # Team standards
  parallel_jobs: 5
  default_sync_mode: "ff-only"
  confirm_destructive_operations: true
```

#### Export/Import Configuration
```bash
# Export repositories and groups (for sharing)
gman repo list --format yaml > team-repos.yml
gman repo group list --format yaml > team-groups.yml

# Import (manual merge required)
# Copy relevant sections to personal config.yml
```

### Multi-Environment Configuration

#### Work vs Personal
```bash
# Work configuration
export GMAN_CONFIG="$HOME/.config/gman/work-config.yml"
alias gman-work="GMAN_CONFIG=$HOME/.config/gman/work-config.yml gman"

# Personal configuration  
export GMAN_CONFIG="$HOME/.config/gman/personal-config.yml"
alias gman-personal="GMAN_CONFIG=$HOME/.config/gman/personal-config.yml gman"
```

#### Environment-Specific Settings
```yaml
# work-config.yml
repositories:
  api-service: /work/projects/api
  frontend-app: /work/projects/frontend

settings:
  parallel_jobs: 8                   # Powerful work machine
  default_sync_mode: "ff-only"       # Conservative for work
  confirm_destructive_operations: true

---
# personal-config.yml  
repositories:
  blog: /home/user/personal/blog
  experiments: /home/user/experiments

settings:
  parallel_jobs: 4                   # Personal laptop
  default_sync_mode: "rebase"        # Flexible for personal projects
  confirm_destructive_operations: false
```

## Troubleshooting Configuration

### Common Configuration Issues

#### Configuration File Not Found
**Error**: `Configuration file not found`
**Solution**:
```bash
# Check file exists
ls -la ~/.config/gman/config.yml

# Create directory if missing
mkdir -p ~/.config/gman

# Run setup to create configuration
gman tools setup
```

#### Invalid YAML Syntax
**Error**: `Error parsing configuration: yaml: ...`
**Solution**:
```bash
# Validate YAML syntax
python -c "import yaml; yaml.safe_load(open('~/.config/gman/config.yml'))"

# Or use online YAML validator
# Common issues: inconsistent indentation, missing quotes, invalid characters
```

#### Repository Path Issues
**Error**: `Repository path does not exist`
**Solution**:
```bash
# Verify paths in configuration
gman repo list --verbose

# Fix invalid paths
gman repo remove invalid-repo
gman repo add /correct/path valid-repo

# Update existing repository path (edit config manually)
```

#### Permission Issues
**Error**: `Permission denied accessing repository`
**Solution**:
```bash
# Check repository permissions
ls -la /path/to/repository

# Fix permissions if necessary
chmod -R u+rw /path/to/repository

# Check gman config directory permissions
ls -la ~/.config/gman/
chmod 644 ~/.config/gman/config.yml
```

### Configuration Validation

#### Manual Validation
```bash
# Test repository access
gman repo info repository-alias

# Test group configuration
gman repo group list group-name

# Test settings
gman work status --verbose
```

#### Automated Validation
```bash
# Validate all repositories
for repo in $(gman repo list --format json | jq -r '.[].alias'); do
    echo "Checking $repo..."
    gman repo info "$repo" || echo "❌ $repo has issues"
done
```

### Recovery Procedures

#### Reset Configuration
```bash
# Backup current config
cp ~/.config/gman/config.yml ~/.config/gman/config.yml.broken

# Remove broken configuration
rm ~/.config/gman/config.yml

# Run setup wizard
gman tools setup
```

#### Merge Configurations
```bash
# If you have partial backups or multiple configs to merge
# Manual editing required - there's no automatic merge tool yet

# Backup first
cp ~/.config/gman/config.yml ~/.config/gman/config.yml.backup

# Edit manually to merge sections
$EDITOR ~/.config/gman/config.yml
```

### Getting Help

For configuration issues:

1. **Validate syntax**: Use YAML validator
2. **Check paths**: Verify all repository paths exist
3. **Test permissions**: Ensure read/write access
4. **Review logs**: Use `gman --verbose` for detailed output
5. **Reset if necessary**: Use setup wizard to start fresh

For additional support:
- [USER_GUIDE.md](USER_GUIDE.md) - Complete user guide
- [COMMAND_REFERENCE.md](COMMAND_REFERENCE.md) - Command documentation
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - General troubleshooting
- [DEPLOYMENT.md](../DEPLOYMENT.md) - Installation and setup