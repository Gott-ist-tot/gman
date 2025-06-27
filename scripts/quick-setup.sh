#!/bin/bash

# gman Quick Setup Script
# One-command setup for new users with automatic dependency installation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
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

log_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

# Check if running from gman directory
check_directory() {
    if [ ! -f "main.go" ] || [ ! -f "go.mod" ]; then
        log_error "This script must be run from the gman project directory"
        log_info "Please 'cd' to the gman directory and run: ./scripts/quick-setup.sh"
        exit 1
    fi
}

# Check prerequisites
check_prerequisites() {
    log_step "Checking prerequisites..."
    
    # Check Go installation
    if ! command -v go >/dev/null 2>&1; then
        log_error "Go is not installed or not in PATH"
        log_info "Please install Go from: https://golang.org/doc/install"
        exit 1
    fi
    
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Found Go version: $go_version"
    
    # Check Git installation
    if ! command -v git >/dev/null 2>&1; then
        log_error "Git is not installed or not in PATH"
        log_info "Please install Git from: https://git-scm.com/downloads"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Build gman binary
build_gman() {
    log_step "Building gman binary..."
    
    if go build -o gman .; then
        log_success "gman binary built successfully"
        
        # Make executable
        chmod +x gman
        
        # Show binary info
        local binary_size=$(ls -lh gman | awk '{print $5}')
        log_info "Binary size: $binary_size"
    else
        log_error "Failed to build gman binary"
        exit 1
    fi
}

# Install binary
install_binary() {
    log_step "Installing gman binary..."
    
    local install_dir="/usr/local/bin"
    
    if [ -w "$install_dir" ]; then
        cp gman "$install_dir/"
        log_success "gman installed to $install_dir"
    else
        log_info "Need sudo permissions to install to $install_dir"
        if sudo cp gman "$install_dir/"; then
            sudo chmod +x "$install_dir/gman"
            log_success "gman installed to $install_dir"
        else
            log_error "Failed to install gman to $install_dir"
            exit 1
        fi
    fi
}

# Setup configuration directory
setup_config() {
    log_step "Setting up configuration directory..."
    
    local config_dir="$HOME/.config/gman"
    
    if [ ! -d "$config_dir" ]; then
        mkdir -p "$config_dir"
        log_success "Created configuration directory: $config_dir"
    else
        log_info "Configuration directory already exists: $config_dir"
    fi
    
    # Copy shell integration
    if [ -f "scripts/shell-integration.sh" ]; then
        cp scripts/shell-integration.sh "$config_dir/"
        log_success "Shell integration script copied"
    else
        log_warning "Shell integration script not found"
    fi
}

# Setup shell integration
setup_shell() {
    log_step "Setting up shell integration..."
    
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
                log_info "Please add this to your shell configuration:"
                log_info "source ~/.config/gman/shell-integration.sh"
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

# gman Shell Integration - Added by quick-setup.sh
if [ -f "$HOME/.config/gman/shell-integration.sh" ]; then
    source "$HOME/.config/gman/shell-integration.sh"
fi
EOF
    
    log_success "Shell integration added to $shell_rc"
    log_info "Please restart your shell or run: source $shell_rc"
}

# Install external dependencies
install_dependencies() {
    log_step "Installing external dependencies (fd, ripgrep, fzf)..."
    
    if [ -f "scripts/setup-dependencies.sh" ]; then
        if ./scripts/setup-dependencies.sh --auto-confirm; then
            log_success "External dependencies installed"
        else
            log_warning "Some dependencies failed to install"
            log_info "You can install them later with: ./scripts/setup-dependencies.sh"
        fi
    else
        log_warning "Dependency installer not found"
        log_info "Please install fd, ripgrep, and fzf manually for enhanced search"
    fi
}

# Generate completions
generate_completions() {
    log_step "Generating shell completions..."
    
    local completions_dir="$HOME/.config/gman/completions"
    mkdir -p "$completions_dir"
    
    if command -v gman >/dev/null 2>&1; then
        gman completion bash > "$completions_dir/gman.bash" 2>/dev/null || true
        gman completion zsh > "$completions_dir/gman.zsh" 2>/dev/null || true
        log_success "Shell completions generated"
    else
        log_warning "Could not generate completions - gman not in PATH yet"
    fi
}

# Run interactive setup
run_setup_wizard() {
    log_step "Running gman setup wizard..."
    
    if command -v gman >/dev/null 2>&1; then
        log_info "Starting interactive setup..."
        echo ""
        
        # Run setup in a new shell to ensure PATH is updated
        if bash -c "source ~/.bashrc 2>/dev/null || source ~/.zshrc 2>/dev/null || true; gman tools setup"; then
            log_success "Setup wizard completed"
        else
            log_warning "Setup wizard had issues - you can run it later with: gman tools setup"
        fi
    else
        log_warning "gman not found in PATH - setup wizard skipped"
        log_info "Please restart your shell and run: gman tools setup"
    fi
}

# Show completion message
show_completion() {
    echo ""
    echo "======================================"
    echo -e "${GREEN}🎉 gman Quick Setup Complete! 🎉${NC}"
    echo "======================================"
    echo ""
    echo "What was installed:"
    echo "  ✅ gman binary to /usr/local/bin"
    echo "  ✅ Configuration directory created"
    echo "  ✅ Shell integration configured"
    echo "  ✅ External dependencies (fd, rg, fzf)"
    echo "  ✅ Shell completions generated"
    echo ""
    echo "Next steps:"
    echo "  1. Restart your shell or run: source ~/.bashrc (or ~/.zshrc)"
    echo "  2. Verify installation: gman --version"
    echo "  3. Add repositories: gman repo add /path/to/repo"
    echo "  4. Explore features: gman --help"
    echo ""
    echo "Key commands to get started:"
    echo "  gman tools setup          # Interactive setup wizard"
    echo "  gman repo add . my-repo   # Add current directory"
    echo "  gman work status          # Check repository status"
    echo "  gman tools dashboard      # Launch TUI interface"
    echo "  gman tools find file config.yaml  # Search for files"
    echo ""
    echo "Documentation:"
    echo "  README.md                 # Getting started guide"
    echo "  docs/USER_GUIDE.md        # Comprehensive user manual"
    echo "  docs/SEARCH_GUIDE.md      # Enhanced search features"
    echo "  DEPLOYMENT.md             # Detailed installation guide"
    echo ""
    echo "Need help? Run: gman --help"
    echo ""
}

# Parse command line arguments
parse_args() {
    SKIP_DEPENDENCIES=false
    SKIP_WIZARD=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-deps)
                SKIP_DEPENDENCIES=true
                shift
                ;;
            --skip-wizard)
                SKIP_WIZARD=true
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
    echo "gman Quick Setup Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --skip-deps     Skip external dependency installation"
    echo "  --skip-wizard   Skip interactive setup wizard"
    echo "  --help, -h      Show this help message"
    echo ""
    echo "This script provides one-command setup for gman including:"
    echo "  • Building and installing the gman binary"
    echo "  • Setting up configuration and shell integration"
    echo "  • Installing external dependencies (fd, rg, fzf)"
    echo "  • Generating shell completions"
    echo "  • Running the interactive setup wizard"
    echo ""
    echo "Requirements:"
    echo "  • Go 1.19+ for building"
    echo "  • Git for repository operations"
    echo "  • Internet connection for dependency installation"
    echo ""
}

# Main function
main() {
    # Parse arguments
    parse_args "$@"
    
    echo ""
    echo "======================================"
    echo -e "${BLUE}🚀 gman Quick Setup${NC}"
    echo "======================================"
    echo ""
    echo "This script will:"
    echo "  1. Build and install gman binary"
    echo "  2. Set up configuration and shell integration"
    echo "  3. Install external dependencies (fd, rg, fzf)"
    echo "  4. Generate shell completions"
    echo "  5. Run interactive setup wizard"
    echo ""
    
    # Confirmation
    read -p "Continue with quick setup? [Y/n]: " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        log_info "Setup cancelled by user"
        exit 0
    fi
    
    echo ""
    
    # Execute setup steps
    check_directory
    check_prerequisites
    build_gman
    install_binary
    setup_config
    setup_shell
    
    if [ "$SKIP_DEPENDENCIES" = false ]; then
        install_dependencies
    else
        log_info "Skipping dependency installation (--skip-deps)"
    fi
    
    generate_completions
    
    if [ "$SKIP_WIZARD" = false ]; then
        echo ""
        read -p "Run interactive setup wizard now? [Y/n]: " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Nn]$ ]]; then
            run_setup_wizard
        fi
    else
        log_info "Skipping setup wizard (--skip-wizard)"
    fi
    
    show_completion
}

# Run main function
main "$@"