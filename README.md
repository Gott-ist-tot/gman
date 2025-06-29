# gman - Git Repository Manager

A modern, production-ready CLI tool for efficient multi-repository management. Built with Go and designed for developers who work with multiple Git repositories simultaneously.

## ✨ Key Features

- 🚀 **Lightning-Fast Operations**: Concurrent Git operations across multiple repositories
- 📊 **Real-Time Status Monitoring**: Visual status indicators with live updates
- 🔄 **Seamless Directory Switching**: Instant navigation between repositories
- 🔍 **Powerful Search System**: Real-time file and content search across all repositories
- 📱 **Interactive TUI Dashboard**: Modern terminal interface for visual management
- 🎯 **Smart Repository Groups**: Organize and operate on related repositories
- 🔗 **Deep Shell Integration**: Native bash/zsh integration with tab completion
- ⚡ **Batch Git Operations**: Commit, push, sync across multiple repositories
- 🛠 **Flexible Configuration**: YAML-based configuration with intelligent defaults

## 🚀 Quick Installation

```bash
# 1. Clone and build
git clone <repository-url>
cd gman
go build -o gman .

# 2. Run automated setup (includes dependencies + shell integration)
./scripts/quick-setup.sh

# 3. Start using gman
gman tools setup  # Interactive setup wizard
```

> 📋 **For detailed installation instructions, platform-specific guidance, and troubleshooting, see [Installation Guide](docs/getting-started/INSTALLATION.md)**

### Requirements

- **Go 1.19+** (for building)
- **Git** (for repository operations)
- **External Tools** (optional but recommended):
  - `fd` - Lightning-fast file search
  - `rg` (ripgrep) - Powerful content search
  - `fzf` - Interactive fuzzy finder

**⚠️ Important**: Shell integration is required for `gman switch` to work. The automated setup handles this, or see the [Installation Guide](docs/getting-started/INSTALLATION.md) for manual setup.

## ⚡ Quick Start

```bash
# 1. Interactive setup wizard
gman tools setup

# 2. Add repositories
gman repo add /path/to/project my-project
gman repo add . current-project

# 3. Check status across all repositories
gman work status --extended

# 4. Create repository groups
gman repo group create webdev frontend-app backend-api

# 5. Search across repositories
gman tools find file config.yaml
gman tools find content "TODO"

# 6. Switch to a repository (changes directory)
gman switch my-project

# 7. Launch interactive dashboard
gman tools dashboard
```

> 🎯 **New to gman?** Start with the [Quick Start Tutorial](docs/getting-started/QUICK_START.md) for a guided walkthrough!

## Commands

gman uses a modern, organized command structure with logical grouping:

### Repository Management (`gman repo` or `gman r`)
- **`gman repo add [path] [alias]`** - Add a repository
- **`gman repo remove <alias>`** - Remove a repository from configuration
- **`gman repo list`** - List all configured repositories
- **`gman repo group create <name> <repos...>`** - Create repository groups

### Git Workflow (`gman work` or `gman w`)
- **`gman work status [--extended]`** - Show status of all repositories
- **`gman work sync [--group <name>]`** - Synchronize repositories with remotes
- **`gman work commit -m "message" [--add]`** - Commit changes across repositories
- **`gman work push [--group <name>]`** - Push changes to remotes

### Enhanced Search & Tools (`gman tools` or `gman t`)
- **`gman tools find file <pattern>`** - Lightning-fast file search across repos
- **`gman tools find content <pattern>`** - Powerful content search with regex
- **`gman tools dashboard`** - Launch interactive TUI dashboard
- **`gman tools setup`** - Interactive setup wizard

### Quick Operations
- **`gman switch <alias>`** - Switch to repository directory (requires shell integration)
- **`gman recent`** - Show recently accessed repositories

## Examples

### Adding Repositories

```bash
# Add current directory with auto-generated alias
gman repo add

# Add specific path with auto-generated alias  
gman repo add /home/user/projects/webapp

# Add with custom alias
gman repo add /home/user/projects/api backend-api
gman repo add . frontend-app

# Create repository groups
gman repo group create webdev frontend-app backend-api
gman repo group create tools cli-utils monitoring
```

### Viewing Status

```bash
$ gman work status
Alias       Branch   Workspace      Sync Status     
─────────── ──────── ─────────────── ────────────────
* backend   main     🟢 CLEAN       ✅ UP-TO-DATE   
  frontend  develop  🔴 DIRTY       ↑ 2 AHEAD      
  infra     main     🟡 STASHED     ↓ 1 BEHIND     

# Extended view with file counts and commit times
$ gman work status --extended
```

### Enhanced Search Examples

```bash
# Search for files across all repositories
gman tools find file config.yaml
gman tools find file "*.go" --group backend

# Search content with regex patterns
gman tools find content "TODO.*urgent"
gman tools find content "func.*Error" --group frontend

# Interactive dashboard with live search
gman tools dashboard
```

### Batch Operations

```bash
# Sync repositories (with modern options)
gman work sync                    # All repositories
gman work sync --group webdev     # Specific group only
gman work sync --only-dirty       # Only repositories with changes

# Commit changes across repositories
gman work commit -m "Update dependencies" --add --group webdev

# Push changes
gman work push --group webdev
```

## Configuration

gman uses a YAML configuration file located at `~/.config/gman/config.yml`:

```yaml
repositories:
  backend-api: /home/user/projects/backend
  frontend-app: /home/user/projects/frontend
  infrastructure: /home/user/projects/infra

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

- **`parallel_jobs`**: Number of concurrent operations (default: 5)
- **`show_last_commit`**: Show last commit in status (default: true)
- **`default_sync_mode`**: Default sync mode - "ff-only", "rebase", or "autostash" (default: "ff-only")
- **`groups`**: Repository groups for organized batch operations
- **`recent`**: Automatically tracked recent repository access

## Enhanced Search System (Phase 2)

gman's Phase 2 revolution brings powerful search capabilities that transform multi-repository management:

### Real-Time File Search
- **Powered by `fd`**: Lightning-fast file discovery across all repositories
- **Smart filtering**: Exclude .git, node_modules, build directories automatically
- **Pattern matching**: Support for glob patterns and regex
- **Interactive selection**: fzf integration for intuitive file selection

```bash
gman tools find file config.yaml       # Find specific files
gman tools find file "*.go" --group backend  # Pattern search in groups
```

### Content Search
- **Powered by `ripgrep`**: Ultra-fast text search with regex support
- **Context-aware**: Shows surrounding lines for better understanding
- **Multi-repository**: Search across all repositories simultaneously
- **Respect .gitignore**: Honors repository ignore patterns

```bash
gman tools find content "TODO.*urgent"     # Regex patterns
gman tools find content "import.*axios"    # Find imports
```

### Interactive Dashboard
- **Modern TUI**: Built with Bubble Tea for responsive terminal interface
- **Live search**: Real-time file and content search with preview
- **Repository management**: Status monitoring and quick operations
- **Keyboard navigation**: Vim-style shortcuts and intuitive controls

```bash
gman tools dashboard                    # Launch TUI
```

### Dependencies

The search system requires external tools:
- **fd**: File search engine
- **ripgrep**: Content search engine  
- **fzf**: Interactive fuzzy finder

Install automatically: `./scripts/setup-dependencies.sh`

## Sync Modes

- **`ff-only`** (default): Only fast-forward merges, safest option
- **`rebase`**: Use `git pull --rebase`
- **`autostash`**: Use `git pull --autostash`

## Status Indicators

### Workspace Status
- 🟢 **CLEAN**: No uncommitted changes
- 🔴 **DIRTY**: Uncommitted changes present  
- 🟡 **STASHED**: Stashed changes available

### Sync Status
- ✅ **UP-TO-DATE**: In sync with remote
- ↑ **X AHEAD**: Local commits not pushed
- ↓ **X BEHIND**: Remote commits not pulled
- 🔄 **X↑ Y↓**: Both ahead and behind

## Development

### Project Structure

```
gman/
├── cmd/                  # Command implementations
├── internal/
│   ├── config/          # Configuration management
│   ├── git/             # Git operations
│   └── display/         # Output formatting
├── pkg/types/           # Shared types
├── scripts/             # Installation and shell integration
└── main.go             # Entry point
```

### Building

```bash
go build -o gman
```

### Testing

```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Troubleshooting

> 📋 **For comprehensive troubleshooting, installation guides, and platform-specific instructions, see:**
> - **[DEPLOYMENT.md](DEPLOYMENT.md)** - Complete installation and setup guide
> - **[docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)** - Detailed troubleshooting guide

### Quick Fixes

#### `gman switch` doesn't change directory

**Symptom**: You see `GMAN_CD:/path/to/repo` but directory doesn't change.

**Solution**: Shell integration not installed. Run:
```bash
source ~/.config/gman/shell-integration.sh
# OR add to ~/.bashrc or ~/.zshrc
```

#### Search commands not working

**Symptom**: `gman tools find` commands fail or show warnings.

**Solution**: Install external dependencies:
```bash
./scripts/setup-dependencies.sh
# OR see DEPLOYMENT.md for manual installation
```

#### TUI dashboard not launching

**Symptom**: `gman tools dashboard` fails or displays incorrectly.

**Solution**: Check terminal compatibility:
```bash
gman tools dashboard --debug
# Use SSH with TTY: ssh -t user@host
```

#### Command not found

**Solution**: Ensure gman is in PATH:
```bash
# Check installation
which gman

# Install to system location
sudo mv gman /usr/local/bin/

# OR add to PATH
export PATH="/path/to/gman:$PATH"
```

#### Configuration issues

**Solution**: Initialize with setup wizard:
```bash
gman tools setup
# OR manually add first repository
gman repo add /path/to/repo repo-name
```

### Getting Help

1. **Installation Issues**: See [DEPLOYMENT.md](DEPLOYMENT.md)
2. **Command Usage**: Run `gman --help` or `gman <command> --help`
3. **Search Problems**: Run `./scripts/setup-dependencies.sh --verify-only`
4. **Advanced Troubleshooting**: See [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- Uses [Viper](https://github.com/spf13/viper) for configuration management
- Colorized output with [color](https://github.com/fatih/color)