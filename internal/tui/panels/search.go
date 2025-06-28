package panels

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"gman/internal/tui/models"
	"gman/internal/tui/styles"
)

// SearchPanel provides search functionality integration
type SearchPanel struct {
	state *models.AppState
}

// NewSearchPanel creates a new search panel
func NewSearchPanel(state *models.AppState) *SearchPanel {
	return &SearchPanel{
		state: state,
	}
}

// Init initializes the search panel
func (s *SearchPanel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the search panel
func (s *SearchPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if s.state.FocusedPanel != models.SearchPanel {
			return s, nil
		}

		return s, s.handleKeyMsg(msg)

	case models.SearchModeMsg:
		s.state.SearchState.Mode = msg.Mode
		return s, nil

	case models.SearchResultsMsg:
		s.state.SearchState.Results = msg.Results
		s.state.SearchState.Query = msg.Query
		return s, nil
	}

	return s, nil
}

// View renders the search panel
func (s *SearchPanel) View() string {
	var content strings.Builder

	// Search mode indicator with visual enhancement
	modeIcon := "📁"
	if s.state.SearchState.Mode == models.SearchCommits {
		modeIcon = "📝"
	}

	modeStyle := styles.SubHeaderStyle
	mode := modeStyle.Render(fmt.Sprintf("%s %s search", modeIcon, s.state.SearchState.Mode.String()))
	content.WriteString(mode)
	content.WriteString("\n\n")

	// Search status indicator
	if s.state.SearchState.IsActive {
		activeStyle := styles.ListItemSelectedStyle
		status := activeStyle.Render("🔍 Search active - select result to preview")
		content.WriteString(status)
		content.WriteString("\n\n")
	}

	// Search input/query
	if s.state.SearchState.Query != "" {
		queryStyle := styles.InputStyle
		query := queryStyle.Render("Query: " + s.state.SearchState.Query)
		content.WriteString(query)
		content.WriteString("\n\n")
	}

	// Results or instructions
	if len(s.state.SearchState.Results) > 0 {
		content.WriteString(s.renderResults())
	} else if s.state.SearchState.IsActive {
		content.WriteString(s.renderSearching())
	} else {
		content.WriteString(s.renderInstructions())
	}

	return content.String()
}

// handleKeyMsg handles keyboard input for the search panel
func (s *SearchPanel) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "tab":
		// Toggle search mode
		if s.state.SearchState.Mode == models.SearchFiles {
			return models.SearchModeCmd(models.SearchCommits)
		} else {
			return models.SearchModeCmd(models.SearchFiles)
		}

	case "/":
		// Start file search
		s.state.SearchState.Mode = models.SearchFiles
		return s.launchFzf()

	case "c":
		// Start commit search
		s.state.SearchState.Mode = models.SearchCommits
		return s.launchFzf()

	case "enter":
		// If we have search results, update preview; otherwise launch search
		if len(s.state.SearchState.Results) > 0 && s.state.SearchState.SelectedItem < len(s.state.SearchState.Results) {
			return s.updatePreview()
		} else {
			// Launch search with current mode
			return s.launchFzf()
		}

	case "up", "k":
		if len(s.state.SearchState.Results) > 0 && s.state.SearchState.SelectedItem > 0 {
			s.state.SearchState.SelectedItem--
			// Update preview with newly selected item
			return s.updatePreview()
		}

	case "down", "j":
		if len(s.state.SearchState.Results) > 0 && s.state.SearchState.SelectedItem < len(s.state.SearchState.Results)-1 {
			s.state.SearchState.SelectedItem++
			// Update preview with newly selected item
			return s.updatePreview()
		}
	}

	return nil
}

// launchFzf launches fzf for searching
func (s *SearchPanel) launchFzf() tea.Cmd {
	// Set search as active and launch fzf
	s.state.SearchState.IsActive = true
	s.state.SearchState.Query = "" // Will be filled by fzf

	return func() tea.Msg {
		return models.FzfLaunchMsg{
			Mode:  s.state.SearchState.Mode,
			Query: s.state.SearchState.Query,
		}
	}
}

// renderResults renders search results
func (s *SearchPanel) renderResults() string {
	var content strings.Builder

	content.WriteString(styles.BodyStyle.Render("Results:"))
	content.WriteString("\n")

	for i, result := range s.state.SearchState.Results {
		style := styles.ListItemStyle
		if i == s.state.SearchState.SelectedItem {
			style = styles.ListItemSelectedStyle
		}

		line := style.Render(result.DisplayText)
		content.WriteString(line)
		content.WriteString("\n")
	}

	return content.String()
}

// renderSearching renders the searching state
func (s *SearchPanel) renderSearching() string {
	content := "🔍 Launching fzf search...\n\n"
	content += fmt.Sprintf("Mode: %s\n", s.state.SearchState.Mode.String())
	content += "Searching across all repositories\n\n"
	content += "Note: fzf will open in a separate window.\n"
	content += "Select your item and results will appear here."

	return styles.MutedStyle.Render(content)
}

// renderInstructions renders search instructions
func (s *SearchPanel) renderInstructions() string {
	instructions := []string{
		"Search Instructions:",
		"",
		"/ or Enter - Search files",
		"c - Search commits",
		"Tab - Toggle search mode",
		"",
		"📁 File Search:",
		"  • Find files across all repositories",
		"  • Fuzzy matching on file names and paths",
		"  • Real-time preview in panel 4",
		"",
		"📝 Commit Search:",
		"  • Search commit messages and authors",
		"  • Browse commit history across repos",
		"  • View commit diffs in preview",
		"",
		"Integration with fzf provides powerful",
		"fuzzy search across all repositories.",
	}

	var content strings.Builder
	for _, line := range instructions {
		if line == "" {
			content.WriteString("\n")
		} else if strings.HasSuffix(line, ":") {
			content.WriteString(styles.SubHeaderStyle.Render(line))
			content.WriteString("\n")
		} else if strings.HasPrefix(line, "  ") {
			content.WriteString(styles.MutedStyle.Render(line))
			content.WriteString("\n")
		} else {
			content.WriteString(styles.BodyStyle.Render(line))
			content.WriteString("\n")
		}
	}

	return content.String()
}

// updatePreview updates the preview panel with the currently selected search result
func (s *SearchPanel) updatePreview() tea.Cmd {
	if len(s.state.SearchState.Results) == 0 || s.state.SearchState.SelectedItem >= len(s.state.SearchState.Results) {
		return nil
	}

	result := s.state.SearchState.Results[s.state.SearchState.SelectedItem]

	return func() tea.Msg {
		var content string
		var contentType models.PreviewType

		switch result.Type {
		case "file":
			// Try to read file content for preview
			if result.Path != "" {
				if data, err := os.ReadFile(result.Path); err == nil {
					content = string(data)
				} else {
					content = fmt.Sprintf("Error reading file: %v\n\nFile: %s\nRepository: %s",
						err, result.Path, result.Repository)
				}
			} else {
				content = fmt.Sprintf("File: %s\nRepository: %s\n\nNo content available for preview.",
					result.DisplayText, result.Repository)
			}
			contentType = models.PreviewFile

		case "commit":
			// Show detailed commit information using git show
			if result.Hash != "" && result.Path != "" {
				// Execute git show to get commit details
				gitCmd := exec.Command("git", "show", "--color=always", result.Hash)
				gitCmd.Dir = result.Path
				output, err := gitCmd.Output()
				if err == nil {
					content = string(output)
				} else {
					content = fmt.Sprintf("Commit: %s\nRepository: %s\nError: %v\n\n%s",
						result.Hash, result.Repository, err, result.DisplayText)
				}
			} else {
				content = fmt.Sprintf("Commit: %s\nRepository: %s\n\n%s",
					result.Hash, result.Repository, result.DisplayText)
			}
			contentType = models.PreviewCommit

		default:
			content = fmt.Sprintf("Type: %s\nRepository: %s\nPath: %s\n\n%s",
				result.Type, result.Repository, result.Path, result.DisplayText)
			contentType = models.PreviewStatus
		}

		return models.PreviewContentMsg{
			Content:     content,
			ContentType: contentType,
			FilePath:    result.Path,
			CommitHash:  result.Hash,
		}
	}
}
