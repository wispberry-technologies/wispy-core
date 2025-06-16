package core

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"wispy-core/common"
	"wispy-core/models"
)

// parsePageHTML parses an HTML file with HTML comment metadata
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
			metadataEnd += metadataStart + 3 // Include the -->

			// Extract metadata content
			metadataContent := content[metadataStart+4 : metadataEnd-3] // Remove <!-- and -->

			// Parse HTML comment metadata
			if err := ParseHTMLCommentMetadata(metadataContent, page); err != nil {
				return nil, fmt.Errorf("error parsing HTML comment metadata: %w", err)
			}

			// Extract content after metadata (skip any whitespace)
			remainingContent := strings.TrimSpace(content[metadataEnd:])

			// remove all other HTML comments from the remaining content and all content in between
			for strings.Contains(remainingContent, "<!--") {
				commentStart := strings.Index(remainingContent, "<!--")
				commentEnd := strings.Index(remainingContent[commentStart:], "-->")
				if commentEnd != -1 {
					commentEnd += commentStart + 3 // Include the -->
					// Remove the comment
					remainingContent = remainingContent[:commentStart] + remainingContent[commentEnd:]
				} else {
					// No closing comment found, break to avoid infinite loop
					break
				}
			}
		} else {
			// Malformed comment, treat entire content as body
			log.Printf("Error: Malformed HTML comment metadata, treating entire content as body.")
			return nil, fmt.Errorf("malformed HTML comment metadata")
		}
	} else {
		// No metadata comment, treat entire content as body
		log.Printf("Warning: HTML comment metadata not found in content!")
	}

	return page, nil
}

// parseHTMLCommentMetadata parses metadata from HTML comment format
func ParseHTMLCommentMetadata(commentContent string, meta *models.Page) error {
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
			case "slug":
				meta.Slug = value
			case "author":
				meta.Author = value
			case "layout":
				meta.LayoutName = value
			case "is_draft":
				if val, err := strconv.ParseBool(value); err == nil {
					meta.IsDraft = val
				}
			case "require_auth":
				if val, err := strconv.ParseBool(value); err == nil {
					meta.RequireAuth = val
				}
			case "required_roles":
				// Parse array format like ["admin", "editor"]
				value = strings.Trim(value, "[]")
				if value != "" {
					roles := strings.Split(value, ",")
					for i, role := range roles {
						// Remove extra quotes and spaces
						role = strings.TrimSpace(role)
						role = strings.Trim(role, `"'`)
						roles[i] = role
					}
					meta.RequiredRoles = roles
				} else {
					meta.RequiredRoles = []string{}
				}
			default:
				// Store as custom data
				meta.PageData[key] = value
			}
		}
	}

	return scanner.Err()
}

// LoadAllSites loads all sites and their pages from the sites directory
func LoadAllSites(sitesPath string) (map[string]*models.SiteInstance, error) {
	sites := make(map[string]*models.SiteInstance)
	sitesPathAbs := common.RootSitesPath() // Use secure path util
	dirs, err := os.ReadDir(sitesPathAbs)
	if err != nil {
		return nil, err
	}

	// Track site init timing
	startTime := time.Now()
	siteCount := 0

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		domain := dir.Name()
		siteStartTime := time.Now()

		// Create a new SiteInstance for each domain
		siteInstance := NewSiteInstance(domain)

		// Initialize site database
		// TODO

		// Load pages
		if err := LoadPagesForSite(siteInstance); err != nil {
			log.Printf("[WARN] Failed to load pages for site %s: %v", domain, err)
			continue
		}

		sites[domain] = siteInstance
		siteCount++

		siteLoadTime := time.Since(siteStartTime)
		log.Printf("Site %s loaded in %v", domain, siteLoadTime)
	}

	totalTime := time.Since(startTime)
	log.Printf("Loaded %d sites in %v", siteCount, totalTime)

	return sites, nil
}

// LoadPagesForSite loads all pages for a given site instance
func LoadPagesForSite(siteInstance *models.SiteInstance) error {
	pagesDir := common.RootSitesPath(siteInstance.Domain, "pages") // Use secure path util
	err := filepath.WalkDir(pagesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip errored files/dirs
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".html" {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		content, err := io.ReadAll(f)
		if err != nil {
			return nil
		}
		page, err := ParsePageHTML(siteInstance, string(content))
		// should be safe since we are using filepath.WalkDir on a path that is guaranteed to be within the site instance's pages directory
		pathParts := strings.SplitN(path, "pages", 2)
		page.FilePath = strings.Trim(pathParts[1], "/\\") // Store the file path in the Page struct
		if err != nil {
			return nil
		}
		if page.Slug == "" {
			return nil
		}
		// Store the page in the SiteInstance's Pages map, keyed by Slug
		siteInstance.Pages[page.Slug] = page
		return nil
	})
	return err
}

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
