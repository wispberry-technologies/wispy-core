package page

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"wispy-core/core/site"

	"github.com/pelletier/go-toml/v2"
)

// LoadPage loads a page from the content directory
func LoadPage(site site.Site, path string, tenantsRoot string) (*Page, error) {
	contentPath := filepath.Join(tenantsRoot, site.GetDomain(), site.GetContentDir(), path)
	if !strings.HasSuffix(contentPath, ".md") {
		contentPath += ".md"
	}

	data, err := os.ReadFile(contentPath)
	if err != nil {
		return nil, err
	}

	// Split front matter and content
	parts := strings.SplitN(string(data), "+++", 3) // TOML uses +++ for front matter
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid page format in %s", contentPath)
	}

	var page Page
	if err := toml.Unmarshal([]byte(parts[1]), &page); err != nil {
		return nil, err
	}

	page.Content = strings.TrimSpace(parts[2])

	// Set defaults
	if page.Template == "" {
		page.Template = "default"
	}

	return &page, nil
}

// LoadPagesForSite loads all pages for a given site
func LoadPagesForSite(site site.Site, tenantsRoot string) ([]*Page, error) {
	var pages []*Page
	contentRoot := filepath.Join(tenantsRoot, site.GetDomain(), site.GetContentDir())

	err := filepath.Walk(contentRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		relPath, _ := filepath.Rel(contentRoot, path)
		page, err := LoadPage(site, relPath, tenantsRoot)
		if err != nil {
			return err
		}

		pages = append(pages, page)
		return nil
	})

	return pages, err
}
