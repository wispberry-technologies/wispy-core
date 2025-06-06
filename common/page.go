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

// GetPage loads a page by slug, searching through subdirectories
func (pm *PageManager) GetPage(slug string) (*Page, error) {
	// Handle index/home page
	if slug == "" || slug == "/" {
		slug = "home"
	}

	// Remove leading slash if present
	slug = strings.TrimPrefix(slug, "/")

	var pagePath string
	var found bool

	// First try to find the page in the main directory
	mainPagePath := filepath.Join(pm.site.PagesPath, "(main)", slug+".html")
	if _, err := os.Stat(mainPagePath); err == nil {
		pagePath = mainPagePath
		found = true
	} else {
		// Search through all subdirectories
		err := filepath.Walk(pm.site.PagesPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Check if this is our target file
			if info.Name() == slug+".html" {
				pagePath = path
				found = true
				return filepath.SkipDir // Stop searching once found
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("error searching for page: %w", err)
		}
	}

	if !found {
		return nil, fmt.Errorf("page not found: %s", slug)
	}

	// Read page file
	content, err := os.ReadFile(pagePath)
	if err != nil {
		return nil, fmt.Errorf("error reading page file: %w", err)
	}

	// Parse page HTML with metadata
	page, err := pm.parsePageHTML(string(content))
	if err != nil {
		return nil, fmt.Errorf("error parsing page HTML: %w", err)
	}

	page.Slug = slug
	page.Path = "/" + slug
	page.Site = pm.site

	// Handle home page path
	if slug == "home" {
		page.Path = "/"
	}

	// If the page has a URL specified in metadata, use that instead
	if page.Meta.URL != "" {
		page.Path = page.Meta.URL
	}

	return page, nil
}

// parsePageHTML parses an HTML file with HTML comment metadata
func (pm *PageManager) parsePageHTML(content string) (*Page, error) {
	page := &Page{
		Site: pm.site,
		Meta: PageMeta{
			Title:      "Untitled Page",
			IsStatic:   true,
			Template:   "default",
			Layout:     "default",
			CustomData: make(map[string]string),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	// Find HTML comment metadata block
	metadataStart := strings.Index(content, "<!--")
	if metadataStart != -1 {
		// Find the end of the metadata comment block
		metadataEnd := strings.Index(content[metadataStart:], "-->")
		if metadataEnd != -1 {
			metadataEnd += metadataStart + 3 // Include the -->

			// Extract metadata content
			metadataContent := content[metadataStart+4 : metadataEnd-3] // Remove <!-- and -->

			// Parse HTML comment metadata
			if err := pm.parseHTMLCommentMetadata(metadataContent, &page.Meta); err != nil {
				return nil, fmt.Errorf("error parsing HTML comment metadata: %w", err)
			}

			// Extract content after metadata (skip any whitespace)
			remainingContent := strings.TrimSpace(content[metadataEnd:])
			page.Content = remainingContent
		} else {
			// Malformed comment, treat entire content as body
			page.Content = content
		}
	} else {
		// No metadata comment, treat entire content as body
		page.Content = content
	}

	return page, nil
}

// parseHTMLCommentMetadata parses metadata from HTML comment format
func (pm *PageManager) parseHTMLCommentMetadata(commentContent string, meta *PageMeta) error {
	scanner := bufio.NewScanner(strings.NewReader(commentContent))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse @key value format
		if strings.HasPrefix(line, "@") {
			// Find the space separator
			spaceIndex := strings.Index(line, " ")
			if spaceIndex == -1 {
				// No value, just a flag
				key := strings.TrimPrefix(line, "@")
				switch key {
				case "is_draft":
					meta.IsDraft = true
				case "require_auth":
					meta.RequireAuth = true
				}
				continue
			}

			key := strings.TrimPrefix(line[:spaceIndex], "@")
			value := strings.TrimSpace(line[spaceIndex+1:])

			// Handle different metadata fields
			switch key {
			case "name":
				// Extract title from filename
				if strings.HasSuffix(value, ".html") {
					meta.Title = strings.TrimSuffix(value, ".html")
				} else {
					meta.Title = value
				}
			case "url":
				meta.URL = value
			case "author":
				meta.Author = value
			case "layout":
				meta.Layout = value
			case "template":
				meta.Template = value
			case "is_draft":
				if val, err := strconv.ParseBool(value); err == nil {
					meta.IsDraft = val
				}
			case "require_auth":
				if val, err := strconv.ParseBool(value); err == nil {
					meta.RequireAuth = val
				}
			case "required_roles":
				// Parse array format like []
				value = strings.Trim(value, "[]")
				if value != "" {
					roles := strings.Split(value, ",")
					for i, role := range roles {
						roles[i] = strings.TrimSpace(strings.Trim(role, `"`))
					}
					meta.RequiredRoles = roles
				}
			default:
				// Store as custom data
				meta.CustomData[key] = value
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
	// Ensure pages directory and (main) subdirectory exist
	mainPagesDir := filepath.Join(pm.site.PagesPath, "(main)")
	if err := os.MkdirAll(mainPagesDir, 0755); err != nil {
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

	// Handle home page
	if slug == "home" {
		page.Path = "/"
	}

	// Convert to HTML format
	htmlContent := pm.generatePageHTML(meta, content)

	// Write to file in (main) subdirectory
	pagePath := filepath.Join(mainPagesDir, slug+".html")
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

	// Find the existing page file to update in place
	var pagePath string
	var found bool

	// First try the main directory
	mainPagePath := filepath.Join(pm.site.PagesPath, "(main)", slug+".html")
	if _, err := os.Stat(mainPagePath); err == nil {
		pagePath = mainPagePath
		found = true
	} else {
		// Search through all subdirectories
		err := filepath.Walk(pm.site.PagesPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Check if this is our target file
			if info.Name() == slug+".html" {
				pagePath = path
				found = true
				return filepath.SkipDir // Stop searching once found
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("error searching for page to update: %w", err)
		}
	}

	if !found {
		return fmt.Errorf("page not found for update: %s", slug)
	}

	// Write to the found file location
	if err := os.WriteFile(pagePath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("error writing page file: %w", err)
	}

	return nil
}

// DeletePage deletes a page
func (pm *PageManager) DeletePage(slug string) error {
	// Find the page file to delete
	var pagePath string
	var found bool

	// First try the main directory
	mainPagePath := filepath.Join(pm.site.PagesPath, "(main)", slug+".html")
	if _, err := os.Stat(mainPagePath); err == nil {
		pagePath = mainPagePath
		found = true
	} else {
		// Search through all subdirectories
		err := filepath.Walk(pm.site.PagesPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Check if this is our target file
			if info.Name() == slug+".html" {
				pagePath = path
				found = true
				return filepath.SkipDir // Stop searching once found
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("error searching for page to delete: %w", err)
		}
	}

	if !found {
		return fmt.Errorf("page not found for deletion: %s", slug)
	}

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

		// Get just the filename without extension as the slug
		slug := strings.TrimSuffix(info.Name(), ".html")

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

// generatePageHTML generates HTML content with HTML comment metadata
func (pm *PageManager) generatePageHTML(meta PageMeta, content string) string {
	var head strings.Builder

	head.WriteString("<!--\n")

	// Write metadata in @key value format
	if meta.Title != "" {
		head.WriteString(fmt.Sprintf("@name %s.html\n", meta.Title))
	}

	if meta.URL != "" {
		head.WriteString(fmt.Sprintf("@url %s\n", meta.URL))
	}

	if meta.Author != "" {
		head.WriteString(fmt.Sprintf("@author %s\n", meta.Author))
	}

	if meta.Layout != "" {
		head.WriteString(fmt.Sprintf("@layout %s\n", meta.Layout))
	}

	if meta.Template != "" && meta.Template != "default" {
		head.WriteString(fmt.Sprintf("@template %s\n", meta.Template))
	}

	head.WriteString(fmt.Sprintf("@is_draft %t\n", meta.IsDraft))
	head.WriteString(fmt.Sprintf("@require_auth %t\n", meta.RequireAuth))

	if len(meta.RequiredRoles) > 0 {
		rolesStr := strings.Join(meta.RequiredRoles, ",")
		head.WriteString(fmt.Sprintf("@required_roles [%s]\n", rolesStr))
	} else {
		head.WriteString("@required_roles []\n")
	}

	// Write custom data
	for key, value := range meta.CustomData {
		head.WriteString(fmt.Sprintf("@%s %s\n", key, value))
	}

	head.WriteString("-->\n\n")
	head.WriteString(content)

	return head.String()
}
