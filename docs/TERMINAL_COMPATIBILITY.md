# gman TUI Dashboard - Terminal Compatibility Guide

## Overview

The `gman dashboard` command provides an interactive Terminal User Interface (TUI) that requires specific terminal capabilities. This guide explains how to troubleshoot and resolve terminal compatibility issues.

## Quick Start

### Basic Usage
```bash
gman dashboard          # Launch TUI dashboard
gman dashboard --debug  # Show terminal diagnostic information
gman dashboard --force  # Force TUI mode (bypass checks)
```

### Theme Options
```bash
gman dashboard --theme dark   # Dark theme (default)
gman dashboard --theme light  # Light theme
```

## Terminal Requirements

### Minimum Requirements
1. **TTY Support**: Terminal must support cursor movement and escape sequences
2. **TERM Environment**: `TERM` environment variable should be set to a supported terminal type
3. **Color Support**: Terminal should support ANSI colors (recommended)
4. **Minimum Size**: At least 80x24 characters for optimal experience

### Supported Terminal Types
- `xterm`, `xterm-256color`
- `screen`, `screen-256color`
- `tmux`, `tmux-256color`
- Most modern terminal emulators (iTerm2, Terminal.app, GNOME Terminal, etc.)

### Unsupported Environments
- `TERM=dumb` environments
- Non-interactive shells
- Redirected output environments
- Some CI/CD environments

## Troubleshooting

### Common Issues and Solutions

#### Issue: "stdout is not connected to a TTY"
```bash
# Problem: Running in a non-terminal environment
./gman dashboard
# Error: stdout is not connected to a TTY

# Solutions:
1. Run in a real terminal emulator
2. Use SSH with proper TTY allocation: ssh -t user@host
3. Use the --force flag: ./gman dashboard --force
```

#### Issue: "TERM environment variable not set"
```bash
# Check current TERM value
echo $TERM

# Set TERM if empty
export TERM=xterm-256color

# Or use specific terminal type
export TERM=screen-256color  # For screen/tmux
```

#### Issue: "cannot access terminal device (/dev/tty)"
```bash
# This typically happens in containerized or sandboxed environments
# Solutions:
1. Run with proper TTY allocation
2. Use --force flag to bypass the check
3. Ensure container has access to /dev/tty
```

### Diagnostic Commands

#### Basic Diagnostics
```bash
# Show detailed terminal information
gman dashboard --debug

# Check if terminal supports TUI
gman dashboard --debug 2>&1 | grep "TTY\|TERM\|supported"
```

#### Manual Checks
```bash
# Check if stdout is a TTY
test -t 1 && echo "TTY" || echo "Not TTY"

# Check TERM environment
echo "TERM: $TERM"

# Test terminal capabilities
tput colors    # Number of colors supported
tput lines     # Terminal height
tput cols      # Terminal width
```

## Advanced Usage

### SSH and Remote Access
```bash
# Proper SSH TTY allocation
ssh -t user@remote-host gman dashboard

# tmux/screen sessions
tmux new-session -d -s gman
tmux send-keys -t gman 'gman dashboard' Enter
tmux attach -t gman
```

### Docker/Container Environments
```bash
# Docker with TTY support
docker run -it --rm -v $(pwd):/workspace gman-container gman dashboard

# Docker Compose with TTY
# In docker-compose.yml:
services:
  gman:
    tty: true
    stdin_open: true
```

### CI/CD Environments
```bash
# Most CI environments don't support TUI
# Use CLI commands instead:
gman status
gman list
gman find file pattern
```

## Override Options

### Force Mode
```bash
# Bypass all terminal checks (use with caution)
gman dashboard --force
```
**Warning**: Force mode may cause display issues in incompatible terminals.

### Debug Mode
```bash
# Get detailed diagnostic information
gman dashboard --debug
```

### Fallback to CLI
```bash
# When TUI doesn't work, use CLI commands:
gman status --extended    # Detailed repository status
gman list                 # Repository list
gman recent              # Recently used repositories
gman find file pattern   # Search files
gman find commit pattern # Search commits
```

## Environment-Specific Guides

### macOS
- **Terminal.app**: Full support ✅
- **iTerm2**: Full support ✅
- **VS Code Terminal**: May require `--force` ⚠️

### Linux
- **GNOME Terminal**: Full support ✅
- **Konsole**: Full support ✅
- **xterm**: Full support ✅
- **SSH sessions**: Use `ssh -t` ✅

### Windows
- **Windows Terminal**: Full support ✅
- **PowerShell**: May require `--force` ⚠️
- **Git Bash**: Usually works ✅
- **WSL**: Full support ✅

### Remote/Cloud
- **SSH**: Use `ssh -t` for proper TTY allocation
- **tmux/screen**: Full support ✅
- **VS Code Remote**: May require `--force` ⚠️
- **Cloud IDE**: Environment-dependent ⚠️

## Best Practices

### For Users
1. **Test First**: Try `gman dashboard --debug` to check compatibility
2. **Use SSH Properly**: Always use `ssh -t` for remote access
3. **Terminal Choice**: Use modern terminal emulators when possible
4. **Fallback Ready**: Know CLI alternatives for non-TUI environments

### For Developers
1. **Environment Detection**: Check terminal capabilities before deployment
2. **CI/CD Scripts**: Use CLI commands in automated environments
3. **Documentation**: Include terminal requirements in setup docs
4. **Testing**: Test in various terminal environments

## Performance Tips

### Optimization
- Use terminals with GPU acceleration when available
- Ensure sufficient terminal buffer size
- Consider terminal font and rendering performance

### Resource Usage
- TUI dashboard uses minimal system resources
- Memory usage typically < 50MB
- CPU usage is minimal during idle state

## Getting Help

### When TUI Doesn't Work
1. Run diagnostics: `gman dashboard --debug`
2. Check environment variables: `env | grep TERM`
3. Test basic terminal features: `tput colors`
4. Use CLI alternatives: `gman status`, `gman list`
5. Try force mode: `gman dashboard --force`

### Common Workarounds
```bash
# For CI/CD environments
gman status --extended

# For restricted environments
gman dashboard --force

# For SSH without TTY
ssh -t user@host 'gman dashboard'

# For containers
docker run -it gman-container gman dashboard
```

---

**Note**: The TUI dashboard is designed to gracefully handle terminal compatibility issues and provide helpful error messages. When in doubt, use the `--debug` flag to understand what's happening with your terminal environment.