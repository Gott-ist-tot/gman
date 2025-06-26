#!/bin/bash

# gman External Dependencies Setup Script
# Installs fd, ripgrep, and fzf for enhanced search functionality

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running on macOS or Linux
detect_os() {
    case "$(uname -s)" in
        Darwin*)    echo "darwin" ;;
        Linux*)     echo "linux" ;;
        MINGW*|MSYS*) echo "windows" ;;
        *)          echo "unknown" ;;
    esac
}

# Detect Linux distribution
detect_linux_distro() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        echo "$ID"
    elif [ -f /etc/redhat-release ]; then
        echo "rhel"
    elif [ -f /etc/debian_version ]; then
        echo "debian"
    else
        echo "unknown"
    fi
}

# Check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Install dependencies on macOS
install_dependencies_macos() {
    if command_exists brew; then
        log_info "Using Homebrew to install dependencies..."
        
        # Check what's already installed
        local to_install=()
        
        if ! command_exists fd; then
            to_install+=("fd")
        else
            log_info "fd is already installed"
        fi
        
        if ! command_exists rg; then
            to_install+=("ripgrep")
        else
            log_info "ripgrep is already installed"
        fi
        
        if ! command_exists fzf; then
            to_install+=("fzf")
        else
            log_info "fzf is already installed"
        fi
        
        if [ ${#to_install[@]} -gt 0 ]; then
            log_info "Installing: ${to_install[*]}"
            brew install "${to_install[@]}" || log_warning "Some dependencies failed to install via Homebrew"
        else
            log_success "All dependencies are already installed"
        fi
        
    elif command_exists port; then
        log_info "Using MacPorts to install dependencies..."
        sudo port install fd ripgrep fzf || log_warning "Some dependencies failed to install via MacPorts"
    else
        log_error "Neither Homebrew nor MacPorts found"
        echo
        echo "Please install a package manager first:"
        echo "  • Homebrew: https://brew.sh"
        echo "  • MacPorts: https://www.macports.org"
        echo
        echo "Then run one of:"
        echo "  brew install fd ripgrep fzf"
        echo "  sudo port install fd ripgrep fzf"
        exit 1
    fi
}

# Install dependencies on Linux
install_dependencies_linux() {
    local distro=$(detect_linux_distro)
    
    case "$distro" in
        "ubuntu"|"debian")
            log_info "Installing dependencies via apt..."
            sudo apt update
            
            # Install packages individually to handle partial failures
            local packages=("fd-find" "ripgrep" "fzf")
            for pkg in "${packages[@]}"; do
                if sudo apt install -y "$pkg"; then
                    log_success "$pkg installed successfully"
                else
                    log_warning "$pkg failed to install"
                fi
            done
            
            # Ubuntu installs fd as fdfind, create symlink
            if command_exists fdfind && ! command_exists fd; then
                if sudo ln -sf "$(which fdfind)" /usr/local/bin/fd; then
                    log_success "Created fd symlink for fdfind"
                else
                    log_warning "Failed to create fd symlink - you may need to use 'fdfind' instead of 'fd'"
                fi
            fi
            ;;
        "fedora")
            log_info "Installing dependencies via dnf..."
            sudo dnf install -y fd-find ripgrep fzf || log_warning "Some dependencies failed to install"
            ;;
        "centos"|"rhel")
            log_info "Installing dependencies via yum (requires EPEL)..."
            sudo yum install -y epel-release
            sudo yum install -y fd-find ripgrep fzf || log_warning "Some dependencies failed to install"
            ;;
        "arch"|"manjaro")
            log_info "Installing dependencies via pacman..."
            sudo pacman -S --noconfirm fd ripgrep fzf || log_warning "Some dependencies failed to install"
            ;;
        "opensuse"*)
            log_info "Installing dependencies via zypper..."
            sudo zypper install -y fd ripgrep fzf || log_warning "Some dependencies failed to install"
            ;;
        "alpine")
            log_info "Installing dependencies via apk..."
            sudo apk add --no-cache fd ripgrep fzf || log_warning "Some dependencies failed to install"
            ;;
        *)
            log_error "Unsupported Linux distribution: $distro"
            echo
            echo "Please install the following packages manually using your distribution's package manager:"
            echo "  • fd (or fd-find)"
            echo "  • ripgrep"
            echo "  • fzf"
            echo
            echo "Package names may vary. Common alternatives:"
            echo "  fd: fd-find, fdfind"
            echo "  ripgrep: rg"
            echo "  fzf: fzf"
            exit 1
            ;;
    esac
}

# Install dependencies on Windows (MSYS2/Git Bash)
install_dependencies_windows() {
    if command_exists pacman; then
        log_info "Installing dependencies via MSYS2 pacman..."
        pacman -S --noconfirm mingw-w64-x86_64-fd mingw-w64-x86_64-ripgrep mingw-w64-x86_64-fzf || log_warning "Some dependencies failed to install"
    elif command_exists winget; then
        log_info "Installing dependencies via winget..."
        
        # Install packages individually
        winget install sharkdp.fd || log_warning "fd installation failed"
        winget install BurntSushi.ripgrep.MSVC || log_warning "ripgrep installation failed"
        winget install junegunn.fzf || log_warning "fzf installation failed"
    else
        log_error "Neither MSYS2 nor winget found"
        echo
        echo "Please install dependencies manually using one of:"
        echo
        echo "Option 1: Using winget (Windows Package Manager)"
        echo "  winget install sharkdp.fd"
        echo "  winget install BurntSushi.ripgrep.MSVC"
        echo "  winget install junegunn.fzf"
        echo
        echo "Option 2: Using MSYS2 (https://www.msys2.org/)"
        echo "  pacman -S mingw-w64-x86_64-fd mingw-w64-x86_64-ripgrep mingw-w64-x86_64-fzf"
        echo
        echo "Option 3: Download binaries manually"
        echo "  • fd: https://github.com/sharkdp/fd/releases"
        echo "  • ripgrep: https://github.com/BurntSushi/ripgrep/releases"
        echo "  • fzf: https://github.com/junegunn/fzf/releases"
        exit 1
    fi
}

# Verify dependencies installation
verify_dependencies() {
    log_info "Verifying dependency installation..."
    echo
    
    local all_good=true
    
    # Check fd
    if command_exists fd; then
        local fd_version=$(fd --version 2>/dev/null || echo "unknown")
        log_success "fd: $fd_version"
    else
        log_error "fd: Not found"
        all_good=false
    fi
    
    # Check ripgrep
    if command_exists rg; then
        local rg_version=$(rg --version | head -1 2>/dev/null || echo "unknown")
        log_success "ripgrep: $rg_version"
    else
        log_error "ripgrep: Not found"
        all_good=false
    fi
    
    # Check fzf
    if command_exists fzf; then
        local fzf_version=$(fzf --version 2>/dev/null || echo "unknown")
        log_success "fzf: $fzf_version"
    else
        log_error "fzf: Not found"
        all_good=false
    fi
    
    echo
    if [ "$all_good" = true ]; then
        log_success "All dependencies are installed and working!"
        echo
        echo "You can now use enhanced gman search commands:"
        echo "  gman tools find file <pattern>     # Fast file search"
        echo "  gman tools find content <pattern>  # Content search"
        echo "  gman tools dashboard               # TUI with search"
    else
        log_warning "Some dependencies are missing"
        echo
        echo "gman will work with reduced functionality."
        echo "Install missing dependencies for full search capabilities."
        return 1
    fi
}

# Show what will be installed
show_installation_plan() {
    local os=$(detect_os)
    
    echo "gman External Dependencies Setup"
    echo "================================="
    echo
    echo "This script will install the following tools for enhanced gman search functionality:"
    echo
    echo "• fd         - Lightning-fast file finder (replaces 'find')"
    echo "• ripgrep    - Ultra-fast text search (replaces 'grep')"
    echo "• fzf        - Interactive fuzzy finder for selection"
    echo
    echo "Detected platform: $os"
    
    case "$os" in
        "darwin")
            if command_exists brew; then
                echo "Package manager: Homebrew"
                echo "Command: brew install fd ripgrep fzf"
            elif command_exists port; then
                echo "Package manager: MacPorts"
                echo "Command: sudo port install fd ripgrep fzf"
            fi
            ;;
        "linux")
            local distro=$(detect_linux_distro)
            echo "Distribution: $distro"
            case "$distro" in
                "ubuntu"|"debian")
                    echo "Package manager: apt"
                    echo "Command: sudo apt install fd-find ripgrep fzf"
                    ;;
                "fedora")
                    echo "Package manager: dnf"
                    echo "Command: sudo dnf install fd-find ripgrep fzf"
                    ;;
                "arch"|"manjaro")
                    echo "Package manager: pacman"
                    echo "Command: sudo pacman -S fd ripgrep fzf"
                    ;;
            esac
            ;;
        "windows")
            if command_exists pacman; then
                echo "Package manager: MSYS2"
            elif command_exists winget; then
                echo "Package manager: winget"
            fi
            ;;
    esac
    echo
}

# Parse command line arguments
parse_args() {
    AUTO_CONFIRM=false
    VERIFY_ONLY=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --auto-confirm|-y)
                AUTO_CONFIRM=true
                shift
                ;;
            --verify-only)
                VERIFY_ONLY=true
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Show help
show_help() {
    echo "gman Dependencies Setup Script"
    echo
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  --auto-confirm, -y    Skip confirmation prompt"
    echo "  --verify-only         Only verify existing installations"
    echo "  --help, -h            Show this help message"
    echo
    echo "This script installs external dependencies for enhanced gman functionality:"
    echo "  • fd: Fast file search"
    echo "  • ripgrep: Fast content search"
    echo "  • fzf: Interactive selection"
}

# Main function
main() {
    # Parse arguments
    parse_args "$@"
    
    # If verify only, just check and exit
    if [ "$VERIFY_ONLY" = true ]; then
        verify_dependencies
        exit $?
    fi
    
    # Show installation plan
    show_installation_plan
    
    # Prompt for confirmation unless auto-confirm
    if [ "$AUTO_CONFIRM" = false ]; then
        echo
        while true; do
            read -p "Proceed with installation? [Y/n]: " yn
            case $yn in
                [Yy]* | "" )
                    break
                    ;;
                [Nn]* )
                    log_info "Installation cancelled"
                    exit 0
                    ;;
                * )
                    echo "Please answer yes (y) or no (n)."
                    ;;
            esac
        done
    fi
    
    echo
    
    # Detect OS and install
    local os=$(detect_os)
    
    case "$os" in
        "darwin")
            install_dependencies_macos
            ;;
        "linux")
            install_dependencies_linux
            ;;
        "windows")
            install_dependencies_windows
            ;;
        *)
            log_error "Unsupported operating system: $os"
            exit 1
            ;;
    esac
    
    echo
    verify_dependencies
}

# Run main function
main "$@"