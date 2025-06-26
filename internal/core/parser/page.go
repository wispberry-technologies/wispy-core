package parser

import (
	"bufio"
	"strings"
	"time"

	"wispy-core/internal/models"
)

// ParsePageHTML parses an HTML file with HTML comment metadata
// This is a pure function that transforms input to output without side effects
func ParsePageHTML(instance *models.SiteInstance, content string) (*models.Page, error) {
	page := &models.Page{
		Title:         "Untitled Page",
		Description:   "",
		Lang:          "en",
		Slug:          "",
		Keywords:      []string{},
		Author:        "",
		LayoutName:    "default",
		IsDraft:       false,
		IsStatic:      true,
		RequireAuth:   false,
		RequiredRoles: []string{},
		FilePath:      "",
		Protected:     "",
		PageData:      make(map[string]string),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		PublishedAt:   nil, // New field in the updated struct
		MetaTags:      []models.HtmlMetaTag{},
		SiteDetails: models.SiteDetails{
			Domain: instance.Domain,
			Name:   instance.Name,
		},
	}

	// Find HTML comment metadata block
	metadataStart := strings.Index(content, "<!--")
	if metadataStart != -1 {
		// Find the end of the metadata comment block
		metadataEnd := strings.Index(content[metadataStart:], "-->")
		if metadataEnd != -1 {
			metadataEnd += metadataStart                       // Get the absolute position
			metadata := content[metadataStart+4 : metadataEnd] // Extract metadata content

			// Parse metadata line by line
			scanner := bufio.NewScanner(strings.NewReader(metadata))
			for scanner.Scan() {
				line := scanner.Text()
				line = strings.TrimSpace(line)

				// Skip empty lines
				if line == "" {
					continue
				}

				// Look for metadata markers (@ symbol)
				if strings.HasPrefix(line, "@") {
					parts := strings.SplitN(line, " ", 2)
					if len(parts) < 2 {
						continue
					}

					key := strings.TrimPrefix(parts[0], "@")
					value := strings.TrimSpace(parts[1])

					parseMetadataField(page, key, value)
				}
			}

			if err := scanner.Err(); err != nil {
				return nil, err
			}

			// Set published time if not draft
			if !page.IsDraft {
				now := time.Now()
				page.PublishedAt = &now
			}

			// Store the raw content in PageData for rendering
			page.PageData["content"] = strings.TrimSpace(content[metadataEnd+3:])
		}
	} else {
		// No metadata block found, use the entire content
		page.PageData["content"] = content
	}

	return page, nil
}

// parseMetadataField parses a metadata field and sets the corresponding field in the Page struct
// This is a pure function that modifies the page parameter based on key/value inputs
func parseMetadataField(page *models.Page, key, value string) {
	switch key {
	case "title":
		page.Title = value
	case "description":
		page.Description = value
	case "lang":
		page.Lang = value
	case "slug":
		page.Slug = value
	case "keywords":
		// Parse comma-separated list
		keywords := strings.Split(value, ",")
		for i, k := range keywords {
			keywords[i] = strings.TrimSpace(k)
		}
		page.Keywords = keywords
	case "author":
		page.Author = value
	case "layout":
		page.LayoutName = value
	case "is_draft":
		page.IsDraft = value == "true" || value == "1" || value == "yes"
	case "is_static":
		page.IsStatic = value == "true" || value == "1" || value == "yes"
	case "require_auth":
		page.RequireAuth = value == "true" || value == "1" || value == "yes"
	case "protected":
		page.Protected = value
	case "required_roles":
		if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
			// Parse JSON array
			value = strings.TrimPrefix(value, "[")
			value = strings.TrimSuffix(value, "]")
			roles := strings.Split(value, ",")
			for i, r := range roles {
				r = strings.TrimSpace(r)
				r = strings.Trim(r, "\"'")
				roles[i] = r
			}
			page.RequiredRoles = roles
		}
	default:
		// Store unknown metadata as custom page data
		page.PageData[key] = value
	}
}

// RemoveMetadataFromContent removes metadata comments from page content
// This is a pure function that transforms input content to output content
func RemoveMetadataFromContent(content string) string {
	// Find HTML comment metadata block
	metadataStart := strings.Index(content, "<!--")
	if metadataStart != -1 {
		// Find the end of the metadata comment block
		metadataEnd := strings.Index(content[metadataStart:], "-->")
		if metadataEnd != -1 {
			metadataEnd += metadataStart + 3 // Include the -->
			// Remove the metadata comment block
			content = content[:metadataStart] + content[metadataEnd:]
		}
	}
	// Remove all other HTML comments from the remaining content
	for strings.Contains(content, "<!--") {
		commentStart := strings.Index(content, "<!--")
		commentEnd := strings.Index(content[commentStart:], "-->")
		if commentEnd != -1 {
			commentEnd += commentStart + 3 // Include the -->
			// Remove the comment
			content = content[:commentStart] + content[commentEnd:]
		} else {
			// No closing comment found, break to avoid infinite loop
			break
		}
	}
	return content
}
