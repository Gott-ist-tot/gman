package interactive

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

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