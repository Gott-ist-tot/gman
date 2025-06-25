package panels

import (
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

	// Search mode indicator
	modeStyle := styles.SubHeaderStyle
	mode := modeStyle.Render(s.state.SearchState.Mode.String() + " search")
	content.WriteString(mode)
	content.WriteString("\n\n")

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
		// Launch search with current mode
		return s.launchFzf()

	case "up", "k":
		if s.state.SearchState.SelectedItem > 0 {
			s.state.SearchState.SelectedItem--
		}

	case "down", "j":
		if s.state.SearchState.SelectedItem < len(s.state.SearchState.Results)-1 {
			s.state.SearchState.SelectedItem++
		}
	}

	return nil
}

// launchFzf launches fzf for searching
func (s *SearchPanel) launchFzf() tea.Cmd {
	// This would launch the fzf integration from Phase 5.1
	// For now, return a placeholder command
	return func() tea.Msg {
		return models.FzfLaunchMsg{
			Mode:  s.state.SearchState.Mode,
			Query: "",
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

// renderInstructions renders search instructions
func (s *SearchPanel) renderInstructions() string {
	instructions := []string{
		"Search Instructions:",
		"",
		"/ or Enter - Search files",
		"c - Search commits",
		"Tab - Toggle search mode",
		"",
		"Integration with fzf from Phase 5.1",
		"provides powerful fuzzy search",
		"across all repositories.",
	}

	var content strings.Builder
	for _, line := range instructions {
		if line == "" {
			content.WriteString("\n")
		} else if strings.HasSuffix(line, ":") {
			content.WriteString(styles.SubHeaderStyle.Render(line))
			content.WriteString("\n")
		} else {
			content.WriteString(styles.BodyStyle.Render(line))
			content.WriteString("\n")
		}
	}

	return content.String()
}