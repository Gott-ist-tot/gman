#!/bin/bash

# gman Shell Integration Script
# Add this to your ~/.bashrc or ~/.zshrc
# This script provides:
# - Directory switching for 'gman switch'
# - Shell completion for gman commands
# - Optional dependency checking and installation guidance

# Set environment variable to indicate shell integration is active
export GMAN_SHELL_INTEGRATION=1

# Main gman wrapper function with smart command detection
gman() {
    # Check if the first argument is 'switch' or its aliases
    if [[ "$1" == "switch" || "$1" == "sw" || "$1" == "cd" ]]; then
        local output gman_cd_line
        # For switch commands, capture output to handle GMAN_CD
        output=$(command gman "$@" 2>&1)
        local exit_code=$?
        
        # Extract GMAN_CD line while preserving other output (like warnings)
        gman_cd_line=$(echo "$output" | grep "^GMAN_CD:")
        
        if [[ -n "$gman_cd_line" ]]; then
            local target_dir="${gman_cd_line#GMAN_CD:}"
            # Print all non-GMAN_CD output first (warnings, messages, etc.)
            echo "$output" | grep -v "^GMAN_CD:" | grep -v "^$"
            
            if [ -d "$target_dir" ]; then
                cd "$target_dir"
                echo "Switched to: $target_dir"
            else
                echo "Error: Directory not found: $target_dir" >&2
                return 1
            fi
        else
            # If switch failed, print the error output
            echo "$output"
        fi
        return $exit_code
    else
        # For all other commands (including dashboard), execute directly without capturing
        # This allows interactive commands like dashboard to work properly
        command gman "$@"
    fi
}

# Enable bash completion for gman
if [ -n "$BASH_VERSION" ]; then
    # Check if gman completion is available
    if command -v gman &> /dev/null; then
        eval "$(gman completion bash)"
    fi
elif [ -n "$ZSH_VERSION" ]; then
    # Check if gman completion is available
    if command -v gman &> /dev/null; then
        eval "$(gman completion zsh)"
    fi
fi

# Dependency checking functions
gman_check_deps() {
    local missing_deps=()
    
    if ! command -v fd >/dev/null 2>&1; then
        missing_deps+=("fd")
    fi
    
    if ! command -v rg >/dev/null 2>&1; then
        missing_deps+=("rg")
    fi
    
    if ! command -v fzf >/dev/null 2>&1; then
        missing_deps+=("fzf")
    fi
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        return 1
    else
        return 0
    fi
}

gman_show_dep_warning() {
    local missing_deps=()
    
    if ! command -v fd >/dev/null 2>&1; then
        missing_deps+=("fd")
    fi
    
    if ! command -v rg >/dev/null 2>&1; then
        missing_deps+=("rg")
    fi
    
    if ! command -v fzf >/dev/null 2>&1; then
        missing_deps+=("fzf")
    fi
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        echo "⚠️  gman: Missing external dependencies for enhanced search: ${missing_deps[*]}" >&2
        echo "   Install with: ./scripts/setup-dependencies.sh" >&2
        echo "   Or see: DEPLOYMENT.md for manual installation" >&2
        return 1
    fi
    
    return 0
}

# Enhanced gman wrapper with dependency awareness
gman_enhanced() {
    # Check if this is a search command that requires external dependencies
    if [[ "$1" == "tools" && "$2" == "find" ]] || [[ "$1" == "find" ]]; then
        if ! gman_check_deps; then
            gman_show_dep_warning
            echo "" >&2
            echo "Continuing with basic functionality..." >&2
        fi
    fi
    
    # Call the main gman function
    gman "$@"
}

# Optional: Check dependencies on shell startup
# Uncomment the following lines to get a one-time warning about missing dependencies
# when starting a new shell session (only if gman config exists)
#
# if [ -f "$HOME/.config/gman/config.yml" ] && ! gman_check_deps >/dev/null 2>&1; then
#     echo "💡 Tip: Install gman's external dependencies for enhanced search functionality:"
#     echo "   ./scripts/setup-dependencies.sh"
# fi

# Optional: Add gman status to your prompt (uncomment to enable)
# This shows a git-like status indicator for the current repository
# 
# gman_prompt_status() {
#     if command -v gman &> /dev/null; then
#         local current_repo=$(gman status 2>/dev/null | grep "^\*" | awk '{print $2}')
#         if [ -n "$current_repo" ]; then
#             echo " [$current_repo]"
#         fi
#     fi
# }

# Example of how to add gman status to PS1 (bash) or PROMPT (zsh)
# Uncomment and customize as needed:
#
# if [ -n "$BASH_VERSION" ]; then
#     PS1="${PS1}\$(gman_prompt_status)"
# elif [ -n "$ZSH_VERSION" ]; then
#     PROMPT="${PROMPT}\$(gman_prompt_status)"
# fi

# Optional: Use enhanced gman with dependency checking
# Uncomment the following line to enable dependency warnings for search commands:
# alias gman=gman_enhanced