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
        *)          echo "unknown" ;;
    esac
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

# Show post-install instructions
show_instructions() {
    echo
    log_success "Installation completed successfully!"
    echo
    echo "Next steps:"
    echo "1. Restart your shell or run: source ~/.bashrc (or ~/.zshrc)"
    echo "2. Add your first repository: gman add /path/to/repo alias"
    echo "3. Check status: gman status"
    echo "4. Switch to repository: gman switch alias"
    echo
    echo "For help: gman --help"
    echo "For more information, visit: https://github.com/yourusername/gman"
}

# Main installation function
main() {
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
    
    # Show instructions
    show_instructions
}

# Run installation
main "$@"