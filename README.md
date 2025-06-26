# gman - Git Repository Manager

A powerful CLI tool for managing multiple Git repositories efficiently. Built with Go and designed for developers who work with multiple repositories simultaneously.

## Features

### Core Capabilities
- 🚀 **Fast and Concurrent**: Built with Go, supports parallel operations across repositories
- 📊 **Visual Status**: Colorized table display showing repository status at a glance
- 🔄 **Quick Switching**: Instantly switch between repository directories
- 🔗 **Shell Integration**: Seamless integration with bash/zsh for directory changes
- ⚡ **Batch Operations**: Sync all repositories with one command
- 🎯 **Auto-completion**: Tab completion for commands and repository aliases
- 🛠 **Configurable**: YAML-based configuration with sensible defaults

### Enhanced Search System (Phase 2)
- 🔍 **Lightning-Fast File Search**: Real-time file discovery using `fd` across all repositories
- 🔎 **Powerful Content Search**: Regex-powered text search using `ripgrep` with instant results
- 🎛️ **Interactive Selection**: Fuzzy finder (`fzf`) integration for intuitive selection
- 📱 **TUI Dashboard**: Modern terminal interface with live search and repository management
- 🏷️ **Group-Based Operations**: Organize repositories into groups for targeted operations

## Installation

### Automated Installation (Recommended)

For the complete setup with external dependencies and shell integration:

```bash
# Clone and build
git clone <repository-url>
cd gman
go build -o gman .

# Run automated installation (includes fd, rg, fzf)
./scripts/install.sh
```

### Quick Manual Installation

```bash
# Build and install binary only
go build -o gman .
sudo mv gman /usr/local/bin/

# Setup shell integration
source scripts/shell-integration.sh
```

> 📋 **For detailed installation instructions, troubleshooting, and platform-specific guidance, see [DEPLOYMENT.md](DEPLOYMENT.md)**

### External Dependencies

gman's enhanced search features require external tools:

- **fd**: Lightning-fast file search
- **ripgrep (rg)**: Powerful content search  
- **fzf**: Interactive fuzzy finder

Install automatically:
```bash
./scripts/setup-dependencies.sh
```

Or install manually based on your platform - see [DEPLOYMENT.md](DEPLOYMENT.md) for details.

### Shell Integration Setup

**⚠️ Critical**: Shell integration is **required** for `gman switch` to work properly. Without it, `gman switch` will only output the target path but won't actually change your current directory.

**Automatic Setup**: The installation script handles this automatically, or you can source the integration manually:

```bash
# Add to ~/.bashrc or ~/.zshrc
source ~/.config/gman/shell-integration.sh
```

**Manual Setup**: Add this to your `~/.bashrc` or `~/.zshrc`:
   ```bash
   # gman Git Repository Manager - Shell Integration
   # Add gman to PATH (adjust path to your gman binary)
   export PATH="/usr/local/bin:$PATH"  # if installed via sudo mv
   # OR for local installation:
   # export PATH="/path/to/gman/directory:$PATH"
   
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
       eval "$(gman completion bash)"  # for bash
       eval "$(gman completion zsh)"   # for zsh
   fi
   ```

4. **Restart your shell** or run `source ~/.bashrc` (or `~/.zshrc`)

5. **Verify installation**:
   ```bash
   # Check if gman is available
   which gman
   
   # Test the shell integration
   gman list
   
   # Test directory switching (should actually change directory)
   gman switch your-repo-alias
   pwd  # Should show the repository path
   ```

## Quick Start

1. **Setup gman with the interactive wizard**:
   ```bash
   gman tools setup
   ```

2. **Add your first repository**:
   ```bash
   gman repo add /path/to/your/repo my-project
   # or add current directory
   gman repo add . current-project
   ```

3. **Check status across all repositories**:
   ```bash
   gman work status
   # or extended view with details
   gman work status --extended
   ```

4. **Search for files across repositories**:
   ```bash
   gman tools find file config.yaml
   gman tools find content "TODO"
   ```

5. **Switch to a repository**:
   ```bash
   gman switch my-project
   ```

6. **Launch the interactive dashboard**:
   ```bash
   gman tools dashboard
   ```

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