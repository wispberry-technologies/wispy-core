package common

import (
	"bufio"
	"fmt"
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
	Site     Site     `json:"-"`
}

// parsePageHTML parses an HTML file with HTML comment metadata
func ParsePageHTML(instance *SiteInstance, content string) (*Page, error) {
	page := &Page{
		Site: *instance.Site,
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
			if err := ParseHTMLCommentMetadata(metadataContent, &page.Meta); err != nil {
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
func ParseHTMLCommentMetadata(commentContent string, meta *PageMeta) error {
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
