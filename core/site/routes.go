package site

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"wispy-core/common"
	"wispy-core/tpl"
)

// ScaffoldAllSites sets up routes for all sites
func ScaffoldAllTenantSites(sites map[string]Site) {
	for domain, site := range sites {
		common.Info("Scaffolding routes for site: %s", domain)
		ScaffoldTenantSiteRoutes(site)
	}
}

// ScaffoldSiteRoutes sets up routes based on pages found in the site's directory
func ScaffoldTenantSiteRoutes(tenantSite Site) {
	router := tenantSite.GetRouter()

	// Get site paths
	sitePath := filepath.Join("_data", "tenants", tenantSite.GetDomain())
	layoutsDir := filepath.Join(sitePath, "layouts")
	pagesDir := filepath.Join(sitePath, "pages")
	supportingTemplatesDirs := []string{
		filepath.Join(sitePath, "design/atoms"),
		filepath.Join(sitePath, "design/components"),
		filepath.Join(sitePath, "design/partials"),
	}

	// Create template engine for this site
	templateEngine := tpl.NewTemplateEngine(layoutsDir, pagesDir)
	_, suppTmplErrs := templateEngine.LoadSupportingTemplates(supportingTemplatesDirs)
	if len(suppTmplErrs) > 0 {
		common.Error("Failed to load supporting templates!")
		for _, err := range suppTmplErrs {
			common.Warning("-->: %v", err)
		}
	}

	// Scan for pages and create routes
	pages, err := ScanPages(pagesDir)
	if err != nil {
		common.Warning("Failed to scan pages for site %s: %v", tenantSite.GetName(), err)
		// Create default route as fallback
		return
	}

	// Create routes for each page
	for _, pagePath := range pages {
		CreatePageRoute(router, tenantSite, templateEngine, pagePath)
	}

	// Setup static file routes for site assets
	SetupStaticRoutes(router, tenantSite)

	common.Info("Scaffolded routes for site: %s (%d pages)", tenantSite.GetName(), len(pages))
}

// ScanPages scans the pages directory and returns a list of available pages
func ScanPages(pagesDir string) ([]string, error) {
	var pages []string

	err := filepath.Walk(pagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".html") {
			// Get relative path from pages directory
			relPath, err := filepath.Rel(pagesDir, path)
			if err != nil {
				return err
			}
			pages = append(pages, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan pages directory: %w", err)
	}

	return pages, nil
}
