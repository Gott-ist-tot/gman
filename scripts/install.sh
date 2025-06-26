#!/bin/bash

# gman Installation Script

set -e

# Configuration
BINARY_NAME="gman"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/gman"

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

# Install external dependencies
install_dependencies() {
    local os=$(detect_os)
    log_info "Installing external dependencies (fd, rg, fzf)..."
    
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
            log_warning "Unsupported OS for automatic dependency installation"
            log_info "Please install fd, ripgrep, and fzf manually for enhanced search functionality"
            return 0
            ;;
    esac
}

# Install dependencies on macOS
install_dependencies_macos() {
    if command_exists brew; then
        log_info "Using Homebrew to install dependencies..."
        brew install fd ripgrep fzf || log_warning "Some dependencies failed to install via Homebrew"
    elif command_exists port; then
        log_info "Using MacPorts to install dependencies..."
        sudo port install fd ripgrep fzf || log_warning "Some dependencies failed to install via MacPorts"
    else
        log_warning "Neither Homebrew nor MacPorts found"
        log_info "Please install Homebrew (https://brew.sh) or MacPorts and run:"
        log_info "  brew install fd ripgrep fzf"
        log_info "  # OR"
        log_info "  sudo port install fd ripgrep fzf"
    fi
}

# Install dependencies on Linux
install_dependencies_linux() {
    local distro=$(detect_linux_distro)
    
    case "$distro" in
        "ubuntu"|"debian")
            log_info "Installing dependencies via apt..."
            sudo apt update
            sudo apt install -y fd-find ripgrep fzf || log_warning "Some dependencies failed to install"
            # Ubuntu installs fd as fdfind, create symlink
            if command_exists fdfind && ! command_exists fd; then
                sudo ln -sf "$(which fdfind)" /usr/local/bin/fd
                log_info "Created fd symlink for fdfind"
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
        *)
            log_warning "Unsupported Linux distribution: $distro"
            log_info "Please install fd, ripgrep, and fzf manually using your package manager"
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
        winget install sharkdp.fd || log_warning "fd installation failed"
        winget install BurntSushi.ripgrep.MSVC || log_warning "ripgrep installation failed"
        winget install junegunn.fzf || log_warning "fzf installation failed"
    else
        log_warning "Neither MSYS2 nor winget found"
        log_info "Please install dependencies manually:"
        log_info "  winget install sharkdp.fd BurntSushi.ripgrep.MSVC junegunn.fzf"
    fi
}

# Verify dependencies installation
verify_dependencies() {
    log_info "Verifying dependency installation..."
    local missing_deps=()
    
    if ! command_exists fd; then
        missing_deps+=("fd")
    fi
    
    if ! command_exists rg; then
        missing_deps+=("rg (ripgrep)")
    fi
    
    if ! command_exists fzf; then
        missing_deps+=("fzf")
    fi
    
    if [ ${#missing_deps[@]} -eq 0 ]; then
        log_success "All dependencies are installed and available"
        return 0
    else
        log_warning "Missing dependencies: ${missing_deps[*]}"
        log_info "gman will work with reduced search functionality"
        log_info "For full functionality, please install missing dependencies manually"
        return 1
    fi
}

# Check if binary exists
check_binary() {
    if [ ! -f "$BINARY_NAME" ]; then
        log_error "Binary '$BINARY_NAME' not found in current directory"
        log_info "Please run 'go build' first to create the binary"
        exit 1
    fi
}

# Install binary
install_binary() {
    log_info "Installing $BINARY_NAME to $INSTALL_DIR..."
    
    if [ ! -w "$INSTALL_DIR" ]; then
        log_info "Need sudo permissions to install to $INSTALL_DIR"
        sudo cp "$BINARY_NAME" "$INSTALL_DIR/"
        sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
    else
        cp "$BINARY_NAME" "$INSTALL_DIR/"
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
    fi
    
    log_success "Binary installed successfully"
}

# Create config directory
create_config_dir() {
    if [ ! -d "$CONFIG_DIR" ]; then
        log_info "Creating config directory: $CONFIG_DIR"
        mkdir -p "$CONFIG_DIR"
        log_success "Config directory created"
    else
        log_info "Config directory already exists: $CONFIG_DIR"
    fi
}

# Setup shell integration
setup_shell_integration() {
    local shell_rc=""
    local shell_name=""
    
    # Detect shell
    if [ -n "$BASH_VERSION" ]; then
        shell_rc="$HOME/.bashrc"
        shell_name="bash"
    elif [ -n "$ZSH_VERSION" ]; then
        shell_rc="$HOME/.zshrc"
        shell_name="zsh"
    else
        # Try to detect from $SHELL
        case "$SHELL" in
            */bash)
                shell_rc="$HOME/.bashrc"
                shell_name="bash"
                ;;
            */zsh)
                shell_rc="$HOME/.zshrc"
                shell_name="zsh"
                ;;
            *)
                log_warning "Could not detect shell type. Manual setup required."
                return
                ;;
        esac
    fi
    
    log_info "Setting up shell integration for $shell_name..."
    
    # Check if integration is already added
    if grep -q "gman Shell Integration" "$shell_rc" 2>/dev/null; then
        log_info "Shell integration already configured in $shell_rc"
        return
    fi
    
    # Add shell integration
    cat >> "$shell_rc" << 'EOF'

# gman Shell Integration
# Source the gman shell integration if available
if [ -f "$HOME/.config/gman/shell-integration.sh" ]; then
    source "$HOME/.config/gman/shell-integration.sh"
fi
EOF
    
    # Copy shell integration script
    cp "scripts/shell-integration.sh" "$CONFIG_DIR/"
    
    log_success "Shell integration added to $shell_rc"
    log_info "Please restart your shell or run: source $shell_rc"
}

# Generate completion scripts
generate_completions() {
    log_info "Generating shell completion scripts..."
    
    # Create completions directory
    local completions_dir="$CONFIG_DIR/completions"
    mkdir -p "$completions_dir"
    
    # Generate completions for both bash and zsh
    if command -v "$INSTALL_DIR/$BINARY_NAME" &> /dev/null; then
        "$INSTALL_DIR/$BINARY_NAME" completion bash > "$completions_dir/gman.bash"
        "$INSTALL_DIR/$BINARY_NAME" completion zsh > "$completions_dir/gman.zsh"
        log_success "Completion scripts generated"
    else
        log_warning "Could not generate completion scripts - binary not in PATH"
    fi
}

# Prompt for dependency installation
prompt_dependency_installation() {
    echo
    log_info "External dependencies enhance gman's search capabilities:"
    echo "  • fd: Lightning-fast file search"
    echo "  • rg (ripgrep): Powerful content search"
    echo "  • fzf: Interactive fuzzy finder"
    echo
    
    while true; do
        read -p "Install external dependencies? [Y/n]: " yn
        case $yn in
            [Yy]* | "" )
                install_dependencies
                verify_dependencies
                break
                ;;
            [Nn]* )
                log_info "Skipping dependency installation"
                log_info "You can install them later using: ./scripts/setup-dependencies.sh"
                break
                ;;
            * )
                echo "Please answer yes (y) or no (n)."
                ;;
        esac
    done
}

# Show post-install instructions
show_instructions() {
    echo
    log_success "Installation completed successfully!"
    echo
    echo "Next steps:"
    echo "1. Restart your shell or run: source ~/.bashrc (or ~/.zshrc)"
    echo "2. Run the setup wizard: gman tools setup"
    echo "3. Add your first repository: gman repo add /path/to/repo alias"
    echo "4. Check status: gman work status"
    echo "5. Switch to repository: gman switch alias"
    echo
    echo "Enhanced search commands:"
    echo "  • gman tools find file <pattern>     # Search for files"
    echo "  • gman tools find content <pattern>  # Search file contents"
    echo "  • gman tools dashboard              # Launch TUI dashboard"
    echo
    echo "For help: gman --help"
    echo "For troubleshooting: see DEPLOYMENT.md"
}

# Parse command line arguments
parse_args() {
    SKIP_DEPS=false
    AUTO_DEPS=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-deps)
                SKIP_DEPS=true
                shift
                ;;
            --auto-deps)
                AUTO_DEPS=true
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
    echo "gman Installation Script"
    echo
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  --skip-deps    Skip dependency installation"
    echo "  --auto-deps    Automatically install dependencies without prompting"
    echo "  --help, -h     Show this help message"
    echo
    echo "This script will:"
    echo "1. Install the gman binary to /usr/local/bin"
    echo "2. Set up configuration directory"
    echo "3. Configure shell integration for directory switching"
    echo "4. Generate completion scripts"
    echo "5. Optionally install external dependencies (fd, rg, fzf)"
}

# Main installation function
main() {
    # Parse arguments
    parse_args "$@"
    
    log_info "Starting gman installation..."
    echo
    
    # Detect OS
    local os=$(detect_os)
    log_info "Detected OS: $os"
    
    if [ "$os" = "unknown" ]; then
        log_error "Unsupported operating system"
        exit 1
    fi
    
    # Check if binary exists
    check_binary
    
    # Install components
    install_binary
    create_config_dir
    setup_shell_integration
    generate_completions
    
    # Handle dependencies
    if [ "$SKIP_DEPS" = true ]; then
        log_info "Skipping dependency installation (--skip-deps)"
    elif [ "$AUTO_DEPS" = true ]; then
        log_info "Installing dependencies automatically (--auto-deps)"
        install_dependencies
        verify_dependencies
    else
        prompt_dependency_installation
    fi
    
    # Show instructions
    show_instructions
}

# Run installation
main "$@"