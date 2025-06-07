package common

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
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
	siteManager   *SiteManager
	apiDispatcher APIDispatcher
	funcMap       template.FuncMap
}

// NewRenderEngine creates a new render engine
func NewRenderEngine(siteManager *SiteManager) *RenderEngine {
	return &RenderEngine{
		siteManager:   siteManager,
		apiDispatcher: nil, // Will be set later via SetAPIDispatcher
		funcMap:       nil, // Will be created when API dispatcher is set
	}
}

// SetAPIDispatcher sets the API dispatcher and creates the function map
func (re *RenderEngine) SetAPIDispatcher(dispatcher APIDispatcher) {
	re.apiDispatcher = dispatcher
	re.funcMap = re.createTemplateFuncMap()
}

// RenderPage renders a page using its template and layout
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
			// Also add to page meta for template access
			context.Page.Meta.CustomData[key] = value
		}
	}

	// Determine layout to use
	layoutName := page.Meta.Layout
	if layoutName == "" {
		layoutName = "default"
	}

	// Create a function map with request context
	funcMap := re.createTemplateFuncMapWithRequest(r)

	// Create a new template with our function map and load all template files
	tmpl := template.New("").Funcs(funcMap)

	// Ensure all template directories exist
	templateDirs := []string{
		site.LayoutPath,
		site.SnippetsPath,
		site.SectionsPath,
		site.BlocksPath,
	}

	for _, dir := range templateDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create template directory %s: %w", dir, err)
		}
	}

	// Load layout templates
	layouts, err := SecureGlob("layout/*.html")
	if err == nil && len(layouts) > 0 {
		tmpl, err = tmpl.ParseFiles(layouts...)
		if err != nil {
			return fmt.Errorf("error parsing layout templates: %w", err)
		}
	}

	// Load snippets
	snippets, err := SecureGlob("snippets/*.html")
	if err == nil && len(snippets) > 0 {
		tmpl, err = tmpl.ParseFiles(snippets...)
		if err != nil {
			return fmt.Errorf("error parsing snippet templates: %w", err)
		}
	}

	// Load sections
	sections, err := SecureGlob("sections/*.html")
	if err == nil && len(sections) > 0 {
		tmpl, err = tmpl.ParseFiles(sections...)
		if err != nil {
			return fmt.Errorf("error parsing section templates: %w", err)
		}
	}

	// Load blocks
	blocks, err := SecureGlob("blocks/*.html")
	if err == nil && len(blocks) > 0 {
		tmpl, err = tmpl.ParseFiles(blocks...)
		if err != nil {
			return fmt.Errorf("error parsing block templates: %w", err)
		}
	}

	// Parse the page content last (contains the page-content definition)
	tmpl, err = tmpl.Parse(page.Content)
	if err != nil {
		return fmt.Errorf("error parsing page content: %w", err)
	}

	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute the layout template (which will call the page-content block)
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, layoutName+".html", context); err != nil {
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
	tmpl, err := re.loadTemplateWithRequest(site, templateName, r)
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
	return re.loadTemplateWithRequest(site, templateName, nil)
}

// loadTemplateWithRequest loads and parses a template with all its dependencies and request context
func (re *RenderEngine) loadTemplateWithRequest(site *Site, templateName string, r *http.Request) (*template.Template, error) {
	// Create base template with function map including request context
	funcMap := re.createTemplateFuncMapWithRequest(r)
	tmpl := template.New("").Funcs(funcMap)

	// Ensure all template directories exist
	templateDirs := []string{
		site.LayoutPath,
		site.SnippetsPath,
		site.SectionsPath,
		site.BlocksPath,
		site.TemplatesPath,
	}

	for _, dir := range templateDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create template directory %s: %w", dir, err)
		}
	}

	// Load layout templates first
	layouts, err := SecureGlob("layout/*.html")
	if err == nil && len(layouts) > 0 {
		tmpl, err = tmpl.ParseFiles(layouts...)
		if err != nil {
			return nil, fmt.Errorf("error parsing layout templates: %w", err)
		}
	}

	// Load snippets
	snippets, err := SecureGlob("snippets/*.html")
	if err == nil && len(snippets) > 0 {
		tmpl, err = tmpl.ParseFiles(snippets...)
		if err != nil {
			return nil, fmt.Errorf("error parsing snippet templates: %w", err)
		}
	}

	// Load sections
	sections, err := SecureGlob("sections/*.html")
	if err == nil && len(sections) > 0 {
		tmpl, err = tmpl.ParseFiles(sections...)
		if err != nil {
			return nil, fmt.Errorf("error parsing section templates: %w", err)
		}
	}

	// Load blocks
	blocks, err := SecureGlob("blocks/*.html")
	if err == nil && len(blocks) > 0 {
		tmpl, err = tmpl.ParseFiles(blocks...)
		if err != nil {
			return nil, fmt.Errorf("error parsing block templates: %w", err)
		}
	}

	// Load the main template
	templatePath := filepath.Join(site.TemplatesPath, templateName+".html")
	tmpl, err = tmpl.ParseFiles(templatePath)
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
	tmpl, err := re.loadTemplateWithRequest(site, fmt.Sprintf("error-%d", statusCode), r)
	if err != nil {
		// Fallback to generic error template
		tmpl, err = re.loadTemplateWithRequest(site, "error", r)
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
