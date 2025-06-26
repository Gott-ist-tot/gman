#!/bin/bash

# gman Shell Integration Script
# Add this to your ~/.bashrc or ~/.zshrc

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