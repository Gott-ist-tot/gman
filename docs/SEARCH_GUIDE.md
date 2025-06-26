# gman Enhanced Search Guide

Complete guide to gman's Phase 2 enhanced search system - the revolutionary approach to multi-repository discovery and navigation.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [External Dependencies](#external-dependencies)
- [File Search](#file-search)
- [Content Search](#content-search)
- [Interactive Dashboard Search](#interactive-dashboard-search)
- [Search Patterns and Techniques](#search-patterns-and-techniques)
- [Performance Optimization](#performance-optimization)
- [Integration with Workflows](#integration-with-workflows)
- [Troubleshooting Search Issues](#troubleshooting-search-issues)

## Overview

gman's Phase 2 search system represents a fundamental shift from traditional file indexing to real-time, high-performance search across multiple Git repositories. This system eliminates the need for index maintenance while providing lightning-fast discovery capabilities.

### Key Innovations

- **Real-time search**: No index building or maintenance required
- **External tool integration**: Leverages best-in-class command-line tools
- **Multi-repository scope**: Search across all configured repositories simultaneously
- **Interactive selection**: Seamless fzf integration for intuitive navigation
- **Context-aware results**: Respects .gitignore and excludes build artifacts

### Search Capabilities

| Search Type | Tool | Purpose | Speed |
|-------------|------|---------|-------|
| File Search | `fd` | Find files by name/pattern | Lightning-fast |
| Content Search | `ripgrep` | Search text within files | Ultra-fast |
| Interactive Selection | `fzf` | Navigate and select results | Instant |

## Architecture

### Real-Time Search Architecture

```
┌─────────────────┐    ┌──────────────┐    ┌─────────────────┐
│   gman Client   │───▶│ Search Layer │───▶│ External Tools  │
└─────────────────┘    └──────────────┘    └─────────────────┘
                              │                     │
                              ▼                     ▼
                    ┌──────────────────┐   ┌─────────────────┐
                    │ Repository       │   │ fd, rg, fzf     │
                    │ Coordination     │   │ Direct Execution│
                    └──────────────────┘   └─────────────────┘
```

### Benefits of Real-Time Architecture

1. **No Index Maintenance**: No SQLite databases to build, update, or corrupt
2. **Always Current**: Results reflect the current state of files
3. **Repository Awareness**: Respects .gitignore and Git boundaries
4. **Performance**: External tools are optimized for their specific tasks
5. **Scalability**: Handles repositories of any size efficiently

## External Dependencies

### fd (File Discovery)

**Purpose**: Lightning-fast file search replacement for `find`

**Features**:
- Parallel execution across directories
- Automatic .gitignore respect
- Smart filtering (excludes .git, node_modules, build directories)
- Unicode support
- Glob pattern matching

**Installation**:
```bash
# Automatic
./scripts/setup-dependencies.sh

# Manual (see DEPLOYMENT.md for all platforms)
brew install fd          # macOS
sudo apt install fd-find # Ubuntu/Debian
```

### ripgrep (Content Search)

**Purpose**: Ultra-fast text search with regex support

**Features**:
- Multi-threading for parallel searching
- Advanced regex engine
- Context line display
- File type filtering
- Automatic .gitignore respect

**Installation**:
```bash
# Automatic
./scripts/setup-dependencies.sh

# Manual
brew install ripgrep        # macOS
sudo apt install ripgrep    # Ubuntu/Debian
```

### fzf (Fuzzy Finder)

**Purpose**: Interactive selection interface

**Features**:
- Real-time fuzzy matching
- Preview window support
- Keyboard navigation
- Multi-selection capabilities
- Customizable interface

**Installation**:
```bash
# Automatic
./scripts/setup-dependencies.sh

# Manual
brew install fzf        # macOS
sudo apt install fzf    # Ubuntu/Debian
```

### Dependency Verification

```bash
# Check if all dependencies are installed
./scripts/setup-dependencies.sh --verify-only

# Or check manually
fd --version
rg --version
fzf --version
```

## File Search

### Basic File Search

Search for files by name across all repositories:

```bash
# Find specific files
gman tools find file config.yaml
gman tools find file package.json
gman tools find file Dockerfile

# Pattern matching with glob patterns
gman tools find file "*.go"
gman tools find file "*.js"
gman tools find file "test_*.py"
```

### Advanced File Search Patterns

```bash
# Complex glob patterns
gman tools find file "src/**/*.tsx"     # TypeScript React components
gman tools find file "**/test/**/*.js"  # Test files in any test directory
gman tools find file "*.{yml,yaml}"     # YAML files with either extension

# Case-insensitive search
gman tools find file readme --ignore-case
gman tools find file LICENSE --ignore-case

# File type filtering
gman tools find file --type f "config"  # Files only (not directories)
gman tools find file --type d "test"    # Directories only
```

### Group-Specific File Search

```bash
# Search within specific repository groups
gman tools find file "*.go" --group backend
gman tools find file "*.jsx" --group frontend
gman tools find file "docker-compose.yml" --group infrastructure

# Multiple groups (if supported)
gman tools find file "*.config.js" --group "frontend,backend"
```

### File Search Options

| Option | Description | Example |
|--------|-------------|---------|
| `--ignore-case, -i` | Case-insensitive search | `gman tools find file readme -i` |
| `--type TYPE` | File type filter | `gman tools find file config --type f` |
| `--exclude PATTERN` | Exclude pattern | `gman tools find file "*.js" --exclude "*test*"` |
| `--max-results N` | Limit results | `gman tools find file "*.go" --max-results 20` |
| `--group GROUP` | Search in group | `gman tools find file "*.py" --group backend` |

### File Search Performance

File search performance characteristics:

- **Small repositories** (< 1K files): Instant results
- **Medium repositories** (1K-10K files): Sub-second results
- **Large repositories** (10K+ files): 1-3 second results
- **Concurrent search**: Searches multiple repositories in parallel

## Content Search

### Basic Content Search

Search for text within files across repositories:

```bash
# Simple text search
gman tools find content "TODO"
gman tools find content "FIXME"
gman tools find content "console.log"

# Case-sensitive search
gman tools find content "API_KEY"
gman tools find content "DATABASE_URL"
```

### Regex Content Search

ripgrep supports powerful regex patterns:

```bash
# Function definitions
gman tools find content "func.*Error"
gman tools find content "function\s+\w+"
gman tools find content "def\s+\w+\("

# Import statements
gman tools find content "import.*axios"
gman tools find content "from.*react"
gman tools find content "require\(.*express"

# Configuration patterns
gman tools find content "process\.env\.\w+"
gman tools find content "config\.\w+\s*="
gman tools find content "\$\{.*\}"
```

### Content Search with Context

```bash
# Show surrounding lines for context
gman tools find content "error" --context 3      # 3 lines before/after
gman tools find content "import" --context 2     # 2 lines before/after
gman tools find content "TODO" --context 5       # 5 lines before/after

# Before/after context separately
gman tools find content "function" --before 2 --after 1
```

### File Type Filtering

```bash
# Search in specific file types
gman tools find content "console.log" --type js
gman tools find content "fmt.Println" --type go
gman tools find content "print(" --type py

# Multiple file types
gman tools find content "import" --type "js,ts,jsx,tsx"
gman tools find content "SELECT" --type "sql,py,js"

# Exclude specific types
gman tools find content "test" --type-not min.js
```

### Content Search Options

| Option | Description | Example |
|--------|-------------|---------|
| `--ignore-case, -i` | Case-insensitive search | `gman tools find content todo -i` |
| `--context N` | Lines of context | `gman tools find content error --context 3` |
| `--type TYPE` | File type filter | `gman tools find content log --type js` |
| `--max-results N` | Limit results | `gman tools find content TODO --max-results 50` |
| `--group GROUP` | Search in group | `gman tools find content error --group backend` |

### Content Search Examples

#### Development Patterns
```bash
# Find error handling
gman tools find content "catch\s*\(" --type js
gman tools find content "except:" --type py
gman tools find content "if.*err.*!=" --type go

# Find database queries
gman tools find content "SELECT.*FROM" --context 2
gman tools find content "INSERT INTO" --context 1
gman tools find content "UPDATE.*SET" --context 2

# Find configuration usage
gman tools find content "process\.env\." --type js
gman tools find content "os\.environ" --type py
gman tools find content "os\.Getenv" --type go
```

#### Security Patterns
```bash
# Find potential secrets
gman tools find content "(password|secret|key).*=" --ignore-case
gman tools find content "Bearer\s+[A-Za-z0-9]+"
gman tools find content "api[_-]?key.*=" --ignore-case

# Find authentication code
gman tools find content "auth.*token" --ignore-case
gman tools find content "jwt\." --ignore-case
gman tools find content "passport\." --type js
```

## Interactive Dashboard Search

The TUI dashboard provides a unified search interface combining file and content search with live preview.

### Launching Dashboard Search

```bash
# Launch dashboard
gman tools dashboard

# Launch with specific theme
gman tools dashboard --theme light
gman tools dashboard --theme dark
```

### Dashboard Search Features

#### Real-Time Search
- **Live results**: Search updates as you type
- **Multi-repository**: Results from all configured repositories
- **Preview pane**: Live file content preview
- **Context display**: Surrounding lines for content matches

#### Search Modes

| Mode | Trigger | Purpose |
|------|---------|---------|
| File Search | `f` or `/f` | Search for files by name |
| Content Search | `c` or `/c` | Search file contents |
| Combined Search | `/` | Search both files and content |

#### Navigation

| Key | Action |
|-----|--------|
| `/` | Start search |
| `Tab` | Switch between search modes |
| `↑↓` | Navigate results |
| `Enter` | Select result |
| `Esc` | Clear search |
| `Ctrl-C` | Exit dashboard |

### Dashboard Search Workflow

1. **Launch Dashboard**: `gman tools dashboard`
2. **Start Search**: Press `/` to enter search mode
3. **Type Pattern**: Enter search terms or patterns
4. **Navigate Results**: Use arrow keys to browse matches
5. **Preview Content**: Results show in preview pane
6. **Select Result**: Press Enter to open/navigate to file
7. **Refine Search**: Continue typing to narrow results

## Search Patterns and Techniques

### Effective Search Strategies

#### File Discovery Strategies
```bash
# Start broad, then narrow
gman tools find file config                    # All config files
gman tools find file "*.config.js"            # JavaScript config files
gman tools find file "webpack.config.js"      # Specific config file

# Use extensions effectively
gman tools find file "*.{json,yml,yaml}"      # Configuration files
gman tools find file "*.{js,ts,jsx,tsx}"      # JavaScript/TypeScript
gman tools find file "*.{py,pyc,pyo}"         # Python files
```

#### Content Discovery Strategies
```bash
# Progressive refinement
gman tools find content "error"               # All error mentions
gman tools find content "error.*handler"      # Error handling code
gman tools find content "function.*error.*handler" # Specific functions

# Context-aware searching
gman tools find content "import" --context 0  # Just import lines
gman tools find content "import" --context 3  # Imports with context
```

### Common Search Patterns

#### Finding Configuration
```bash
# Configuration files
gman tools find file "*config*"
gman tools find file "*.env*"
gman tools find file "*settings*"

# Configuration usage
gman tools find content "config\." --context 2
gman tools find content "\.env\." --context 1
```

#### Finding Dependencies
```bash
# Package files
gman tools find file "package.json"
gman tools find file "requirements.txt"
gman tools find file "go.mod"

# Import/require statements
gman tools find content "import.*from" --context 1
gman tools find content "require\(" --context 1
gman tools find content "import\s+\w+" --context 1
```

#### Finding Tests
```bash
# Test files
gman tools find file "*test*"
gman tools find file "*spec*"
gman tools find file "__test__"

# Test patterns
gman tools find content "describe\(" --context 2
gman tools find content "it\(" --context 1
gman tools find content "test.*function" --context 2
```

#### Finding Documentation
```bash
# Documentation files
gman tools find file "*.md" 
gman tools find file "README*"
gman tools find file "CHANGELOG*"

# Documentation patterns
gman tools find content "##\s+\w+" --context 1  # Markdown headers
gman tools find content "TODO:" --context 2      # TODO items
```

### Advanced Regex Patterns

#### JavaScript/TypeScript Patterns
```bash
# Function definitions
gman tools find content "function\s+\w+\s*\(" --type js
gman tools find content "const\s+\w+\s*=\s*\(" --type js
gman tools find content "=>\s*{" --type js

# React patterns
gman tools find content "useState\(" --type "jsx,tsx"
gman tools find content "useEffect\(" --type "jsx,tsx"
gman tools find content "export\s+default" --type "js,jsx,ts,tsx"
```

#### Python Patterns
```bash
# Function/class definitions
gman tools find content "def\s+\w+\(" --type py
gman tools find content "class\s+\w+:" --type py
gman tools find content "async\s+def" --type py

# Import patterns
gman tools find content "from\s+\w+\s+import" --type py
gman tools find content "import\s+\w+" --type py
```

#### Go Patterns
```bash
# Function definitions
gman tools find content "func\s+\w+\(" --type go
gman tools find content "func\s+\(\w+\)" --type go

# Error handling
gman tools find content "if.*err.*!=" --type go
gman tools find content "return.*err" --type go
```

## Performance Optimization

### Search Performance Tips

#### File Search Optimization
1. **Use specific patterns**: `"*.go"` instead of `"*go*"`
2. **Limit scope with groups**: `--group backend` for targeted searches
3. **Use appropriate file types**: `--type f` for files only
4. **Exclude unnecessary results**: `--exclude "node_modules"`

#### Content Search Optimization
1. **Use specific regex**: `"func\s+\w+"` instead of `"func.*"`
2. **Limit context**: Use `--context 1` instead of `--context 10`
3. **Filter by file type**: `--type js` for JavaScript-only searches
4. **Limit results**: `--max-results 50` for large searches

### Repository-Level Optimizations

#### .gitignore Effectiveness
Ensure `.gitignore` files are properly configured:
```gitignore
# Build directories
build/
dist/
target/

# Dependencies
node_modules/
vendor/

# IDE files
.vscode/
.idea/

# Temporary files
*.tmp
*.cache
```

#### Repository Structure
- **Shallow clones**: For read-only repositories
- **Regular cleanup**: `git gc` and `git prune`
- **Efficient .gitignore**: Exclude irrelevant files

### System-Level Optimizations

#### Hardware Considerations
- **SSD storage**: Dramatically improves search performance
- **Sufficient RAM**: Enables effective file system caching
- **Multi-core CPU**: Parallel search execution

#### Operating System Tuning
```bash
# Increase file descriptor limits (Linux/macOS)
ulimit -n 8192

# Enable file system caching
# (Usually enabled by default)
```

## Integration with Workflows

### IDE Integration

#### VS Code Integration
```bash
# Search and open in VS Code
gman tools find file "*.go" | xargs code

# Search content and open specific files
gman tools find content "TODO" --max-results 10
# Then manually open files of interest
```

#### Vim Integration
```bash
# Search and open in Vim
gman tools find file config.yaml | xargs vim

# Use with vim's grep integration
gman tools find content "error" --group backend
```

### Command Line Workflows

#### Combining with Other Tools
```bash
# Search and count results
gman tools find content "TODO" | wc -l

# Search and filter
gman tools find file "*.js" | grep -v test

# Search and process
gman tools find content "console.log" --type js | cut -d: -f1 | sort -u
```

#### Scripting Integration
```bash
#!/bin/bash
# Script to find and clean up console.log statements

echo "Finding console.log statements..."
gman tools find content "console\.log" --type js --max-results 100

echo "Found in the following files:"
gman tools find content "console\.log" --type js | cut -d: -f1 | sort -u
```

### Git Workflow Integration

#### Pre-commit Checks
```bash
# Check for debugging statements
gman tools find content "console\.log" --type js
gman tools find content "debugger" --type js
gman tools find content "print\(" --type py

# Check for TODO items
gman tools find content "TODO.*urgent" --ignore-case
```

#### Code Review Assistance
```bash
# Find recent changes by pattern
gman tools find content "// REVIEW" --context 3
gman tools find content "// FIXME" --context 2

# Find configuration changes
gman tools find file "*config*" --group "frontend,backend"
```

## Troubleshooting Search Issues

### Common Problems and Solutions

#### "Command not found" Errors

**Problem**: Search commands fail with "fd not found" or similar errors.

**Solution**:
```bash
# Check dependency installation
./scripts/setup-dependencies.sh --verify-only

# Install missing dependencies
./scripts/setup-dependencies.sh

# Manual verification
which fd rg fzf
```

#### Slow Search Performance

**Problem**: Search operations take longer than expected.

**Diagnostic Steps**:
```bash
# Check repository sizes
gman repo list --verbose

# Test external tools directly
time fd config.yaml
time rg "TODO" --stats

# Check system resources
top
df -h
```

**Solutions**:
- Ensure repositories have proper `.gitignore` files
- Use more specific search patterns
- Limit search scope with `--group` option
- Consider excluding large repositories from searches

#### No Results Found

**Problem**: Search returns no results when files/content should exist.

**Debugging**:
```bash
# Verify repository configuration
gman repo list

# Test in specific repository
cd /path/to/repo
fd config.yaml
rg "TODO"

# Check file permissions
ls -la /path/to/repo
```

**Common Causes**:
- Repository not properly added to gman
- Files excluded by `.gitignore`
- Permission issues
- Incorrect search patterns

#### fzf Interface Issues

**Problem**: Interactive selection doesn't work properly.

**Diagnostics**:
```bash
# Test fzf directly
echo -e "option1\noption2\noption3" | fzf

# Check terminal compatibility
gman tools dashboard --debug

# Test with different terminal
# Try in standard terminal, SSH session, or tmux
```

**Solutions**:
- Use SSH with TTY allocation: `ssh -t user@host`
- Update terminal software
- Check terminal environment variables
- Use `--force` flag for dashboard: `gman tools dashboard --force`

#### Group-Specific Search Issues

**Problem**: Search with `--group` option doesn't work as expected.

**Verification**:
```bash
# Verify group configuration
gman repo group list
gman repo group list my-group

# Test without group filter
gman tools find file "*.js"
gman tools find file "*.js" --group my-group
```

### Performance Troubleshooting

#### Memory Usage
```bash
# Monitor memory during search
top -p $(pgrep gman)

# Check available memory
free -h
```

#### Disk I/O
```bash
# Monitor disk usage during search
iotop

# Check disk space
df -h
```

#### Network Issues (for remote repositories)
```bash
# Test network connectivity
ping github.com

# Check DNS resolution
nslookup github.com
```

### Getting Help

If search issues persist:

1. **Check Documentation**: Review [DEPLOYMENT.md](../DEPLOYMENT.md) for installation issues
2. **Verify Dependencies**: Run `./scripts/setup-dependencies.sh --verify-only`
3. **Test Isolation**: Test external tools (fd, rg, fzf) independently
4. **Check Configuration**: Verify repository and group configuration
5. **System Resources**: Ensure adequate memory and disk space
6. **Terminal Environment**: Test in different terminal environments

For additional support, see:
- [USER_GUIDE.md](USER_GUIDE.md) - Complete user guide
- [COMMAND_REFERENCE.md](COMMAND_REFERENCE.md) - Detailed command documentation
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - General troubleshooting guide