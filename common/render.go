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
	funcMap  *template.FuncMap
	template *template.Template
}

// NewRenderEngine creates a new render engine
func NewRenderEngine(siteManager *SiteInstanceManager) *RenderEngine {
	var tmpl *template.Template
	var newEngine = &RenderEngine{
		funcMap:  createTemplateFuncMapWithRequest(siteManager),
		template: tmpl,
	}

	// Load templates files first
	sitePath := filepath.Join(rootPath(), "template-sections")
	templateGlob := filepath.Join(sitePath, "*.html")
	templates, err := SecureGlob(templateGlob)
	if err == nil && len(templates) > 0 {
		tmpl, err = tmpl.ParseFiles(templates...)
		if err != nil {
			panic(fmt.Errorf("error parsing template files: %w", err))
		}
	}

	return newEngine
}

// RenderPage renders a page using its template and layout
func (re *RenderEngine) RenderPage(w http.ResponseWriter, r *http.Request, instance *SiteInstance, page *Page) error {
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
		Site:    instance.Site,
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
		layoutName = "layouts/default"
	}

	// Create a fresh template for this request (to avoid "cannot Parse after Execute" error)
	tmpl := template.New("").Funcs(*re.funcMap)

	// Load layout files first
	sitePath := filepath.Join(MustGetEnv("SITES_PATH"), instance.Domain)
	layoutGlob := filepath.Join(sitePath, "layout", "*.html")
	layouts, err := SecureGlob(layoutGlob)
	if err == nil && len(layouts) > 0 {
		tmpl, err = tmpl.ParseFiles(layouts...)
		if err != nil {
			return fmt.Errorf("error parsing layout templates: %w", err)
		}
	}

	// Load snippets
	snippetGlob := filepath.Join(sitePath, "snippets", "*.html")
	snippets, err := SecureGlob(snippetGlob)
	if err == nil && len(snippets) > 0 {
		tmpl, err = tmpl.ParseFiles(snippets...)
		if err != nil {
			return fmt.Errorf("error parsing snippet templates: %w", err)
		}
	}

	// Load sections
	sectionGlob := filepath.Join(sitePath, "sections", "*.html")
	sections, err := SecureGlob(sectionGlob)
	if err == nil && len(sections) > 0 {
		tmpl, err = tmpl.ParseFiles(sections...)
		if err != nil {
			return fmt.Errorf("error parsing section templates: %w", err)
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
func (re *RenderEngine) RenderTemplate(w http.ResponseWriter, r *http.Request, instance *SiteInstance, templateName string, data map[string]interface{}) error {
	// Load template
	tmpl, err := re.loadTemplateWithRequest(instance, templateName, r)
	if err != nil {
		return fmt.Errorf("error loading template: %w", err)
	}

	// Create template context
	context := &TemplateContext{
		Site:    instance.Site,
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

// loadTemplateWithRequest loads and parses a template with all its dependencies and request context
func (re *RenderEngine) loadTemplateWithRequest(instance *SiteInstance, templateName string, r *http.Request) (*template.Template, error) {
	var site = instance.Site
	// Ensure all template directories exist
	templateDirs := []string{
		filepath.Join(MustGetEnv("SITES_PATH"), site.Domain, "layout"),
		filepath.Join(MustGetEnv("SITES_PATH"), site.Domain, "snippets"),
		filepath.Join(MustGetEnv("SITES_PATH"), site.Domain, "sections"),
		filepath.Join(MustGetEnv("SITES_PATH"), site.Domain, "blocks"),
		filepath.Join(MustGetEnv("SITES_PATH"), site.Domain, "templates"),
	}

	// copy existing templates from re.Template
	tmpl, err := re.template.Clone()
	if err != nil {
		panic(fmt.Errorf("failed to clone template: %w", err))
	}
	tmpl.Funcs(*re.funcMap)

	for _, dir := range templateDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create template directory %s: %w", dir, err)
		}
	}

	for _, dir := range templateDirs {
		// Load all HTML files in the directory
		globPattern := filepath.Join(dir, "*.html")
		files, err := SecureGlob(globPattern)
		if err != nil {
			return nil, fmt.Errorf("error loading templates from %s: %w", dir, err)
		}
		if len(files) == 0 {
			continue // No templates in this directory
		}
		// Parse all templates in the directory
		tmpl, err = tmpl.ParseFiles(files...)
		if err != nil {
			return nil, fmt.Errorf("error parsing templates from %s: %w", dir, err)
		}
	}

	instance.Templates, err = tmpl.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone templates: %w", err)
	}

	return tmpl, nil
}

// RenderError renders an error page
func (re *RenderEngine) RenderError(w http.ResponseWriter, r *http.Request, instance *SiteInstance, statusCode int, message string) {
	// Set status code
	w.WriteHeader(statusCode)

	// Try to load error template
	tmpl, err := re.loadTemplateWithRequest(instance, fmt.Sprintf("error-%d", statusCode), r)
	if err != nil {
		// Fallback to generic error template
		tmpl, err = re.loadTemplateWithRequest(instance, "error", r)
		if err != nil {
			// Last resort: plain text response
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			fmt.Fprintf(w, "Error %d: %s", statusCode, message)
			return
		}
	}

	// Create error context
	context := &TemplateContext{
		Site:    instance.Site,
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
