# gman - Git Repository Manager

A powerful CLI tool for managing multiple Git repositories efficiently. Built with Go and designed for developers who work with multiple repositories simultaneously.

## Features

- 🚀 **Fast and Concurrent**: Built with Go, supports parallel operations across repositories
- 📊 **Visual Status**: Colorized table display showing repository status at a glance
- 🔄 **Quick Switching**: Instantly switch between repository directories
- 🔗 **Shell Integration**: Seamless integration with bash/zsh for directory changes
- ⚡ **Batch Operations**: Sync all repositories with one command
- 🎯 **Auto-completion**: Tab completion for commands and repository aliases
- 🛠 **Configurable**: YAML-based configuration with sensible defaults

## Installation

### Quick Install (if binary is available)

```bash
# Make the install script executable and run it
chmod +x scripts/install.sh
./scripts/install.sh
```

### Manual Installation

1. **Build from source**:
   ```bash
   git clone <repository-url>
   cd gman
   go build -o gman
   ```

2. **Move to PATH**:
   ```bash
   sudo mv gman /usr/local/bin/
   ```

3. **Setup shell integration** (Required for `gman switch`):
   
   **⚠️ Important**: The shell integration is **required** for `gman switch` to work properly. Without it, `gman switch` will only output the target path but won't actually change your current directory.
   
   Add this to your `~/.bashrc` or `~/.zshrc`:
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

1. **Add your first repository**:
   ```bash
   gman add /path/to/your/repo my-project
   # or add current directory
   gman add . current-project
   ```

2. **Check status**:
   ```bash
   gman status
   ```

3. **Switch to a repository**:
   ```bash
   gman switch my-project
   ```

4. **Sync all repositories**:
   ```bash
   gman sync
   ```

## Commands

### Core Commands

- **`gman status`** - Show status of all repositories
- **`gman switch <alias>`** - Switch to repository directory
- **`gman list`** - List all configured repositories
- **`gman sync`** - Synchronize all repositories with remotes

### Repository Management

- **`gman add [path] [alias]`** - Add a repository
- **`gman remove <alias>`** - Remove a repository from configuration

### Utility Commands

- **`gman completion [bash|zsh|fish|powershell]`** - Generate completion scripts
- **`gman help`** - Show help information

## Examples

### Adding Repositories

```bash
# Add current directory with auto-generated alias
gman add

# Add specific path with auto-generated alias
gman add /home/user/projects/webapp

# Add with custom alias
gman add /home/user/projects/api backend-api
gman add . frontend-app
```

### Viewing Status

```bash
$ gman status
Alias       Branch   Workspace      Sync Status     
─────────── ──────── ─────────────── ────────────────
* backend   main     🟢 CLEAN       ✅ UP-TO-DATE   
  frontend  develop  🔴 DIRTY       ↑ 2 AHEAD      
  infra     main     🟡 STASHED     ↓ 1 BEHIND     
```

### Syncing Repositories

```bash
# Default sync (fast-forward only)
gman sync

# Sync with rebase
gman sync --rebase

# Sync with autostash
gman sync --autostash
```

## Configuration

gman uses a YAML configuration file located at `~/.config/gman/config.yml`:

```yaml
repositories:
  backend-api: /home/user/projects/backend
  frontend-app: /home/user/projects/frontend
  infrastructure: /home/user/projects/infra

settings:
  parallel_jobs: 5
  show_last_commit: true
  default_sync_mode: "ff-only"

command_aliases:
  update: sync --rebase --autostash
  clean: run "git gc --prune=now"
```

### Configuration Options

- **`parallel_jobs`**: Number of concurrent operations (default: 5)
- **`show_last_commit`**: Show last commit in status (default: true)
- **`default_sync_mode`**: Default sync mode - "ff-only", "rebase", or "autostash" (default: "ff-only")

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

> 📋 **For detailed troubleshooting guide, see [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)**

### `gman switch` doesn't change directory

**Symptom**: When you run `gman switch <repo>`, you see output like `GMAN_CD:/path/to/repo` but your current directory doesn't change.

**Cause**: This happens when the shell integration wrapper function is not properly installed.

**Technical Background**: Go programs (like gman) run as child processes and cannot directly change the parent shell's working directory due to process isolation. The shell wrapper function is required to interpret the `GMAN_CD:` output and execute the `cd` command in the shell.

**Solution**:

1. **Check if gman is in PATH**:
   ```bash
   which gman
   # Should show path to gman binary
   ```

2. **Check if shell function is loaded**:
   ```bash
   type gman
   # Should show "gman is a function" (not "gman is /path/to/gman")
   ```

3. **If shell function is missing**, add this to your `~/.zshrc` or `~/.bashrc`:
   ```bash
   # gman wrapper function
   gman() {
       local output
       local exit_code
       output=$(command gman "$@" 2>&1)
       exit_code=$?
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
           echo "$output"
       fi
       return $exit_code
   }
   ```

4. **Reload your shell configuration**:
   ```bash
   source ~/.zshrc  # or ~/.bashrc
   ```

5. **Test the fix**:
   ```bash
   gman switch <your-repo>
   pwd  # Should show the repository path
   ```

### gman command not found

**Solution**: Make sure gman binary is in your PATH:

```bash
# Option 1: Install to system location
sudo cp gman /usr/local/bin/

# Option 2: Add to PATH in shell config
echo 'export PATH="/path/to/gman/directory:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### Configuration not found

**Solution**: Initialize configuration by adding your first repository:

```bash
gman add /path/to/your/repo repo-name
```

The configuration file will be created at `~/.config/gman/config.yml`.

### Permission denied errors

**Solution**: Ensure gman binary has execute permissions:

```bash
chmod +x /path/to/gman
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- Uses [Viper](https://github.com/spf13/viper) for configuration management
- Colorized output with [color](https://github.com/fatih/color)