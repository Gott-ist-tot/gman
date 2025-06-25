package fzf

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Finder handles fzf integration
type Finder struct {
	fzfPath string
}

// NewFinder creates a new fzf finder instance
func NewFinder() (*Finder, error) {
	fzfPath, err := exec.LookPath("fzf")
	if err != nil {
		return nil, fmt.Errorf("fzf not found in PATH. Please install fzf: %w", err)
	}

	return &Finder{
		fzfPath: fzfPath,
	}, nil
}

// Options represents fzf configuration options
type Options struct {
	Prompt       string   // Custom prompt
	Header       string   // Header text
	Preview      string   // Preview command
	PreviewSize  string   // Preview window size (e.g., "50%", "70%:wrap")
	Multi        bool     // Allow multiple selections
	Height       string   // Height of fzf window (e.g., "40%", "20")
	Layout       string   // Layout (default, reverse, reverse-list)
	Border       bool     // Show border
	InitialQuery string   // Initial search query
	BindKeys     []string // Key bindings (e.g., "ctrl-o:execute(echo {+})")
}

// DefaultFileOptions returns default options for file searching
func DefaultFileOptions() Options {
	return Options{
		Prompt:      "Select file> ",
		Header:      "Press ENTER to select, Ctrl-C to cancel",
		PreviewSize: "70%:wrap",
		Height:      "80%",
		Layout:      "reverse",
		Border:      true,
	}
}

// DefaultCommitOptions returns default options for commit searching
func DefaultCommitOptions() Options {
	return Options{
		Prompt:      "Select commit> ",
		Header:      "Press ENTER to select, Ctrl-C to cancel",
		PreviewSize: "70%:wrap",
		Height:      "80%",
		Layout:      "reverse",
		Border:      true,
	}
}

// Find launches fzf with the given input and options
func (f *Finder) Find(input []string, opts Options) ([]string, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("no input provided to fzf")
	}

	// Build fzf command arguments
	args := f.buildArgs(opts)

	// Create fzf command
	cmd := exec.Command(f.fzfPath, args...)
	
	// Set up stdin pipe
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Set up stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Connect stderr to our stderr for error messages
	cmd.Stderr = os.Stderr

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start fzf: %w", err)
	}

	// Send input to fzf
	go func() {
		defer stdin.Close()
		for _, line := range input {
			fmt.Fprintln(stdin, line)
		}
	}()

	// Read output from fzf
	var results []string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			results = append(results, line)
		}
	}

	// Wait for fzf to complete
	err = cmd.Wait()
	if err != nil {
		// Check if it's just a user cancellation (exit code 130)
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 130 || exitError.ExitCode() == 1 {
				// User canceled or no selection made
				return nil, fmt.Errorf("selection canceled")
			}
		}
		return nil, fmt.Errorf("fzf command failed: %w", err)
	}

	if scanner.Err() != nil {
		return nil, fmt.Errorf("error reading fzf output: %w", scanner.Err())
	}

	return results, nil
}

// FindSingle is a convenience method for single selection
func (f *Finder) FindSingle(input []string, opts Options) (string, error) {
	opts.Multi = false // Ensure single selection
	results, err := f.Find(input, opts)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", fmt.Errorf("no selection made")
	}

	return results[0], nil
}

// buildArgs builds command line arguments for fzf
func (f *Finder) buildArgs(opts Options) []string {
	var args []string

	// Basic options
	if opts.Prompt != "" {
		args = append(args, "--prompt", opts.Prompt)
	}

	if opts.Header != "" {
		args = append(args, "--header", opts.Header)
	}

	if opts.Height != "" {
		args = append(args, "--height", opts.Height)
	}

	if opts.Layout != "" {
		args = append(args, "--layout", opts.Layout)
	}

	if opts.Border {
		args = append(args, "--border")
	}

	if opts.Multi {
		args = append(args, "--multi")
	}

	if opts.InitialQuery != "" {
		args = append(args, "--query", opts.InitialQuery)
	}

	// Preview options
	if opts.Preview != "" {
		args = append(args, "--preview", opts.Preview)
		
		if opts.PreviewSize != "" {
			args = append(args, "--preview-window", opts.PreviewSize)
		}
	}

	// Key bindings
	for _, binding := range opts.BindKeys {
		args = append(args, "--bind", binding)
	}

	// Additional useful options
	args = append(args, 
		"--ansi",           // Support ANSI color codes
		"--no-sort",        // Don't sort, maintain original order
		"--exact",          // Exact matching by default
		"--cycle",          // Enable cycling through results
		"--info=inline",    // Show info inline
	)

	return args
}

// IsAvailable checks if fzf is available
func IsAvailable() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}

// GetInstallInstructions returns installation instructions for fzf
func GetInstallInstructions() string {
	return `fzf is not installed or not found in PATH.

Please install fzf:

macOS (using Homebrew):
  brew install fzf

Ubuntu/Debian:
  sudo apt install fzf

CentOS/RHEL/Fedora:
  sudo dnf install fzf    # Fedora
  sudo yum install fzf    # CentOS/RHEL

From source:
  git clone --depth 1 https://github.com/junegunn/fzf.git ~/.fzf
  ~/.fzf/install

For more information, visit: https://github.com/junegunn/fzf`
}

// RunWithPreview runs fzf with a custom preview command
func (f *Finder) RunWithPreview(input []string, previewCmd string, opts Options) (string, error) {
	opts.Preview = previewCmd
	return f.FindSingle(input, opts)
}

// TestConnection tests if fzf is working correctly
func (f *Finder) TestConnection() error {
	testInput := []string{"test1", "test2", "test3"}
	testOpts := Options{
		Prompt:       "Test> ",
		Header:       "This is a test - press Ctrl-C to cancel",
		Height:       "10",
		InitialQuery: "test",
	}

	// Run fzf with test input but don't wait for user input
	// This just tests if fzf can start correctly
	cmd := exec.Command(f.fzfPath, f.buildArgs(testOpts)...)
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start fzf: %w", err)
	}

	// Send some test data and close
	go func() {
		defer stdin.Close()
		for _, line := range testInput {
			fmt.Fprintln(stdin, line)
		}
	}()

	// Kill the process after a short time (we just want to test if it starts)
	go func() {
		// Give it a moment to start
		// time.Sleep(100 * time.Millisecond)
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// Wait for the process to finish (or be killed)
	err = cmd.Wait()
	
	// If the process was killed, that's expected and means fzf is working
	if err != nil {
		if strings.Contains(err.Error(), "killed") || strings.Contains(err.Error(), "signal") {
			return nil // This is expected
		}
		return fmt.Errorf("fzf test failed: %w", err)
	}

	return nil
}

// CreateTempPreviewScript creates a temporary script for complex preview commands
func CreateTempPreviewScript(script string) (string, func(), error) {
	tempFile, err := os.CreateTemp("", "gman-preview-*.sh")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	// Write script content
	_, err = tempFile.WriteString("#!/bin/bash\n" + script + "\n")
	if err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return "", nil, fmt.Errorf("failed to write script: %w", err)
	}

	// Make executable
	tempFile.Close()
	err = os.Chmod(tempFile.Name(), 0755)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", nil, fmt.Errorf("failed to make script executable: %w", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.Remove(tempFile.Name())
	}

	return tempFile.Name(), cleanup, nil
}