package external

import (
	"fmt"

	"github.com/fatih/color"
)

// FileSearcher interface defines the contract for file searching
type FileSearcher interface {
	SearchFiles(pattern string, repositories map[string]string, groupFilter string) ([]FileResult, error)
	FormatForFZF(results []FileResult) string
	ParseFZFSelection(selection string, results []FileResult) (*FileResult, error)
}

// SmartSearcher automatically selects the best available search strategy
type SmartSearcher struct {
	primarySearcher   FileSearcher
	fallbackSearcher  FileSearcher
	diagnostics       *SystemDiagnostics
	verbose           bool
}

// NewSmartSearcher creates a new intelligent search manager
func NewSmartSearcher(verbose bool) *SmartSearcher {
	// Run diagnostics to determine available tools
	allTools := []*Tool{FD, RipGrep, FZF}
	diagnostics := RunSystemDiagnostics(allTools...)
	
	// Select primary searcher based on tool availability
	var primarySearcher FileSearcher
	if FD.IsAvailable() {
		primarySearcher = NewFDSearcher()
	} else {
		primarySearcher = nil // Will use fallback
	}
	
	fallbackSearcher := NewFallbackSearcher()
	
	return &SmartSearcher{
		primarySearcher:  primarySearcher,
		fallbackSearcher: fallbackSearcher,
		diagnostics:      &diagnostics,
		verbose:          verbose,
	}
}

// SearchFiles performs intelligent file search with automatic fallback
func (ss *SmartSearcher) SearchFiles(pattern string, repositories map[string]string, groupFilter string) ([]FileResult, error) {
	// Try primary searcher first (if available)
	if ss.primarySearcher != nil {
		if ss.verbose {
			fmt.Printf("🔍 Using %s for file search...\n", color.CyanString("fd"))
		}
		
		results, err := ss.primarySearcher.SearchFiles(pattern, repositories, groupFilter)
		if err == nil {
			return results, nil
		}
		
		if ss.verbose {
			fmt.Printf("⚠️  Primary search failed: %v\n", err)
			fmt.Printf("🔄 Falling back to %s...\n", color.YellowString("basic file search"))
		}
	} else {
		if ss.verbose {
			ss.showToolMissingMessage("fd", "file search")
		}
	}
	
	// Use fallback searcher
	results, err := ss.fallbackSearcher.SearchFiles(pattern, repositories, groupFilter)
	if err != nil {
		return nil, fmt.Errorf("both primary and fallback search failed: %w", err)
	}
	
	return results, nil
}

// FormatForFZF formats results using the appropriate searcher
func (ss *SmartSearcher) FormatForFZF(results []FileResult) string {
	if ss.primarySearcher != nil {
		return ss.primarySearcher.FormatForFZF(results)
	}
	return ss.fallbackSearcher.FormatForFZF(results)
}

// ParseFZFSelection parses selection using the appropriate searcher
func (ss *SmartSearcher) ParseFZFSelection(selection string, results []FileResult) (*FileResult, error) {
	if ss.primarySearcher != nil {
		return ss.primarySearcher.ParseFZFSelection(selection, results)
	}
	return ss.fallbackSearcher.ParseFZFSelection(selection, results)
}

// SelectFromResults provides intelligent result selection with fallback
func (ss *SmartSearcher) SelectFromResults(results []FileResult, prompt string) (*FileResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no search results found")
	}
	
	// Try to use fzf for interactive selection
	if FZF.IsAvailable() {
		if ss.verbose {
			fmt.Printf("🎯 Using %s for interactive selection...\n", color.CyanString("fzf"))
		}
		
		return ss.selectWithFZF(results, prompt)
	}
	
	// Fall back to basic selection
	if ss.verbose {
		ss.showToolMissingMessage("fzf", "interactive selection")
	}
	
	basicSelector := NewBasicSelector()
	return basicSelector.SelectFromResults(results, prompt)
}

// selectWithFZF uses fzf for interactive file selection
func (ss *SmartSearcher) selectWithFZF(results []FileResult, prompt string) (*FileResult, error) {
	// TODO: Implement fzf integration
	// For now, fall back to basic selection
	if ss.verbose {
		fmt.Printf("🔄 FZF integration not yet implemented, using basic selection...\n")
	}
	
	basicSelector := NewBasicSelector()
	return basicSelector.SelectFromResults(results, prompt)
}

// showToolMissingMessage displays a helpful message about missing tools
func (ss *SmartSearcher) showToolMissingMessage(toolName, feature string) {
	fmt.Printf("⚠️  %s is not available for enhanced %s\n", 
		color.YellowString(toolName), feature)
	
	// Find the tool in diagnostics
	for _, info := range ss.diagnostics.Tools {
		if info.Tool.Name == toolName {
			fmt.Printf("   Fallback: %s\n", color.CyanString(info.Alternative))
			if ss.verbose {
				fmt.Printf("   Install: %s\n", color.BlueString(info.Tool.GetInstallInstructions()))
			}
			break
		}
	}
}

// GetDiagnostics returns the current system diagnostics
func (ss *SmartSearcher) GetDiagnostics() *SystemDiagnostics {
	return ss.diagnostics
}

// ShowOptimizationTips displays suggestions for improving search performance
func (ss *SmartSearcher) ShowOptimizationTips() {
	if ss.diagnostics.GetReadiness() == 100 {
		fmt.Printf("✅ %s: All search tools are optimally configured!\n", 
			color.GreenString("OPTIMIZED"))
		return
	}
	
	fmt.Printf("💡 %s:\n", color.CyanString("OPTIMIZATION TIPS"))
	
	for _, suggestion := range ss.diagnostics.Suggestions {
		fmt.Printf("   %s\n", suggestion)
	}
	
	// Specific performance impact information
	if !FD.IsAvailable() {
		fmt.Printf("   📊 Performance impact: File search is ~5-10x slower without fd\n")
	}
	
	if !FZF.IsAvailable() {
		fmt.Printf("   📊 UX impact: Interactive selection is less user-friendly without fzf\n")
	}
}