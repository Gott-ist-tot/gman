package di

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// MigrationReport represents the migration analysis report
type MigrationReport struct {
	FilesAnalyzed         int
	FilesNeedingMigration int
	ManualInstantiations  []ManualInstantiation
	Recommendations       []string
}

// ManualInstantiation represents a found manual instantiation
type ManualInstantiation struct {
	File       string
	Line       int
	Column     int
	Pattern    string
	Suggestion string
}

// AnalyzeDependencyUsage analyzes the codebase for manual dependency instantiation
func AnalyzeDependencyUsage(rootPath string) (*MigrationReport, error) {
	report := &MigrationReport{
		ManualInstantiations: make([]ManualInstantiation, 0),
		Recommendations:      make([]string, 0),
	}

	// Patterns to detect manual instantiation
	patterns := []struct {
		Regex       *regexp.Regexp
		Replacement string
		Description string
	}{
		{
			Regex:       regexp.MustCompile(`config\.NewManager\(\)`),
			Replacement: `di.ConfigManager()`,
			Description: "Replace di.ConfigManager() with DI container",
		},
		{
			Regex:       regexp.MustCompile(`git\.NewManager\(\)`),
			Replacement: `di.GitManager()`,
			Description: "Replace di.GitManager() with DI container",
		},
		{
			Regex:       regexp.MustCompile(`git\.NewGitManager\(\)`),
			Replacement: `di.GitFacade()`,
			Description: "Replace di.GitFacade() with DI container",
		},
	}

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files and vendor/test directories
		if !strings.HasSuffix(path, ".go") ||
			strings.Contains(path, "vendor/") ||
			strings.Contains(path, ".git/") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		report.FilesAnalyzed++
		fileContent := string(content)
		lines := strings.Split(fileContent, "\n")

		foundInFile := false
		for i, line := range lines {
			for _, pattern := range patterns {
				if pattern.Regex.MatchString(line) {
					foundInFile = true

					// Find column position
					match := pattern.Regex.FindStringIndex(line)
					column := 0
					if len(match) > 0 {
						column = match[0]
					}

					report.ManualInstantiations = append(report.ManualInstantiations, ManualInstantiation{
						File:       path,
						Line:       i + 1,
						Column:     column + 1,
						Pattern:    pattern.Regex.String(),
						Suggestion: pattern.Replacement,
					})
				}
			}
		}

		if foundInFile {
			report.FilesNeedingMigration++
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Generate recommendations
	report.generateRecommendations()

	return report, nil
}

// generateRecommendations creates actionable recommendations
func (r *MigrationReport) generateRecommendations() {
	if len(r.ManualInstantiations) == 0 {
		r.Recommendations = append(r.Recommendations, "✅ No manual instantiations found - DI migration complete!")
		return
	}

	// Group by file for better recommendations
	fileGroups := make(map[string][]ManualInstantiation)
	for _, inst := range r.ManualInstantiations {
		fileGroups[inst.File] = append(fileGroups[inst.File], inst)
	}

	r.Recommendations = append(r.Recommendations, fmt.Sprintf("📊 Found %d manual instantiations in %d files", len(r.ManualInstantiations), len(fileGroups)))
	r.Recommendations = append(r.Recommendations, "")
	r.Recommendations = append(r.Recommendations, "🔧 Priority migration files:")

	for file, instances := range fileGroups {
		relPath := strings.TrimPrefix(file, "/Users/henrykuo/tui/programming/cli-tool/")
		r.Recommendations = append(r.Recommendations, fmt.Sprintf("  • %s (%d instances)", relPath, len(instances)))

		for _, inst := range instances {
			r.Recommendations = append(r.Recommendations, fmt.Sprintf("    Line %d:%d - %s → %s", inst.Line, inst.Column, inst.Pattern, inst.Suggestion))
		}
	}

	r.Recommendations = append(r.Recommendations, "")
	r.Recommendations = append(r.Recommendations, "🎯 Migration Steps:")
	r.Recommendations = append(r.Recommendations, "  1. Add 'gman/internal/di' import to files")
	r.Recommendations = append(r.Recommendations, "  2. Replace manual instantiations with DI calls")
	r.Recommendations = append(r.Recommendations, "  3. Initialize DI container in main.go or root command")
	r.Recommendations = append(r.Recommendations, "  4. Run tests to ensure functionality")
}

// PrintReport prints the migration report in a readable format
func (r *MigrationReport) PrintReport() {
	fmt.Println("🔍 Dependency Injection Migration Analysis")
	fmt.Println("==========================================")
	fmt.Printf("Files analyzed: %d\n", r.FilesAnalyzed)
	fmt.Printf("Files needing migration: %d\n", r.FilesNeedingMigration)
	fmt.Printf("Total manual instantiations: %d\n", len(r.ManualInstantiations))
	fmt.Println()

	for _, recommendation := range r.Recommendations {
		fmt.Println(recommendation)
	}
}

// ApplyAutomaticMigration performs automatic migration where safe
func ApplyAutomaticMigration(rootPath string, dryRun bool) error {
	report, err := AnalyzeDependencyUsage(rootPath)
	if err != nil {
		return err
	}

	if len(report.ManualInstantiations) == 0 {
		fmt.Println("✅ No migration needed - all files already use DI container")
		return nil
	}

	fmt.Printf("🔄 %s migration for %d files...\n", map[bool]string{true: "DRY-RUN", false: "Applying"}[dryRun], report.FilesNeedingMigration)

	// Group by file for efficient processing
	fileGroups := make(map[string][]ManualInstantiation)
	for _, inst := range report.ManualInstantiations {
		fileGroups[inst.File] = append(fileGroups[inst.File], inst)
	}

	for filePath, instances := range fileGroups {
		if err := migrateFile(filePath, instances, dryRun); err != nil {
			fmt.Printf("❌ Failed to migrate %s: %v\n", filePath, err)
		} else {
			relPath := strings.TrimPrefix(filePath, rootPath+"/")
			fmt.Printf("✅ Migrated %s (%d changes)\n", relPath, len(instances))
		}
	}

	return nil
}

// migrateFile migrates a single file
func migrateFile(filePath string, instances []ManualInstantiation, dryRun bool) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	fileContent := string(content)

	// Apply replacements (in reverse order to maintain line numbers)
	for i := len(instances) - 1; i >= 0; i-- {
		inst := instances[i]

		// Simple string replacement for patterns
		switch inst.Pattern {
		case `config\.NewManager\(\)`:
			fileContent = strings.ReplaceAll(fileContent, "config.NewManager()", "di.ConfigManager()")
		case `git\.NewManager\(\)`:
			fileContent = strings.ReplaceAll(fileContent, "git.NewManager()", "di.GitManager()")
		case `git\.NewGitManager\(\)`:
			fileContent = strings.ReplaceAll(fileContent, "git.NewGitManager()", "di.GitFacade()")
		}
	}

	// Ensure DI import is present
	if strings.Contains(fileContent, "di.") && !strings.Contains(fileContent, `"gman/internal/di"`) {
		fileContent = addDIImport(fileContent)
	}

	if !dryRun {
		return os.WriteFile(filePath, []byte(fileContent), 0644)
	}

	return nil
}

// addDIImport adds the DI import to a Go file
func addDIImport(content string) string {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		// Fallback to simple string replacement
		importBlock := `import (`
		if strings.Contains(content, importBlock) {
			// Add to existing import block
			replacement := importBlock + "\n\t\"gman/internal/di\""
			content = strings.Replace(content, importBlock, replacement, 1)
		}
		return content
	}

	// Check if DI import already exists
	for _, imp := range node.Imports {
		if imp.Path.Value == `"gman/internal/di"` {
			return content // Already imported
		}
	}

	// Insert DI import using AST manipulation would be complex,
	// so we'll use a simple approach for now
	if strings.Contains(content, `"gman/internal/`) {
		// Add after existing gman imports
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			if strings.Contains(line, `"gman/internal/`) && !strings.Contains(line, `"gman/internal/di"`) {
				// Insert DI import after this line
				newLines := make([]string, 0, len(lines)+1)
				newLines = append(newLines, lines[:i+1]...)
				newLines = append(newLines, "\t\"gman/internal/di\"")
				newLines = append(newLines, lines[i+1:]...)
				return strings.Join(newLines, "\n")
			}
		}
	}

	return content
}
