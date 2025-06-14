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
		Site:       *instance.Site,
		Title:      "Untitled Page",
		IsStatic:   true,
		Layout:     "layout/default",
		CustomData: make(map[string]string),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
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
			page.Content = remainingContent
		} else {
			// Malformed comment, treat entire content as body
			page.Content = "Malformed HTML comment metadata!"
			log.Printf("Error: Malformed HTML comment metadata, treating entire content as body.")
			return nil, fmt.Errorf("malformed HTML comment metadata")
		}
	} else {
		// No metadata comment, treat entire content as body
		log.Printf("Warning: HTML comment metadata not found in content!")
		page.Content = content
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
			case "url":
				meta.URL = value
			case "author":
				meta.Author = value
			case "layout":
				meta.Layout = value
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
				meta.CustomData[key] = value
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
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		domain := dir.Name()
		// Create a new SiteInstance for each domain
		siteInstance := &models.SiteInstance{
			Domain: domain,
			Site:   &models.Site{Domain: domain, Name: domain, IsActive: true},
			Pages:  make(map[string]*models.Page),
		}
		if err := LoadPagesForSite(siteInstance, sitesPathAbs); err != nil {
			log.Printf("[WARN] Failed to load pages for site %s: %v", domain, err)
			continue
		}
		sites[domain] = siteInstance
	}
	return sites, nil
}

// LoadPagesForSite loads all pages for a given site instance
func LoadPagesForSite(siteInstance *models.SiteInstance, sitesPathAbs string) error {
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
		if err != nil {
			return nil
		}
		if page.URL == "" {
			return nil
		}
		// Store the page in the SiteInstance's Pages map, keyed by URL
		siteInstance.Pages[page.URL] = page
		return nil
	})
	return err
}
