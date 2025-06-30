package site

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"

	"wispy-core/common"
	"wispy-core/core/render"
	"wispy-core/tpl"
	"wispy-core/wispytail"
)

// ScaffoldSiteRoutes sets up routes based on pages found in the site's directory
func ScaffoldSiteRoutes(s Site) {
	router := s.GetRouter()

	// Get site paths
	sitePath := filepath.Join("_data/tenants", s.GetDomain())
	layoutsDir := filepath.Join(sitePath, "design/layouts")
	pagesDir := filepath.Join(sitePath, "design/pages")

	// Create template engine for this site
	templateEngine := tpl.NewEngine(layoutsDir, pagesDir)

	// Scan for pages and create routes
	pages, err := templateEngine.ScanPages()
	if err != nil {
		common.Warning("Failed to scan pages for site %s: %v", s.GetName(), err)
		// Create default route as fallback
		createDefaultRoute(router, s)
		return
	}

	// Create routes for each page
	for _, pagePath := range pages {
		createPageRoute(router, s, templateEngine, pagePath)
	}

	// If no pages found, create default route
	if len(pages) == 0 {
		createDefaultRoute(router, s)
	}

	// Setup static file routes for site assets
	setupStaticRoutes(router, s)

	common.Info("Scaffolded routes for site: %s (%d pages)", s.GetName(), len(pages))
}

// ScaffoldAllSites sets up routes for all sites
func ScaffoldAllSites(sites map[string]Site) {
	for domain, site := range sites {
		common.Info("Scaffolding routes for site: %s", domain)
		ScaffoldSiteRoutes(site)
	}
}

// createPageRoute creates a route for a specific page
func createPageRoute(router *chi.Mux, s Site, templateEngine *tpl.Engine, pagePath string) {
	route := tpl.PathToRoute(pagePath)
	fullPagePath := filepath.Join("_data/tenants", s.GetDomain(), "design/pages", pagePath)

	router.Get(route, func(w http.ResponseWriter, r *http.Request) {
		// Prepare template data
		templateData := tpl.Data{
			Title:       getPageTitle(pagePath, s.GetName()),
			Description: getPageDescription(pagePath),
			Site: tpl.SiteData{
				Name:    s.GetName(),
				Domain:  s.GetDomain(),
				BaseURL: s.GetBaseURL(),
			},
			Data: make(map[string]interface{}),
		}

		// Render template with CSS generation
		result, err := templateEngine.RenderTemplateWithCSS(fullPagePath, templateData)
		if err != nil {
			common.Error("Failed to render template %s: %v", pagePath, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Generate theme CSS
		themeConfig := wispytail.DefaultThemeConfig()
		themeCSS := wispytail.GenerateThemeLayer(themeConfig)
		baseCSS := wispytail.GenerateCssBaseLayer()

		// Combine all CSS
		fullCSS := themeCSS + "\n" + baseCSS + "\n" + result.CSS

		// Create render state for HTML response
		renderState := render.NewRenderState()
		renderState.SetHeadTitle(templateData.Title)
		renderState.SetBody(result.HTML)
		renderState.AddStyles(render.StyleAsset{Src: "/assets/css/theme.css"})
		renderState.AddScripts(render.ScriptAsset{Src: "/assets/js/main.js"})
		renderState.SetHeadInlineCSS(fullCSS)

		// Set content type
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Render final HTML response
		if err := render.HtmlBaseRenderResponse(w, renderState); err != nil {
			common.Error("Failed to render HTML response: %v", err)
		}
	})
}

// createDefaultRoute creates a default route when no pages are found
func createDefaultRoute(router *chi.Mux, s Site) {
	// router.Get("/", func(w http.ResponseWriter, r *http.Request) {
	// 	// Generate basic CSS for default content
	// 	defaultHTML := "<h1>Welcome to " + s.GetName() + "</h1><p>This site is under construction.</p>"
	// 	classes := wispytail.ExtractClasses(defaultHTML)
	// 	css := wispytail.ResolveClasses(classes, wispytail.NewTrie())

	// 	// Generate theme CSS
	// 	themeConfig := wispytail.DefaultThemeConfig()
	// 	themeCSS := wispytail.GenerateThemeLayer(themeConfig)
	// 	baseCSS := wispytail.GenerateCssBaseLayer()

	// 	fullCSS := themeCSS + "\n" + baseCSS + "\n" + css

	// 	renderState := render.NewRenderState()
	// 	renderState.SetHeadTitle(s.GetName())
	// 	renderState.SetBody(defaultHTML)
	// 	renderState.AddStyles(render.StyleAsset{Src: "/assets/css/theme.css"})
	// 	renderState.AddScripts(render.ScriptAsset{Src: "/assets/js/main.js"})
	// 	renderState.SetHeadInlineCSS(fullCSS)

	// 	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// 	if err := render.HtmlBaseRenderResponse(w, renderState); err != nil {
	// 		common.Error("Failed to render HTML response: %v", err)
	// 	}
	// })
}

// setupStaticRoutes configures static file serving
func setupStaticRoutes(router *chi.Mux, s Site) {
	// Assets route
	router.Get("/assets/*", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/assets/"):]
		assetPath := filepath.Join("_data/tenants", s.GetDomain(), "design/assets", path)
		http.ServeFile(w, r, assetPath)
	})

	// Public files route
	router.Get("/public/*", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/public/"):]
		publicPath := filepath.Join(s.GetContentDir(), "public", path)
		http.ServeFile(w, r, publicPath)
	})
}

// Helper functions
func getPageTitle(pagePath, siteName string) string {
	// Extract title from page path
	title := strings.TrimSuffix(filepath.Base(pagePath), ".html")
	title = strings.ReplaceAll(title, "_", " ")
	title = strings.ReplaceAll(title, "-", " ")
	title = strings.Title(title)

	if title == "Index" {
		return siteName
	}

	return title + " | " + siteName
}

func getPageDescription(pagePath string) string {
	// Default description - could be enhanced to read from front matter
	return "Page content for " + strings.TrimSuffix(filepath.Base(pagePath), ".html")
}
