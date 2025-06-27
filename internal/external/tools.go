package external

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Tool represents an external command-line tool
type Tool struct {
	Name            string
	Command         string
	MinVersion      string
	InstallCommands map[string]string // platform -> install command
	CheckCmd        []string          // command to check if tool exists
	Description     string            // tool description
	Website         string            // official website
	Required        bool              // whether tool is required or optional
}

// Common external tools used by gman
var (
	FD = &Tool{
		Name:        "fd",
		Command:     "fd",
		Description: "Lightning-fast file search tool (faster alternative to find)",
		Website:     "https://github.com/sharkdp/fd",
		Required:    false,
		InstallCommands: map[string]string{
			"darwin":  "brew install fd",
			"linux":   "apt install fd-find || yum install fd-find || pacman -S fd",
			"windows": "winget install sharkdp.fd",
		},
		CheckCmd: []string{"fd", "--version"},
	}

	RipGrep = &Tool{
		Name:        "ripgrep",
		Command:     "rg",
		Description: "Extremely fast regex-based content search tool",
		Website:     "https://github.com/BurntSushi/ripgrep",
		Required:    false,
		InstallCommands: map[string]string{
			"darwin":  "brew install ripgrep",
			"linux":   "apt install ripgrep || yum install ripgrep || pacman -S ripgrep",
			"windows": "winget install BurntSushi.ripgrep.MSVC",
		},
		CheckCmd: []string{"rg", "--version"},
	}

	FZF = &Tool{
		Name:        "fzf",
		Command:     "fzf",
		Description: "Interactive fuzzy finder for enhanced user experience",
		Website:     "https://github.com/junegunn/fzf",
		Required:    false,
		InstallCommands: map[string]string{
			"darwin":  "brew install fzf",
			"linux":   "apt install fzf || yum install fzf || pacman -S fzf",
			"windows": "winget install junegunn.fzf",
		},
		CheckCmd: []string{"fzf", "--version"},
	}
)

// IsAvailable checks if the tool is available on the system
func (t *Tool) IsAvailable() bool {
	_, err := exec.LookPath(t.Command)
	return err == nil
}

// GetVersion returns the version string of the tool
func (t *Tool) GetVersion() (string, error) {
	if !t.IsAvailable() {
		return "", fmt.Errorf("%s not found", t.Name)
	}

	cmd := exec.Command(t.CheckCmd[0], t.CheckCmd[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get %s version: %w", t.Name, err)
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}

// GetInstallInstructions returns installation instructions for the current platform
func (t *Tool) GetInstallInstructions() string {
	platform := getPlatform()
	
	if cmd, exists := t.InstallCommands[platform]; exists {
		return fmt.Sprintf("To install %s on %s:\n  %s", t.Name, platform, cmd)
	}

	// Fallback to generic instructions
	var instructions strings.Builder
	instructions.WriteString(fmt.Sprintf("To install %s:\n", t.Name))
	for platform, cmd := range t.InstallCommands {
		instructions.WriteString(fmt.Sprintf("  %s: %s\n", platform, cmd))
	}
	
	return instructions.String()
}

// GetDetailedInfo returns comprehensive information about the tool
func (t *Tool) GetDetailedInfo() string {
	var info strings.Builder
	
	info.WriteString(fmt.Sprintf("Tool: %s\n", t.Name))
	info.WriteString(fmt.Sprintf("Description: %s\n", t.Description))
	info.WriteString(fmt.Sprintf("Website: %s\n", t.Website))
	info.WriteString(fmt.Sprintf("Required: %t\n", t.Required))
	
	if t.IsAvailable() {
		version, err := t.GetVersion()
		if err == nil {
			info.WriteString(fmt.Sprintf("Status: ✅ INSTALLED (%s)\n", strings.TrimSpace(version)))
		} else {
			info.WriteString("Status: ✅ INSTALLED (version check failed)\n")
		}
	} else {
		info.WriteString("Status: ❌ NOT FOUND\n")
		info.WriteString("\n")
		info.WriteString(t.GetInstallInstructions())
	}
	
	return info.String()
}

// DiagnosticInfo provides diagnostic information about tool availability
type DiagnosticInfo struct {
	Tool        *Tool
	Available   bool
	Version     string
	Error       error
	Alternative string // fallback option when tool is missing
}

// GetDiagnosticInfo returns detailed diagnostic information
func (t *Tool) GetDiagnosticInfo() DiagnosticInfo {
	info := DiagnosticInfo{
		Tool:      t,
		Available: t.IsAvailable(),
	}
	
	if info.Available {
		version, err := t.GetVersion()
		if err != nil {
			info.Error = err
		} else {
			info.Version = strings.TrimSpace(version)
		}
	} else {
		// Set fallback alternatives
		switch t.Name {
		case "fd":
			info.Alternative = "Standard file listing will be used (slower)"
		case "ripgrep":
			info.Alternative = "Content search will be unavailable"
		case "fzf":
			info.Alternative = "Basic numbered selection will be used"
		}
	}
	
	return info
}

// CheckDependencies checks if all required tools are available
func CheckDependencies(tools ...*Tool) []string {
	var missing []string
	
	for _, tool := range tools {
		if !tool.IsAvailable() {
			missing = append(missing, tool.Name)
		}
	}
	
	return missing
}

// GetMissingToolsMessage returns a formatted message about missing tools
func GetMissingToolsMessage(tools ...*Tool) string {
	missing := CheckDependencies(tools...)
	if len(missing) == 0 {
		return ""
	}

	var message strings.Builder
	message.WriteString(fmt.Sprintf("Missing required tools: %s\n\n", strings.Join(missing, ", ")))
	
	for _, tool := range tools {
		if !tool.IsAvailable() {
			message.WriteString(tool.GetInstallInstructions())
			message.WriteString("\n\n")
		}
	}
	
	return message.String()
}

// SystemDiagnostics provides comprehensive system and tool diagnostics
type SystemDiagnostics struct {
	Platform    string
	Tools       []DiagnosticInfo
	Summary     DiagnosticSummary
	Suggestions []string
}

// DiagnosticSummary provides an overview of tool availability
type DiagnosticSummary struct {
	TotalTools     int
	AvailableTools int
	MissingTools   int
	RequiredMissing int
	OptionalMissing int
}

// RunSystemDiagnostics performs comprehensive diagnostics on all tools
func RunSystemDiagnostics(tools ...*Tool) SystemDiagnostics {
	diagnostics := SystemDiagnostics{
		Platform: getPlatform(),
		Tools:    make([]DiagnosticInfo, 0, len(tools)),
		Summary: DiagnosticSummary{
			TotalTools: len(tools),
		},
	}
	
	// Collect diagnostic info for each tool
	for _, tool := range tools {
		info := tool.GetDiagnosticInfo()
		diagnostics.Tools = append(diagnostics.Tools, info)
		
		if info.Available {
			diagnostics.Summary.AvailableTools++
		} else {
			diagnostics.Summary.MissingTools++
			if tool.Required {
				diagnostics.Summary.RequiredMissing++
			} else {
				diagnostics.Summary.OptionalMissing++
			}
		}
	}
	
	// Generate suggestions
	diagnostics.Suggestions = generateDiagnosticSuggestions(&diagnostics)
	
	return diagnostics
}

// generateDiagnosticSuggestions creates actionable suggestions based on diagnostics
func generateDiagnosticSuggestions(diag *SystemDiagnostics) []string {
	var suggestions []string
	
	if diag.Summary.MissingTools == 0 {
		suggestions = append(suggestions, "✅ All tools are available! Your gman installation is fully optimized.")
		return suggestions
	}
	
	if diag.Summary.RequiredMissing > 0 {
		suggestions = append(suggestions, "⚠️  Install required tools to ensure full functionality")
	}
	
	if diag.Summary.OptionalMissing > 0 {
		suggestions = append(suggestions, "💡 Install optional tools for enhanced performance and features")
	}
	
	// Platform-specific suggestions
	switch diag.Platform {
	case "darwin":
		suggestions = append(suggestions, "💻 Consider using Homebrew for easy tool installation: brew install fd ripgrep fzf")
	case "linux":
		suggestions = append(suggestions, "🐧 Use your system package manager (apt/yum/pacman) to install missing tools")
	case "windows":
		suggestions = append(suggestions, "🪟 Use winget for convenient tool installation on Windows")
	}
	
	// Performance suggestions
	missingFd := false
	missingRg := false
	for _, info := range diag.Tools {
		if !info.Available {
			if info.Tool.Name == "fd" {
				missingFd = true
			}
			if info.Tool.Name == "ripgrep" {
				missingRg = true
			}
		}
	}
	
	if missingFd && missingRg {
		suggestions = append(suggestions, "🚀 Installing fd and ripgrep will significantly improve search performance")
	} else if missingFd {
		suggestions = append(suggestions, "⚡ Installing fd will make file search much faster")
	} else if missingRg {
		suggestions = append(suggestions, "🔍 Installing ripgrep will enable powerful content search capabilities")
	}
	
	return suggestions
}

// HasCriticalIssues returns true if there are critical missing tools
func (d *SystemDiagnostics) HasCriticalIssues() bool {
	return d.Summary.RequiredMissing > 0
}

// GetReadiness returns a readiness percentage (0-100)
func (d *SystemDiagnostics) GetReadiness() int {
	if d.Summary.TotalTools == 0 {
		return 100
	}
	return (d.Summary.AvailableTools * 100) / d.Summary.TotalTools
}

// getPlatform returns a simplified platform string
func getPlatform() string {
	switch runtime.GOOS {
	case "darwin":
		return "darwin"
	case "linux":
		return "linux"
	case "windows":
		return "windows"
	default:
		return "linux" // Default fallback to linux for other Unix-like systems
	}
}