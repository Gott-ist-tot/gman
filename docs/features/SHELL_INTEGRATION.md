# gman Shell Integration Guide

Shell integration is a critical component of gman that enables seamless directory switching and enhanced command-line experience. This guide covers setup, functionality, and troubleshooting for shell integration across different shells.

## Overview

### Why Shell Integration is Required

Due to operating system process isolation, child processes (like gman) cannot modify the parent shell's working directory. Shell integration provides a wrapper function that:

1. **Intercepts special output** from gman commands
2. **Executes directory changes** in the shell context  
3. **Preserves command output** and exit codes
4. **Enables tab completion** for commands and repository names

### Technical Background

```
┌─────────────────┐
│ Shell (zsh/bash)│
│ Working Dir: /A │
└─────────┬───────┘
          │ Spawns child process
          ▼
┌─────────────────┐
│ gman binary     │    ╔══════════════════════════════════╗
│ Working Dir: /A │───▶║ Output: GMAN_CD:/path/to/repo    ║
│ (can't change   │    ║ Shell wrapper intercepts this    ║
│  parent's dir)  │    ║ and executes: cd /path/to/repo   ║
└─────────────────┘    ╚══════════════════════════════════╝
```

## Installation

### Automatic Installation

The recommended installation scripts handle shell integration automatically:

```bash
# Complete setup including shell integration
./scripts/install.sh

# Or quick setup
./scripts/quick-setup.sh
```

### Manual Installation

#### Step 1: Copy Integration Script

```bash
# Create gman config directory
mkdir -p ~/.config/gman

# Copy shell integration script
cp scripts/shell-integration.sh ~/.config/gman/
```

#### Step 2: Add to Shell Configuration

**For Zsh** (add to `~/.zshrc`):
```bash
# gman Shell Integration
if [ -f "$HOME/.config/gman/shell-integration.sh" ]; then
    source "$HOME/.config/gman/shell-integration.sh"
fi
```

**For Bash** (add to `~/.bashrc`):
```bash
# gman Shell Integration  
if [ -f "$HOME/.config/gman/shell-integration.sh" ]; then
    source "$HOME/.config/gman/shell-integration.sh"
fi
```

**For Fish** (add to `~/.config/fish/config.fish`):
```fish
# gman Shell Integration (Fish adaptation required)
if test -f "$HOME/.config/gman/shell-integration.sh"
    # Fish requires custom wrapper function
    source "$HOME/.config/gman/shell-integration.fish"
end
```

#### Step 3: Reload Shell Configuration

```bash
# For Zsh
source ~/.zshrc

# For Bash  
source ~/.bashrc

# For Fish
source ~/.config/fish/config.fish

# Or restart your terminal
```

## Shell Integration Components

### Core Wrapper Function

The shell integration provides a `gman()` function that wraps the actual gman binary:

```bash
gman() {
    local output
    local exit_code

    # Call the actual gman binary and capture output
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
        # For all other commands, print the output normally
        echo "$output"
    fi

    return $exit_code
}
```

### Tab Completion

Shell integration includes command and repository name completion:

```bash
# Enable completion for your shell
if command -v gman &> /dev/null; then
    eval "$(gman completion zsh)"  # For Zsh
    eval "$(gman completion bash)" # For Bash
fi
```

### PATH Management

Ensures gman binary is accessible:

```bash
# Add gman to PATH (adjust path to your installation)
export PATH="/usr/local/bin:$PATH"  # System installation
# OR
export PATH="/path/to/gman/directory:$PATH"  # Custom installation
```

## Features

### Directory Switching

The primary feature enabling actual directory changes:

```bash
# Interactive repository selection
gman switch
# Shell integration handles: cd /path/to/selected/repo

# Direct repository switching  
gman switch my-project
# Shell integration handles: cd /path/to/my-project

# Fuzzy matching
gman switch cli
# Shell integration handles: cd /path/to/cli-tool
```

### Warning Preservation

The wrapper preserves warnings while still switching directories:

```bash
# Example: Repository with warnings
gman switch old-project
# Output: 
# Warning: Repository has uncommitted changes
# ✅ Switched to: /path/to/old-project
```

### Command Pass-through

Non-switching commands work normally:

```bash
# These commands are passed through unchanged
gman repo list
gman work status  
gman tools find file config.yaml
```

### Exit Code Preservation

The wrapper maintains proper exit codes for scripting:

```bash
# Exit code from gman binary is preserved
gman switch non-existent-repo
echo $?  # Shows actual error code from gman
```

## Shell-Specific Configurations

### Zsh Configuration

Complete Zsh integration in `~/.zshrc`:

```bash
# gman Git Repository Manager - Shell Integration
export PATH="/usr/local/bin:$PATH"

# Main wrapper function
gman() {
    local output
    local exit_code

    output=$(command gman "$@" 2>&1)
    exit_code=$?

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
        echo "$output"
    fi

    return $exit_code
}

# Enable completion
if command -v gman &> /dev/null; then
    eval "$(gman completion zsh)"
fi

# Optional: Add gman alias shortcuts
alias gs='gman switch'
alias gst='gman work status'
alias gsy='gman work sync'
```

### Bash Configuration

Complete Bash integration in `~/.bashrc`:

```bash
# gman Git Repository Manager - Shell Integration
export PATH="/usr/local/bin:$PATH"

# Main wrapper function
gman() {
    local output
    local exit_code

    output=$(command gman "$@" 2>&1)
    exit_code=$?

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
        echo "$output"
    fi

    return $exit_code
}

# Enable completion
if command -v gman &> /dev/null; then
    eval "$(gman completion bash)"
fi
```

### Fish Shell Configuration

Fish requires a different approach in `~/.config/fish/config.fish`:

```fish
# gman Git Repository Manager - Fish Integration

# Add gman to PATH
set -gx PATH /usr/local/bin $PATH

# Fish wrapper function
function gman
    set output (command gman $argv 2>&1)
    set exit_code $status
    
    if string match -q "GMAN_CD:*" $output
        set target_dir (string replace "GMAN_CD:" "" $output)
        if test -d $target_dir
            cd $target_dir
            echo "✅ Switched to: $target_dir"
        else
            echo "❌ Error: Directory not found: $target_dir" >&2
            return 1
        end
    else
        echo $output
    end
    
    return $exit_code
end

# Enable completion
if command -v gman > /dev/null
    gman completion fish | source
end
```

## Verification

### Quick Tests

```bash
# 1. Verify wrapper function is loaded
type gman
# Expected: "gman is a function"
# Wrong: "gman is /path/to/gman" or "gman not found"

# 2. Test basic functionality
gman repo list
# Should show your repositories

# 3. Test directory switching
pwd  # Note current directory
gman switch <your-repo-alias>
pwd  # Should show repository path
```

### Comprehensive Verification

```bash
# Create a test script for comprehensive verification
cat > test_gman_integration.sh << 'EOF'
#!/bin/bash
echo "=== gman Shell Integration Test ==="
echo "Current directory: $(pwd)"
echo "gman type: $(type gman 2>&1)"
echo "gman binary path: $(which gman 2>&1)"
echo ""

echo "Testing gman list..."
gman repo list
echo ""

echo "Testing directory switching..."
ORIGINAL_DIR=$(pwd)
gman switch <test-repo-alias>
NEW_DIR=$(pwd)
echo "Original: $ORIGINAL_DIR"
echo "New: $NEW_DIR"

if [ "$ORIGINAL_DIR" != "$NEW_DIR" ]; then
    echo "✅ Directory switching works!"
else
    echo "❌ Directory switching failed!"
fi
EOF

chmod +x test_gman_integration.sh
./test_gman_integration.sh
```

## Troubleshooting

### Common Issues

#### Issue: `gman switch` shows `GMAN_CD:` but doesn't change directory

**Diagnosis:**
```bash
type gman
# If shows: "gman is /path/to/gman" instead of "gman is a function"
```

**Solution:**
```bash
# Shell integration not loaded - add to shell config
echo 'source ~/.config/gman/shell-integration.sh' >> ~/.zshrc
source ~/.zshrc
```

#### Issue: "command not found: gman"

**Diagnosis:**
```bash
which gman
# If shows: "gman not found"
```

**Solution:**
```bash
# gman not in PATH
export PATH="/usr/local/bin:$PATH"  # Or your installation path
# Add this to your shell config permanently
```

#### Issue: Shell function conflicts

**Diagnosis:**
```bash
# Multiple gman definitions or conflicts
declare -f gman  # Shows function definition
```

**Solution:**
```bash
# Remove conflicting definitions
unset -f gman
# Re-source integration
source ~/.config/gman/shell-integration.sh
```

#### Issue: Completion not working

**Diagnosis:**
```bash
# Test completion manually
gman <TAB><TAB>
# Should show available commands
```

**Solution:**
```bash
# Reload completion
eval "$(gman completion zsh)"  # or bash
```

### Advanced Debugging

#### Debug Shell Integration

```bash
# Enable shell debugging
set -x  # For bash/zsh
gman switch <repo>
set +x

# Or use function debugging
gman() {
    echo "DEBUG: Called with args: $@"
    # ... rest of function
}
```

#### Check Binary Output

```bash
# Test gman binary directly (bypass shell function)
command gman switch <repo>
# Should output: GMAN_CD:/path/to/repo
```

#### Environment Verification

```bash
# Check environment variables
echo "SHELL: $SHELL"
echo "PATH: $PATH"
echo "PWD: $PWD"

# Check shell-specific variables
echo "ZSH_VERSION: $ZSH_VERSION"  # For Zsh
echo "BASH_VERSION: $BASH_VERSION"  # For Bash
```

## Best Practices

### Shell Configuration

1. **Load Early**: Place gman integration early in shell config
2. **Conditional Loading**: Use `if` statements to prevent errors
3. **Path Management**: Ensure gman is in PATH before integration
4. **Backup Configs**: Keep backup of shell configurations

### Development Workflow

1. **Test Integration**: Verify after any shell config changes
2. **Multiple Shells**: Test on all shells you use
3. **Script Compatibility**: Ensure scripts work with and without integration
4. **Update Management**: Keep integration script updated

### Performance

1. **Lazy Loading**: Load completion only when needed
2. **Cache Results**: Some shells support caching for faster startup
3. **Minimal Dependencies**: Keep integration script lightweight
4. **Profile Startup**: Monitor shell startup time impact

## Integration with Other Tools

### Terminal Multiplexers

```bash
# tmux integration
# Add to ~/.tmux.conf for automatic session naming
set-option -g automatic-rename on
set-option -g automatic-rename-format '#{b:pane_current_path}'

# screen integration  
# Works automatically with proper shell integration
```

### IDE Integration

```bash
# VS Code integrated terminal
# Shell integration works in VS Code terminal

# JetBrains IDEs
# May require manual PATH configuration in IDE settings
```

### SSH and Remote Systems

```bash
# SSH with TTY allocation
ssh -t user@host

# Remote shell integration
# Copy integration script to remote system
scp scripts/shell-integration.sh user@host:~/.config/gman/
```

For troubleshooting directory switching issues, see [Troubleshooting Guide](../troubleshooting/TROUBLESHOOTING.md).