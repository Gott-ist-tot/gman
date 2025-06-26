package external

import (
	"fmt"
	"os/exec"
	"strings"
)

// Tool represents an external command-line tool
type Tool struct {
	Name            string
	Command         string
	MinVersion      string
	InstallCommands map[string]string // platform -> install command
	CheckCmd        []string          // command to check if tool exists
}

// Common external tools used by gman
var (
	FD = &Tool{
		Name:    "fd",
		Command: "fd",
		InstallCommands: map[string]string{
			"darwin":  "brew install fd",
			"linux":   "apt install fd-find || yum install fd-find || pacman -S fd",
			"windows": "winget install sharkdp.fd",
		},
		CheckCmd: []string{"fd", "--version"},
	}

	RipGrep = &Tool{
		Name:    "ripgrep",
		Command: "rg",
		InstallCommands: map[string]string{
			"darwin":  "brew install ripgrep",
			"linux":   "apt install ripgrep || yum install ripgrep || pacman -S ripgrep",
			"windows": "winget install BurntSushi.ripgrep.MSVC",
		},
		CheckCmd: []string{"rg", "--version"},
	}

	FZF = &Tool{
		Name:    "fzf",
		Command: "fzf",
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

// getPlatform returns a simplified platform string
func getPlatform() string {
	// This is a simplified implementation
	// In a production system, you might want to use runtime.GOOS
	// and more sophisticated platform detection
	return "darwin" // For now, since we're on macOS
}