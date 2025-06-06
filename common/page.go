package common

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// PageMeta represents metadata for a page
type PageMeta struct {
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Keywords      []string          `json:"keywords"`
	Author        string            `json:"author"`
	Template      string            `json:"template"`
	Layout        string            `json:"layout"`
	IsDraft       bool              `json:"is_draft"`
	IsStatic      bool              `json:"is_static"`
	RequireAuth   bool              `json:"require_auth"`
	RequiredRoles []string          `json:"required_roles"`
	URL           string            `json:"url"`
	Fetch         string            `json:"fetch"`
	Protected     string            `json:"protected"`
	CustomData    map[string]string `json:"custom_data"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	PublishedAt   *time.Time        `json:"published_at,omitempty"`
}

// Page represents a single page
type Page struct {
	Slug     string   `json:"slug"`
	Path     string   `json:"path"`
	Meta     PageMeta `json:"meta"`
	Content  string   `json:"content"`
	Sections []string `json:"sections"`
	Site     *Site    `json:"-"`
}

// PageManager manages pages for a site
type PageManager struct {
	site *Site
}

// NewPageManager creates a new page manager for a site
func NewPageManager(site *Site) *PageManager {
	return &PageManager{
		site: site,
	}
}

// GetPage loads a page by slug
func (pm *PageManager) GetPage(slug string) (*Page, error) {
	// Handle index/home page
	if slug == "" || slug == "/" {
		slug = "index"
	}

	// Remove leading slash if present
	slug = strings.TrimPrefix(slug, "/")

	pagePath := filepath.Join(pm.site.PagesPath, slug+".html")

	// Check if page exists
	if _, err := os.Stat(pagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("page not found: %s", slug)
	}

	// Read page file
	content, err := os.ReadFile(pagePath)
	if err != nil {
		return nil, fmt.Errorf("error reading page file: %w", err)
	}

	// Parse page HTML with head metadata
	page, err := pm.parsePageHTML(string(content))
	if err != nil {
		return nil, fmt.Errorf("error parsing page HTML: %w", err)
	}

	page.Slug = slug
	page.Path = "/" + slug
	page.Site = pm.site

	// Handle index page path
	if slug == "index" {
		page.Path = "/"
	}

	return page, nil
}

// parsePageHTML parses an HTML file with head metadata
func (pm *PageManager) parsePageHTML(content string) (*Page, error) {
	page := &Page{
		Meta: PageMeta{
			IsStatic:   true,
			Template:   "default",
			Layout:     "base",
			CustomData: make(map[string]string),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	// Find the head section
	headStart := strings.Index(content, "[----HEAD----]")
	headEnd := strings.Index(content, "[----END----]")

	if headStart != -1 && headEnd != -1 && headEnd > headStart {
		// Extract head content
		headContent := content[headStart+14 : headEnd]

		// Extract body content (everything after [----END----])
		page.Content = content[headEnd+13:]

		// Parse head metadata
		if err := pm.parseHeadMetadata(headContent, &page.Meta); err != nil {
			return nil, fmt.Errorf("error parsing head metadata: %w", err)
		}
	} else {
		// No head section, treat entire content as body
		page.Content = content
	}

	return page, nil
}

// parseHeadMetadata parses metadata from the head section
func (pm *PageManager) parseHeadMetadata(headContent string, meta *PageMeta) error {
	scanner := bufio.NewScanner(strings.NewReader(headContent))
	inDynamicSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Check for dynamic section
		if line == "[dynamic]" {
			inDynamicSection = true
			meta.IsStatic = false
			continue
		}

		// Parse key=value pairs
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), `"`)

			if inDynamicSection {
				// Handle dynamic section parameters
				switch key {
				case "fetch":
					meta.Fetch = value
				case "protected":
					meta.Protected = value
				}
			} else {
				// Handle regular metadata
				switch key {
				case "static":
					if val, err := strconv.ParseBool(value); err == nil {
						meta.IsStatic = val
					}
				case "url":
					meta.URL = value
				case "draft":
					if val, err := strconv.ParseBool(value); err == nil {
						meta.IsDraft = val
					}
				case "title":
					meta.Title = value
				case "description":
					meta.Description = value
				case "author":
					meta.Author = value
				case "template":
					meta.Template = value
				case "layout":
					meta.Layout = value
				case "require_auth":
					if val, err := strconv.ParseBool(value); err == nil {
						meta.RequireAuth = val
					}
				case "keywords":
					// Parse comma-separated keywords
					if value != "" {
						meta.Keywords = strings.Split(value, ",")
						for i, keyword := range meta.Keywords {
							meta.Keywords[i] = strings.TrimSpace(keyword)
						}
					}
				case "required_roles":
					// Parse comma-separated roles
					if value != "" {
						meta.RequiredRoles = strings.Split(value, ",")
						for i, role := range meta.RequiredRoles {
							meta.RequiredRoles[i] = strings.TrimSpace(role)
						}
					}
				default:
					// Store as custom data
					meta.CustomData[key] = value
				}
			}
		}
	}

	return scanner.Err()
}

// GetPagesByTemplate returns all pages using a specific template
func (pm *PageManager) GetPagesByTemplate(templateName string) ([]*Page, error) {
	var pages []*Page

	// Walk through pages directory
	err := filepath.Walk(pm.site.PagesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-HTML files
		if info.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}

		// Get relative path and slug
		relPath, err := filepath.Rel(pm.site.PagesPath, path)
		if err != nil {
			return err
		}

		slug := strings.TrimSuffix(relPath, ".html")

		// Load page
		page, err := pm.GetPage(slug)
		if err != nil {
			return err
		}

		// Check if page uses the specified template
		if page.Meta.Template == templateName {
			pages = append(pages, page)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking pages directory: %w", err)
	}

	return pages, nil
}

// CreatePage creates a new page
func (pm *PageManager) CreatePage(slug string, meta PageMeta, content string, sections []string) error {
	// Ensure pages directory exists
	if err := os.MkdirAll(pm.site.PagesPath, 0755); err != nil {
		return fmt.Errorf("error creating pages directory: %w", err)
	}

	// Set timestamps
	now := time.Now()
	meta.CreatedAt = now
	meta.UpdatedAt = now

	if !meta.IsDraft {
		meta.PublishedAt = &now
	}

	// Create page struct
	page := Page{
		Slug:     slug,
		Path:     "/" + slug,
		Meta:     meta,
		Content:  content,
		Sections: sections,
	}

	// Handle index page
	if slug == "index" {
		page.Path = "/"
	}

	// Convert to HTML format
	htmlContent := pm.generatePageHTML(meta, content)

	// Write to file
	pagePath := filepath.Join(pm.site.PagesPath, slug+".html")
	if err := os.WriteFile(pagePath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("error writing page file: %w", err)
	}

	return nil
}

// UpdatePage updates an existing page
func (pm *PageManager) UpdatePage(slug string, meta PageMeta, content string, sections []string) error {
	// Load existing page to preserve created_at
	existingPage, err := pm.GetPage(slug)
	if err != nil {
		return fmt.Errorf("page not found: %w", err)
	}

	// Update timestamps
	meta.CreatedAt = existingPage.Meta.CreatedAt
	meta.UpdatedAt = time.Now()

	if !meta.IsDraft && existingPage.Meta.IsDraft {
		// Publishing for the first time
		now := time.Now()
		meta.PublishedAt = &now
	} else if !meta.IsDraft {
		// Keep existing publish date
		meta.PublishedAt = existingPage.Meta.PublishedAt
	} else {
		// Draft - no publish date
		meta.PublishedAt = nil
	}

	// Create updated page
	page := Page{
		Slug:     slug,
		Path:     "/" + slug,
		Meta:     meta,
		Content:  content,
		Sections: sections,
	}

	if slug == "index" {
		page.Path = "/"
	}

	// Convert to HTML format
	htmlContent := pm.generatePageHTML(meta, content)

	// Write to file
	pagePath := filepath.Join(pm.site.PagesPath, slug+".html")
	if err := os.WriteFile(pagePath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("error writing page file: %w", err)
	}

	return nil
}

// DeletePage deletes a page
func (pm *PageManager) DeletePage(slug string) error {
	pagePath := filepath.Join(pm.site.PagesPath, slug+".html")

	if err := os.Remove(pagePath); err != nil {
		return fmt.Errorf("error deleting page file: %w", err)
	}

	return nil
}

// ListPages returns all pages for the site
func (pm *PageManager) ListPages(includeUnpublished bool) ([]*Page, error) {
	var pages []*Page

	// Walk through pages directory
	err := filepath.Walk(pm.site.PagesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-HTML files
		if info.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}

		// Get relative path and slug
		relPath, err := filepath.Rel(pm.site.PagesPath, path)
		if err != nil {
			return err
		}

		slug := strings.TrimSuffix(relPath, ".html")

		// Load page
		page, err := pm.GetPage(slug)
		if err != nil {
			return err
		}

		// Filter by published status if needed
		if !includeUnpublished && page.Meta.IsDraft {
			return nil
		}

		pages = append(pages, page)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking pages directory: %w", err)
	}

	return pages, nil
}

// generatePageHTML generates HTML content with head metadata
func (pm *PageManager) generatePageHTML(meta PageMeta, content string) string {
	var head strings.Builder

	head.WriteString("[----HEAD----]\n")

	// Write basic metadata
	if meta.IsStatic {
		head.WriteString("static=true,\n")
	} else {
		head.WriteString("static=false,\n")
	}

	if meta.URL != "" {
		head.WriteString(fmt.Sprintf("url=\"%s\",\n", meta.URL))
	}

	if meta.IsDraft {
		head.WriteString("draft=true,\n")
	} else {
		head.WriteString("draft=false,\n")
	}

	if meta.Title != "" {
		head.WriteString(fmt.Sprintf("title=\"%s\",\n", meta.Title))
	}

	if meta.Description != "" {
		head.WriteString(fmt.Sprintf("description=\"%s\",\n", meta.Description))
	}

	if meta.Author != "" {
		head.WriteString(fmt.Sprintf("author=\"%s\",\n", meta.Author))
	}

	if meta.Template != "" && meta.Template != "default" {
		head.WriteString(fmt.Sprintf("template=\"%s\",\n", meta.Template))
	}

	if meta.Layout != "" && meta.Layout != "base" {
		head.WriteString(fmt.Sprintf("layout=\"%s\",\n", meta.Layout))
	}

	if meta.RequireAuth {
		head.WriteString("require_auth=true,\n")
	}

	if len(meta.Keywords) > 0 {
		head.WriteString(fmt.Sprintf("keywords=\"%s\",\n", strings.Join(meta.Keywords, ",")))
	}

	if len(meta.RequiredRoles) > 0 {
		head.WriteString(fmt.Sprintf("required_roles=\"%s\",\n", strings.Join(meta.RequiredRoles, ",")))
	}

	// Write custom data
	for key, value := range meta.CustomData {
		head.WriteString(fmt.Sprintf("%s=\"%s\",\n", key, value))
	}

	// Write dynamic section if needed
	if !meta.IsStatic && (meta.Fetch != "" || meta.Protected != "") {
		head.WriteString("[dynamic]\n")
		if meta.Fetch != "" {
			head.WriteString(fmt.Sprintf("fetch=\"%s\",\n", meta.Fetch))
		}
		if meta.Protected != "" {
			head.WriteString(fmt.Sprintf("protected=\"%s\",\n", meta.Protected))
		}
	}

	head.WriteString("[----END----]\n")
	head.WriteString(content)

	return head.String()
}
