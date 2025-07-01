package tpl

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"wispy-core/common"
)

// TemplateEngine handles template loading and rendering
type templateEngine struct {
	mu        sync.RWMutex
	templates map[string]*template.Template
	layoutDir string
	pagesDir  string
	trie      *common.Trie
}

// TemplateData represents the data passed to templates
type TemplateData struct {
	Title       string
	Description string
	Site        SiteData
	Content     template.HTML
	Data        map[string]interface{}
}

// SiteData represents site information for templates
type SiteData struct {
	Name    string
	Domain  string
	BaseURL string
}

// RenderResult contains the rendered HTML and generated CSS
type RenderResult struct {
	HTML string
	CSS  string
}

type TemplateEngine interface {
	LoadTemplate(templatePath string) (*template.Template, error)
	RenderTemplate(templatePath string, data TemplateData) (string, error)
	RenderTemplateTo(w io.Writer, templatePath string, data TemplateData) error
	ScanPages() ([]string, error)
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine(layoutDir, pagesDir string) TemplateEngine {
	return &templateEngine{
		templates: make(map[string]*template.Template),
		layoutDir: layoutDir,
		pagesDir:  pagesDir,
		trie:      common.NewTrie(),
	}
}

func (te *templateEngine) TemplateEngine() TemplateEngine {
	return te
}

func (te *templateEngine) GetTrie() *common.Trie {
	return te.trie
}

// LoadTemplate loads a template from the pages directory
func (te *templateEngine) LoadTemplate(templatePath string) (*template.Template, error) {
	te.mu.Lock()
	defer te.mu.Unlock()

	// Check if template is already loaded
	if tmpl, exists := te.templates[templatePath]; exists {
		return tmpl, nil
	}

	// Load the default layout
	layoutPath := filepath.Join(te.layoutDir, "default.html")
	layoutContent, err := os.ReadFile(layoutPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read layout: %w", err)
	}

	// Load the page template
	pageContent, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	// Create template with layout and page content
	tmpl := template.New("page")

	// Parse layout first
	tmpl, err = tmpl.Parse(string(layoutContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse layout: %w", err)
	}

	// Define content template
	contentTemplate := fmt.Sprintf(`{{ define "content" }}%s{{ end }}`, string(pageContent))
	tmpl, err = tmpl.Parse(contentTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse page content: %w", err)
	}

	// Cache the template
	te.templates[templatePath] = tmpl

	return tmpl, nil
}

// RenderTemplate renders a template with the given data
func (te *templateEngine) RenderTemplate(templatePath string, data TemplateData) (string, error) {
	tmpl, err := te.LoadTemplate(templatePath)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// RenderTemplateTo renders a template directly to a writer
func (te *templateEngine) RenderTemplateTo(w io.Writer, templatePath string, data TemplateData) error {
	tmpl, err := te.LoadTemplate(templatePath)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// ScanPages scans the pages directory and returns a list of available pages
func (te *templateEngine) ScanPages() ([]string, error) {
	var pages []string

	err := filepath.Walk(te.pagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".html") {
			// Get relative path from pages directory
			relPath, err := filepath.Rel(te.pagesDir, path)
			if err != nil {
				return err
			}
			pages = append(pages, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan pages directory: %w", err)
	}

	return pages, nil
}

// PathToRoute converts a file path to a URL route
func PathToRoute(pagePath string) string {
	// Remove .html extension
	route := strings.TrimSuffix(pagePath, ".html")

	// Convert file separators to URL separators
	route = strings.ReplaceAll(route, "\\", "/")

	// Handle index files
	if strings.HasSuffix(route, "/index") {
		route = strings.TrimSuffix(route, "/index")
	}

	// Ensure route starts with /
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}

	// Handle root index
	if route == "/index" {
		route = "/"
	}

	return route
}
