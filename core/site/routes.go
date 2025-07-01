package site

import (
	"path/filepath"

	"wispy-core/common"
	"wispy-core/core"
	"wispy-core/core/page"
	"wispy-core/tpl"
)

// ScaffoldSiteRoutes sets up routes based on pages found in the site's directory
func ScaffoldSiteRoutes(tenantSite core.Site) {
	router := tenantSite.GetRouter()

	// Get site paths
	sitePath := filepath.Join("_data/tenants", tenantSite.GetDomain())
	layoutsDir := filepath.Join(sitePath, "design/layouts")
	pagesDir := filepath.Join(sitePath, "design/pages")

	// Create template engine for this site
	templateEngine := tpl.NewEngine(layoutsDir, pagesDir)
	siteEngine := NewSiteTplEngine(templateEngine, tenantSite)

	// Scan for pages and create routes
	pages, err := siteEngine.ScanPages()
	if err != nil {
		common.Warning("Failed to scan pages for site %s: %v", tenantSite.GetName(), err)
		// Create default route as fallback
		return
	}

	// Create routes for each page
	for _, pagePath := range pages {
		page.CreatePageRoute(router, tenantSite, siteEngine, pagePath)
	}

	// Setup static file routes for site assets
	page.SetupStaticRoutes(router, tenantSite)

	common.Info("Scaffolded routes for site: %s (%d pages)", tenantSite.GetName(), len(pages))
}

// ScaffoldAllSites sets up routes for all sites
func ScaffoldAllSites(sites map[string]core.Site) {
	for domain, site := range sites {
		common.Info("Scaffolding routes for site: %s", domain)
		ScaffoldSiteRoutes(site)
	}
}
