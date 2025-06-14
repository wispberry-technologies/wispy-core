// Package tail provides a Tailwind CSS v4 compatible compiler
// that takes HTML content with Tailwind classes and outputs
// optimized CSS.
package tail

import (
	"fmt"
	"os"
	"path/filepath"
)

// TailwindConfig represents the configuration for Tailwind CSS compilation
type TailwindConfig struct {
	// Theme configuration for Tailwind
	Theme map[string]interface{}
	// Additional utility classes
	Utilities map[string]string
	// Version of Tailwind to use
	Version string
}

// DefaultConfig returns the default Tailwind configuration
func DefaultConfig() *TailwindConfig {
	return &TailwindConfig{
		Theme:     make(map[string]interface{}),
		Utilities: make(map[string]string),
		Version:   "v4.1.10",
	}
}

// CompileHTML takes HTML content and generates Tailwind CSS
func CompileHTML(htmlContent string) (string, error) {
	return CompileHTMLWithConfig(htmlContent, DefaultConfig())
}

// CompileHTMLWithConfig takes HTML content and a config, and generates Tailwind CSS
func CompileHTMLWithConfig(htmlContent string, config *TailwindConfig) (string, error) {
	// Extract classes from HTML
	classes, err := extractClasses(htmlContent)
	if err != nil {
		return "", fmt.Errorf("error extracting classes: %w", err)
	}

	// Sort classes according to Tailwind's order
	sortedClasses := sortClasses(classes)

	// Generate CSS
	css, err := generateCSS(sortedClasses)
	if err != nil {
		return "", fmt.Errorf("error generating CSS: %w", err)
	}

	return css, nil
}

// CompileHTMLWithDaisyUI takes HTML content and generates Tailwind CSS with daisyUI support
func CompileHTMLWithDaisyUI(htmlContent string) (string, error) {
	config := DefaultConfig()
	// Add daisyUI config
	config.Utilities["daisyui"] = "enabled"

	return CompileHTMLWithConfig(htmlContent, config)
}

// CompileFile takes an HTML file path and generates Tailwind CSS
func CompileFile(filePath string) (string, error) {
	content, err := readFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return CompileHTML(content)
}

// CompileAndWriteFile takes an HTML file, compiles it to CSS, and writes to output path
func CompileAndWriteFile(inputPath, outputPath string) error {
	css, err := CompileFile(inputPath)
	if err != nil {
		return err
	}

	return writeFile(outputPath, css)
}

// Helper functions
func readFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func writeFile(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}
