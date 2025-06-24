#!/bin/bash

# gman Shell Integration Script
# Add this to your ~/.bashrc or ~/.zshrc

# Main gman wrapper function
gman() {
    local output
    local exit_code

    # Call the actual gman binary and capture both output and exit code
    output=$(command gman "$@" 2>&1)
    exit_code=$?

    # Check if this is a directory change request
    if [[ "$output" == GMAN_CD:* ]]; then
        local target_dir="${output#GMAN_CD:}"
        if [ -d "$target_dir" ]; then
            cd "$target_dir"
            echo "Switched to: $target_dir"
        else
            echo "Error: Directory not found: $target_dir" >&2
            return 1
        fi
    else
        # For all other commands, just print the output
        echo "$output"
    fi

    return $exit_code
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