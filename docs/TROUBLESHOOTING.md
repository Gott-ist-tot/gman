# gman Troubleshooting Guide

This guide covers common issues and their solutions when using gman.

## Table of Contents

- [Directory Switching Issues](#directory-switching-issues)
- [Installation Problems](#installation-problems)
- [Configuration Issues](#configuration-issues)
- [Shell Integration Problems](#shell-integration-problems)
- [Performance Issues](#performance-issues)

## Directory Switching Issues

### Problem: `gman switch` doesn't change directory

**Symptoms:**
- Running `gman switch <repo>` shows output like `GMAN_CD:/path/to/repo`
- Current directory remains unchanged
- You see the target path but `pwd` shows the old directory

**Root Cause:**
This is a fundamental limitation of how operating systems handle processes. Go programs (like gman) run as child processes and cannot directly modify the parent shell's working directory due to process isolation - a security feature that prevents programs from interfering with each other.

**Technical Background:**
```
Shell (zsh/bash)
└── Spawns child process: gman
    ├── gman can only change its own working directory
    ├── gman terminates and returns to shell
    └── Shell's working directory is unchanged
```

The `GMAN_CD:` output is a special protocol that the shell wrapper function intercepts to execute the actual `cd` command in the shell context.

**Solution Steps:**

#### 1. Verify gman installation
```bash
# Check if gman binary exists and is executable
which gman
# Expected: /usr/local/bin/gman (or your installation path)

# If not found, gman is not in PATH
echo $PATH | tr ':' '\n' | grep gman
```

#### 2. Check shell function installation
```bash
# Check if gman shell function is loaded
type gman
# Expected: "gman is a function"
# Wrong: "gman is /path/to/gman" or "gman not found"
```

#### 3. Install shell integration

**For zsh** (add to `~/.zshrc`):
```bash
# gman Git Repository Manager - Shell Integration
export PATH="/usr/local/bin:$PATH"  # Adjust path as needed

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
            echo "✅ Switched to: $target_dir"
        else
            echo "❌ Error: Directory not found: $target_dir" >&2
            return 1
        fi
    else
        # For all other commands, just print the output
        echo "$output"
    fi

    return $exit_code
}

# Enable gman completion
if command -v gman &> /dev/null; then
    eval "$(gman completion zsh)"
fi
```

**For bash** (add to `~/.bashrc`):
```bash
# Same as above, but use bash completion:
eval "$(gman completion bash)"
```

#### 4. Reload shell configuration
```bash
# For zsh
source ~/.zshrc

# For bash
source ~/.bashrc

# Or restart your terminal
```

#### 5. Verify the fix
```bash
# Test 1: Check function is loaded
type gman
# Should show: "gman is a function"

# Test 2: List repositories
gman list

# Test 3: Test directory switching
pwd  # Note current directory
gman switch <your-repo-alias>
pwd  # Should show the repository path
```

### Advanced Diagnostics

#### Create a test script
```bash
#!/bin/zsh
echo "=== gman Diagnostic Test ==="
echo "Current directory: $(pwd)"
echo "gman type: $(type gman 2>&1)"
echo "gman path: $(which gman 2>&1)"
echo ""
echo "Testing gman switch..."
gman switch <your-repo> 2>&1
echo "Directory after switch: $(pwd)"
```

#### Manual test without function
```bash
# Test gman binary directly (should show GMAN_CD: output)
/usr/local/bin/gman switch <your-repo>

# Expected output: GMAN_CD:/path/to/your/repo
# If you see this, the binary works and shell function is the issue
```

## Installation Problems

### Problem: `gman: command not found`

**Solution 1: Install to system PATH**
```bash
# Copy binary to system location
sudo cp gman /usr/local/bin/
sudo chmod +x /usr/local/bin/gman
```

**Solution 2: Add to PATH**
```bash
# Add to shell configuration
echo 'export PATH="/path/to/gman/directory:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

**Solution 3: Create symlink**
```bash
# Create symlink in PATH location
sudo ln -s /path/to/gman /usr/local/bin/gman
```

### Problem: Permission denied

**Solution:**
```bash
# Make binary executable
chmod +x /path/to/gman

# If using sudo installation, check ownership
ls -la /usr/local/bin/gman
```

## Configuration Issues

### Problem: No repositories configured

**Symptoms:**
- `gman list` shows empty list
- `gman status` shows "no repositories configured"

**Solution:**
```bash
# Add your first repository
gman add /path/to/your/repo repo-alias

# Or add current directory
cd /path/to/your/repo
gman add . repo-alias
```

### Problem: Configuration file not found

**Solution:**
```bash
# Check configuration location
ls -la ~/.config/gman/config.yml

# If directory doesn't exist, create it
mkdir -p ~/.config/gman

# Add a repository to create configuration
gman add /path/to/repo test-repo
```

### Problem: Corrupted configuration

**Symptoms:**
- YAML parsing errors
- gman commands fail with configuration errors

**Solution:**
```bash
# Backup existing config
cp ~/.config/gman/config.yml ~/.config/gman/config.yml.backup

# View configuration
cat ~/.config/gman/config.yml

# Fix YAML syntax or recreate
# Expected format:
# repositories:
#   repo-name: /path/to/repo
# settings:
#   parallel_jobs: 5
```

## Shell Integration Problems

### Problem: Shell function conflicts

**Symptoms:**
- gman function doesn't work as expected
- Conflicts with other tools

**Solution:**
```bash
# Check for conflicting functions
declare -f gman

# Unload conflicting function
unset -f gman

# Reload correct function
source ~/.zshrc
```

### Problem: Different shells (bash vs zsh)

**Check your shell:**
```bash
echo $SHELL
# /bin/zsh or /bin/bash
```

**Use correct configuration file:**
- **zsh**: `~/.zshrc`
- **bash**: `~/.bashrc`
- **fish**: `~/.config/fish/config.fish` (different syntax)

### Problem: Oh My Zsh or other framework conflicts

**Solution:**
```bash
# Add gman configuration after framework loading
# In ~/.zshrc, place gman setup after:
source $ZSH/oh-my-zsh.sh

# Then add gman configuration
# ... gman setup code ...
```

## Performance Issues

### Problem: Slow repository operations

**Diagnosis:**
```bash
# Test with single repository
time gman status

# Check repository size and complexity
cd /path/to/repo
git log --oneline | wc -l  # Count commits
find . -name "*.git" | wc -l  # Count git objects
```

**Solutions:**
```bash
# Reduce parallel jobs in configuration
# Edit ~/.config/gman/config.yml:
settings:
  parallel_jobs: 3  # Reduce from default 5

# Clean up repositories
cd /path/to/repo
git gc --prune=now
git repack -ad
```

### Problem: Network timeouts

**Solution:**
```bash
# Increase git timeout
git config --global http.lowSpeedLimit 0
git config --global http.lowSpeedTime 999999

# Use SSH instead of HTTPS
git remote set-url origin git@github.com:user/repo.git
```

## Getting Help

### Enable debug mode
```bash
# Set debug environment variable
export GMAN_DEBUG=1
gman <command>
```

### Collect diagnostic information
```bash
# System information
echo "OS: $(uname -a)"
echo "Shell: $SHELL"
echo "PATH: $PATH"
echo "gman location: $(which gman)"
echo "gman type: $(type gman)"

# Configuration
echo "Config location: ~/.config/gman/config.yml"
ls -la ~/.config/gman/
head -20 ~/.config/gman/config.yml

# Shell configuration
grep -n "gman" ~/.zshrc ~/.bashrc 2>/dev/null
```

### Test with minimal configuration
```bash
# Create test configuration
mkdir -p /tmp/gman-test
export GMAN_CONFIG=/tmp/gman-test/config.yml

# Add test repository
gman add /tmp test-repo

# Test functionality
gman list
gman switch test-repo
```

## Reporting Issues

When reporting issues, please include:

1. **Operating System**: `uname -a`
2. **Shell**: `echo $SHELL`
3. **gman version**: `gman help` (version info)
4. **Configuration**: Content of `~/.config/gman/config.yml`
5. **Shell config**: Relevant parts of `~/.zshrc` or `~/.bashrc`
6. **Error messages**: Complete error output
7. **Steps to reproduce**: Exact commands that cause the issue

## Quick Reference

### Essential Commands for Troubleshooting
```bash
# Check installation
which gman && echo "✅ gman in PATH" || echo "❌ gman not found"

# Check shell function
type gman | grep -q "function" && echo "✅ Shell function loaded" || echo "❌ No shell function"

# Test basic functionality
gman list

# Test directory switching
gman switch <repo> && echo "✅ Switch works" || echo "❌ Switch failed"

# Check configuration
cat ~/.config/gman/config.yml

# Reload shell configuration
source ~/.zshrc  # or ~/.bashrc
```