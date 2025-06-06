package common

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

// TemplateContext represents the context passed to templates
type TemplateContext struct {
	Site    *Site                  `json:"site"`
	Page    *Page                  `json:"page"`
	Data    map[string]interface{} `json:"data"`
	Request *http.Request          `json:"-"`
}

// RenderEngine handles template rendering
type RenderEngine struct {
	siteManager *SiteManager
	funcMap     template.FuncMap
}

// NewRenderEngine creates a new render engine
func NewRenderEngine(siteManager *SiteManager) *RenderEngine {
	return &RenderEngine{
		siteManager: siteManager,
		funcMap:     createTemplateFuncMap(),
	}
}

// createTemplateFuncMap creates the function map for templates
func createTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		// String functions
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": func(s string) string {
			// Simple title case implementation
			if s == "" {
				return s
			}
			return strings.ToUpper(string(s[0])) + strings.ToLower(s[1:])
		},
		"trim":      strings.TrimSpace,
		"replace":   strings.ReplaceAll,
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,

		// Utility functions
		"dict": func(values ...interface{}) map[string]interface{} {
			if len(values)%2 != 0 {
				return nil
			}
			dict := make(map[string]interface{})
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					continue
				}
				dict[key] = values[i+1]
			}
			return dict
		},

		"slice": func(items ...interface{}) []interface{} {
			return items
		},

		"add": func(a, b int) int {
			return a + b
		},

		"sub": func(a, b int) int {
			return a - b
		},

		"mul": func(a, b int) int {
			return a * b
		},

		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},

		// Array/slice functions
		"first": func(slice []interface{}) interface{} {
			if len(slice) == 0 {
				return nil
			}
			return slice[0]
		},

		"last": func(slice []interface{}) interface{} {
			if len(slice) == 0 {
				return nil
			}
			return slice[len(slice)-1]
		},

		"len": func(v interface{}) int {
			switch val := v.(type) {
			case []interface{}:
				return len(val)
			case []string:
				return len(val)
			case string:
				return len(val)
			case map[string]interface{}:
				return len(val)
			default:
				return 0
			}
		},

		// Default value function
		"default": func(defaultValue, value interface{}) interface{} {
			if value == nil || value == "" {
				return defaultValue
			}
			return value
		},
	}
}

// renderStaticPage renders a static page with embedded HTML content
func (re *RenderEngine) renderStaticPage(w http.ResponseWriter, r *http.Request, site *Site, page *Page) error {
	// Create template context
	context := &TemplateContext{
		Site:    site,
		Page:    page,
		Data:    make(map[string]interface{}),
		Request: r,
	}

	// Add page custom data to context
	if page.Meta.CustomData != nil {
		for key, value := range page.Meta.CustomData {
			context.Data[key] = value
		}
	}

	// Also add custom data to the Meta for template access
	context.Page.Meta.CustomData = page.Meta.CustomData

	// Parse the HTML content as a template
	tmpl, err := template.New("page").Funcs(re.funcMap).Parse(page.Content)
	if err != nil {
		return fmt.Errorf("error parsing page HTML template: %w", err)
	}

	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Render template directly
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, context); err != nil {
		return fmt.Errorf("error executing page template: %w", err)
	}

	// Write response
	if _, err := w.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("error writing response: %w", err)
	}

	return nil
}

// RenderPage renders a page using its template
func (re *RenderEngine) RenderPage(w http.ResponseWriter, r *http.Request, site *Site, page *Page) error {
	// Check if page requires authentication
	if page.Meta.RequireAuth {
		// TODO: Implement authentication check
		// For now, we'll skip this check
	}

	// Check if page is draft (only show to authenticated admin users)
	if page.Meta.IsDraft {
		// TODO: Implement admin check
		// For now, we'll show drafts
	}

	// For static pages with embedded HTML, render directly
	if page.Meta.IsStatic && page.Content != "" {
		return re.renderStaticPage(w, r, site, page)
	}

	// For dynamic pages or template-based pages, use the template system
	// Determine template to use
	templateName := page.Meta.Template
	if templateName == "" {
		templateName = "default"
	}

	// Load template
	tmpl, err := re.loadTemplate(site, templateName)
	if err != nil {
		return fmt.Errorf("error loading template: %w", err)
	}

	// Create template context
	context := &TemplateContext{
		Site:    site,
		Page:    page,
		Data:    make(map[string]interface{}),
		Request: r,
	}

	// Add page custom data to context
	if page.Meta.CustomData != nil {
		for key, value := range page.Meta.CustomData {
			context.Data[key] = value
		}
	}

	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Render template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, context); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	// Write response
	if _, err := w.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("error writing response: %w", err)
	}

	return nil
}

// RenderTemplate renders a specific template with custom context
func (re *RenderEngine) RenderTemplate(w http.ResponseWriter, r *http.Request, site *Site, templateName string, data map[string]interface{}) error {
	// Load template
	tmpl, err := re.loadTemplate(site, templateName)
	if err != nil {
		return fmt.Errorf("error loading template: %w", err)
	}

	// Create template context
	context := &TemplateContext{
		Site:    site,
		Page:    nil,
		Data:    data,
		Request: r,
	}

	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Render template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, context); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	// Write response
	if _, err := w.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("error writing response: %w", err)
	}

	return nil
}

// loadTemplate loads and parses a template with all its dependencies
func (re *RenderEngine) loadTemplate(site *Site, templateName string) (*template.Template, error) {
	// Create base template with function map
	tmpl := template.New("").Funcs(re.funcMap)

	// Load layout templates first
	layoutGlob := filepath.Join(site.LayoutPath, "*.html")
	if layouts, err := filepath.Glob(layoutGlob); err == nil && len(layouts) > 0 {
		tmpl, err = tmpl.ParseGlob(layoutGlob)
		if err != nil {
			return nil, fmt.Errorf("error parsing layout templates: %w", err)
		}
	}

	// Load snippets
	snippetsGlob := filepath.Join(site.SnippetsPath, "*.html")
	if snippets, err := filepath.Glob(snippetsGlob); err == nil && len(snippets) > 0 {
		tmpl, err = tmpl.ParseGlob(snippetsGlob)
		if err != nil {
			return nil, fmt.Errorf("error parsing snippet templates: %w", err)
		}
	}

	// Load sections
	sectionsGlob := filepath.Join(site.SectionsPath, "*.html")
	if sections, err := filepath.Glob(sectionsGlob); err == nil && len(sections) > 0 {
		tmpl, err = tmpl.ParseGlob(sectionsGlob)
		if err != nil {
			return nil, fmt.Errorf("error parsing section templates: %w", err)
		}
	}

	// Load blocks
	blocksGlob := filepath.Join(site.BlocksPath, "*.html")
	if blocks, err := filepath.Glob(blocksGlob); err == nil && len(blocks) > 0 {
		tmpl, err = tmpl.ParseGlob(blocksGlob)
		if err != nil {
			return nil, fmt.Errorf("error parsing block templates: %w", err)
		}
	}

	// Load the main template
	templatePath := filepath.Join(site.TemplatesPath, templateName+".html")
	tmpl, err := tmpl.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("error parsing main template %s: %w", templateName, err)
	}

	return tmpl, nil
}

// RenderError renders an error page
func (re *RenderEngine) RenderError(w http.ResponseWriter, r *http.Request, site *Site, statusCode int, message string) {
	// Set status code
	w.WriteHeader(statusCode)

	// Try to load error template
	tmpl, err := re.loadTemplate(site, fmt.Sprintf("error-%d", statusCode))
	if err != nil {
		// Fallback to generic error template
		tmpl, err = re.loadTemplate(site, "error")
		if err != nil {
			// Last resort: plain text response
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			fmt.Fprintf(w, "Error %d: %s", statusCode, message)
			return
		}
	}

	// Create error context
	context := &TemplateContext{
		Site:    site,
		Page:    nil,
		Request: r,
		Data: map[string]interface{}{
			"status_code": statusCode,
			"message":     message,
		},
	}

	// Render error template
	if err := tmpl.Execute(w, context); err != nil {
		// Last resort: plain text response
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "Error %d: %s", statusCode, message)
	}
}
