package interactive

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gman/pkg/types"

	"github.com/fatih/color"
)

// RepositorySelector provides interactive repository selection
type RepositorySelector struct {
	repos map[string]string
}

// NewRepositorySelector creates a new repository selector
func NewRepositorySelector(repos map[string]string) *RepositorySelector {
	return &RepositorySelector{repos: repos}
}

// SelectRepository displays an interactive menu and returns the selected repository alias
func (rs *RepositorySelector) SelectRepository() (string, error) {
	if len(rs.repos) == 0 {
		return "", fmt.Errorf("no repositories configured")
	}

	// Convert map to sorted slice for consistent ordering
	var aliases []string
	for alias := range rs.repos {
		aliases = append(aliases, alias)
	}

	// Simple alphabetical sort
	for i := 0; i < len(aliases)-1; i++ {
		for j := i + 1; j < len(aliases); j++ {
			if aliases[i] > aliases[j] {
				aliases[i], aliases[j] = aliases[j], aliases[i]
			}
		}
	}

	// Display the menu
	fmt.Printf("\n%s\n", color.CyanString("Select a repository:"))
	fmt.Println(strings.Repeat("─", 40))

	for i, alias := range aliases {
		path := rs.repos[alias]
		// Truncate path if too long
		displayPath := path
		if len(displayPath) > 50 {
			displayPath = "..." + displayPath[len(displayPath)-47:]
		}
		fmt.Printf("%s %s %s\n",
			color.YellowString("[%d]", i+1),
			color.GreenString("%-15s", alias),
			color.WhiteString("→ %s", displayPath))
	}

	fmt.Println(strings.Repeat("─", 40))
	fmt.Print("Enter number or alias (Ctrl+C to cancel): ")

	// Read user input
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	selection := strings.TrimSpace(input)
	if selection == "" {
		return "", fmt.Errorf("no selection made")
	}

	// Try to parse as number first
	if num, err := strconv.Atoi(selection); err == nil {
		if num >= 1 && num <= len(aliases) {
			return aliases[num-1], nil
		}
		return "", fmt.Errorf("invalid selection number: %d", num)
	}

	// Try as alias
	if _, exists := rs.repos[selection]; exists {
		return selection, nil
	}

	// Try fuzzy matching
	matches := rs.fuzzyMatch(selection, aliases)
	if len(matches) == 1 {
		fmt.Printf("Matched: %s\n", color.GreenString(matches[0]))
		return matches[0], nil
	} else if len(matches) > 1 {
		fmt.Printf("Multiple matches found: %s\n", strings.Join(matches, ", "))
		return "", fmt.Errorf("ambiguous selection, please be more specific")
	}

	return "", fmt.Errorf("repository '%s' not found", selection)
}

// fuzzyMatch performs simple fuzzy matching
func (rs *RepositorySelector) fuzzyMatch(input string, aliases []string) []string {
	input = strings.ToLower(input)
	var matches []string

	for _, alias := range aliases {
		if strings.Contains(strings.ToLower(alias), input) {
			matches = append(matches, alias)
		}
	}

	return matches
}

// SwitchTargetSelector provides interactive selection for repositories and worktrees
type SwitchTargetSelector struct {
	targets []types.SwitchTarget
}

// NewSwitchTargetSelector creates a new switch target selector
func NewSwitchTargetSelector(targets []types.SwitchTarget) *SwitchTargetSelector {
	return &SwitchTargetSelector{targets: targets}
}

// SelectTarget displays an interactive menu and returns the selected target
func (sts *SwitchTargetSelector) SelectTarget() (*types.SwitchTarget, error) {
	if len(sts.targets) == 0 {
		return nil, fmt.Errorf("no repositories or worktrees available")
	}

	// Sort targets: repositories first, then worktrees, alphabetically within each group
	sortedTargets := make([]types.SwitchTarget, len(sts.targets))
	copy(sortedTargets, sts.targets)

	// Simple sort: repositories first, then by alias
	for i := 0; i < len(sortedTargets)-1; i++ {
		for j := i + 1; j < len(sortedTargets); j++ {
			shouldSwap := false

			// Repositories before worktrees
			if sortedTargets[i].Type == "worktree" && sortedTargets[j].Type == "repository" {
				shouldSwap = true
			} else if sortedTargets[i].Type == sortedTargets[j].Type {
				// Same type, sort alphabetically by alias
				if sortedTargets[i].Alias > sortedTargets[j].Alias {
					shouldSwap = true
				}
			}

			if shouldSwap {
				sortedTargets[i], sortedTargets[j] = sortedTargets[j], sortedTargets[i]
			}
		}
	}

	// Display the menu
	fmt.Printf("\n%s\n", color.CyanString("Select a target:"))
	fmt.Println(strings.Repeat("─", 60))

	for i, target := range sortedTargets {
		var icon, typeLabel string
		displayPath := target.Path

		// Customize display based on type
		if target.Type == "repository" {
			icon = "📁"
			typeLabel = color.BlueString("repo")
		} else {
			icon = "🌿"
			typeLabel = color.MagentaString("worktree")
			// For worktrees, show branch info
			if target.Branch != "" {
				typeLabel += color.WhiteString(" (%s)", target.Branch)
			}
		}

		// Truncate path if too long
		if len(displayPath) > 40 {
			displayPath = "..." + displayPath[len(displayPath)-37:]
		}

		fmt.Printf("%s %s %s %-20s %s %s\n",
			color.YellowString("[%d]", i+1),
			icon,
			typeLabel,
			color.GreenString(target.Alias),
			color.WhiteString("→"),
			color.WhiteString(displayPath))
	}

	fmt.Println(strings.Repeat("─", 60))
	fmt.Print("Enter number or alias (Ctrl+C to cancel): ")

	// Read user input
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	selection := strings.TrimSpace(input)
	if selection == "" {
		return nil, fmt.Errorf("no selection made")
	}

	// Try to parse as number first
	if num, err := strconv.Atoi(selection); err == nil {
		if num >= 1 && num <= len(sortedTargets) {
			return &sortedTargets[num-1], nil
		}
		return nil, fmt.Errorf("invalid selection number: %d", num)
	}

	// Try as exact alias match
	for _, target := range sortedTargets {
		if target.Alias == selection {
			return &target, nil
		}
	}

	// Try fuzzy matching
	matches := sts.fuzzyMatch(selection, sortedTargets)
	if len(matches) == 1 {
		fmt.Printf("Matched: %s\n", color.GreenString(matches[0].Alias))
		return &matches[0], nil
	} else if len(matches) > 1 {
		var aliases []string
		for _, match := range matches {
			aliases = append(aliases, match.Alias)
		}
		fmt.Printf("Multiple matches found: %s\n", strings.Join(aliases, ", "))
		return nil, fmt.Errorf("ambiguous selection, please be more specific")
	}

	return nil, fmt.Errorf("target '%s' not found", selection)
}

// fuzzyMatch performs simple fuzzy matching on switch targets
func (sts *SwitchTargetSelector) fuzzyMatch(input string, targets []types.SwitchTarget) []types.SwitchTarget {
	input = strings.ToLower(input)
	var matches []types.SwitchTarget

	for _, target := range targets {
		if strings.Contains(strings.ToLower(target.Alias), input) {
			matches = append(matches, target)
		}
	}

	return matches
}
