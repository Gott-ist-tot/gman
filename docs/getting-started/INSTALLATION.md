# gman Installation Guide

Complete installation guide for gman (Git Repository Manager) with all modern search capabilities across different platforms.

## Quick Installation (Recommended)

For the fastest setup experience:

```bash
# 1. Clone and build gman
git clone <repository-url>
cd gman
go build -o gman .

# 2. Run automated setup (installs dependencies + shell integration)
./scripts/quick-setup.sh

# 3. Restart your shell or source your config
source ~/.bashrc  # or ~/.zshrc

# 4. Start using gman
gman tools setup  # Interactive setup wizard
```

## System Requirements

### Minimum Requirements
- **Go**: 1.19+ (for building from source)
- **Git**: Any recent version (required for core functionality)
- **Shell**: bash, zsh, fish, or compatible POSIX shell

### Recommended External Tools
These tools enable advanced search capabilities:
- **fd**: Lightning-fast file search (replaces traditional `find`)
- **rg (ripgrep)**: Powerful regex content search
- **fzf**: Interactive fuzzy finder for selections

**Note**: gman works without these tools but with reduced search functionality.

## Installation Methods

### Method 1: Automated Installation (Recommended)

```bash
# Download and run the complete setup
git clone <repository-url>
cd gman
./scripts/quick-setup.sh
```

This script:
- Builds the gman binary
- Installs external dependencies (fd, rg, fzf)
- Sets up shell integration
- Configures completions
- Verifies the installation

### Method 2: Manual Installation

#### Step 1: Build gman
```bash
git clone <repository-url>
cd gman
go build -o gman .
sudo mv gman /usr/local/bin/
```

#### Step 2: Install Dependencies
```bash
# Run dependency installer
./scripts/setup-dependencies.sh

# Or install manually (see Platform-Specific Instructions below)
```

#### Step 3: Shell Integration
```bash
# Copy integration script
mkdir -p ~/.config/gman
cp scripts/shell-integration.sh ~/.config/gman/

# Add to your shell config
echo 'source ~/.config/gman/shell-integration.sh' >> ~/.bashrc  # or ~/.zshrc
```

### Method 3: Development Installation

For development or testing:

```bash
git clone <repository-url>
cd gman
go build -o gman .
export PATH="$PWD:$PATH"  # Add to current session
source scripts/shell-integration.sh  # Enable shell integration
```

## External Tool Dependencies

gman's modern search system relies on external tools for optimal performance:

### fd (File Discovery)
**Purpose**: Lightning-fast file search across repositories
**Installation**: See platform-specific instructions below
**Usage**: `gman find file <pattern>` - instant file search

### rg (ripgrep)
**Purpose**: Regex-powered content search within files
**Installation**: See platform-specific instructions below  
**Usage**: `gman find content <pattern>` - search file contents

### fzf (Fuzzy Finder)
**Purpose**: Interactive selection interface
**Installation**: See platform-specific instructions below
**Usage**: Enhanced interactive experience for all find commands

### Dependency Detection

gman automatically detects missing tools and provides installation guidance:

```bash
# Test dependency detection
gman tools find file config  # Will show missing tool instructions if needed
```

## Shell Integration Setup

**Critical**: Shell integration is required for `gman switch` to change directories.

### Automatic Setup
```bash
./scripts/install.sh  # Includes shell integration
```

### Manual Setup

#### For Bash
Add to `~/.bashrc`:
```bash
# gman Shell Integration
if [ -f "$HOME/.config/gman/shell-integration.sh" ]; then
    source "$HOME/.config/gman/shell-integration.sh"
fi
```

#### For Zsh  
Add to `~/.zshrc`:
```bash
# gman Shell Integration
if [ -f "$HOME/.config/gman/shell-integration.sh" ]; then
    source "$HOME/.config/gman/shell-integration.sh"
fi
```

#### For Fish
Add to `~/.config/fish/config.fish`:
```fish
# gman Shell Integration (requires adaptation for fish syntax)
if test -f "$HOME/.config/gman/shell-integration.sh"
    source "$HOME/.config/gman/shell-integration.sh"
end
```

### Shell Integration Features
- **Directory Switching**: `gman switch <repo>` actually changes directories
- **Warning Preservation**: Shows warnings while still switching directories
- **Interactive Commands**: Dashboard and other TUI commands work properly
- **Completion**: Tab completion for commands and repository names

## Platform-Specific Instructions

### macOS

#### Using Homebrew (Recommended)
```bash
# Install dependencies
brew install fd ripgrep fzf

# Install gman
git clone <repository-url>
cd gman
go build -o gman .
brew install --HEAD .  # If Homebrew formula available
# OR
sudo mv gman /usr/local/bin/
```

#### Using MacPorts
```bash
# Install dependencies
sudo port install fd ripgrep fzf

# Build and install gman
git clone <repository-url>
cd gman  
go build -o gman .
sudo mv gman /opt/local/bin/
```

### Linux

#### Ubuntu/Debian
```bash
# Install dependencies
sudo apt update
sudo apt install fd-find ripgrep fzf

# Note: On Ubuntu, fd is installed as 'fdfind'
sudo ln -sf $(which fdfind) /usr/local/bin/fd

# Build and install gman
git clone <repository-url>
cd gman
go build -o gman .
sudo mv gman /usr/local/bin/
```

#### CentOS/RHEL/Fedora
```bash
# Install dependencies (Fedora)
sudo dnf install fd-find ripgrep fzf

# Install dependencies (CentOS/RHEL with EPEL)
sudo yum install epel-release
sudo yum install fd-find ripgrep fzf

# Build and install gman
git clone <repository-url>
cd gman
go build -o gman .
sudo mv gman /usr/local/bin/
```

#### Arch Linux
```bash
# Install dependencies
sudo pacman -S fd ripgrep fzf

# Build and install gman
git clone <repository-url>
cd gman
go build -o gman .
sudo mv gman /usr/local/bin/
```

### Windows (WSL/MSYS2)

#### Windows Subsystem for Linux (WSL)
Follow the Linux instructions for your WSL distribution.

#### MSYS2
```bash
# Install dependencies
pacman -S mingw-w64-x86_64-fd mingw-w64-x86_64-ripgrep mingw-w64-x86_64-fzf

# Build and install gman
git clone <repository-url>
cd gman
go build -o gman.exe .
cp gman.exe /mingw64/bin/
```

#### Using Winget
```bash
# Install dependencies
winget install sharkdp.fd
winget install BurntSushi.ripgrep.MSVC
winget install junegunn.fzf

# Build gman in WSL or Git Bash
```

## Verification & Testing

### Quick Verification
```bash
# Check gman installation
gman --version
gman --help

# Test shell integration
gman switch --help  # Should show switch command help

# Test dependencies
gman tools find file --help  # Should work without errors
gman tools find content --help  # Should work without errors
```

### Comprehensive Testing
```bash
# Run verification script (if available)
./scripts/verify-setup.sh

# OR manual testing:

# 1. Add a test repository
gman repo add test-repo /path/to/any/git/repo

# 2. Test basic functionality
gman repo list
gman work status

# 3. Test search functionality
gman tools find file config    # Should launch fzf (Ctrl-C to cancel)
gman tools find content "import"  # Should search content (Ctrl-C to cancel)

# 4. Test directory switching
gman switch test-repo  # Should change to repository directory
pwd  # Should show repository path
```

### Dependency Verification
```bash
# Check external tool versions
fd --version
rg --version  
fzf --version

# Test gman's dependency detection
gman tools find file test 2>&1 | head -5  # Check for missing tool warnings
```

## Post-Installation

After successful installation:

1. **Setup Repositories**: `gman tools setup` for guided setup
2. **Explore Features**: `gman --help` for command overview  
3. **Try Search**: `gman tools find file config` to test search
4. **Configure Groups**: `gman repo group create` for organization
5. **Launch Dashboard**: `gman tools dashboard` for TUI interface

## Next Steps

- **[Quick Start Tutorial](QUICK_START.md)** - Get started with gman basics
- **[User Guide](../user-guide/USER_GUIDE.md)** - Comprehensive usage guide
- **[Troubleshooting](../troubleshooting/TROUBLESHOOTING.md)** - Common issues and solutions

Enjoy the enhanced multi-repository management experience with gman!