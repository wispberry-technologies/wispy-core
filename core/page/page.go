package page

import (
	"net/http"
	"path/filepath"
	"strings"
	"wispy-core/common"
	"wispy-core/core"
	"wispy-core/core/render"
	"wispy-core/tpl"
	"wispy-core/wispytail"

	"github.com/go-chi/chi/v5"
)

// createPageRoute creates a route for a specific page
func CreatePageRoute(router *chi.Mux, s core.Site, templateEngine core.SiteTplEngine, pagePath string) {
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

		// Generate theme CSS
		themeConfig := wispytail.DefaultThemeConfig()
		// Render template with CSS generation
		result, err := render.RenderTemplateWithCSS(templateEngine, fullPagePath, themeConfig, templateData)
		if err != nil {
			common.Error("Failed to render template %s: %v", pagePath, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Create render state for HTML response
		renderState := render.NewRenderState()
		renderState.SetHeadTitle(templateData.Title)
		renderState.SetBody(result.HTML)
		renderState.SetHeadInlineCSS(result.CSS)

		// Set content type
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Render final HTML response
		if err := render.HtmlBaseRenderResponse(w, renderState); err != nil {
			common.Error("Failed to render HTML response: %v", err)
		}
	})
}

// // createDefaultRoute creates a default route when no pages are found
// func CreateDefaultRoute(router *chi.Mux, s Site, te tpl.Engine) {
// 	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
// 		// Generate basic CSS for default content
// 		defaultHTML := "<h1>Welcome to " + s.GetName() + "</h1><p>This site is under construction.</p>"
// 		classes := wispytail.ExtractClasses(defaultHTML)
// 		themeConfig := wispytail.DefaultThemeConfig()
// 		css := wispytail.GenerateFullCSS(classes, themeConfig, te.GetTrie())

// 		// Generate theme CSS
// 		themeCSS := wispytail.GenerateThemeLayer(themeConfig)
// 		baseCSS := wispytail.GenerateCssBaseLayer()

// 		fullCSS := themeCSS + "\n" + baseCSS + "\n" + css

// 		renderState := render.NewRenderState()
// 		renderState.SetHeadTitle(s.GetName())
// 		renderState.SetBody(defaultHTML)
// 		// renderState.AddStyles(render.StyleAsset{Src: "/assets/css/theme.css"})
// 		// renderState.AddScripts(render.ScriptAsset{Src: "/assets/js/main.js"})
// 		renderState.SetHeadInlineCSS(fullCSS)

// 		w.Header().Set("Content-Type", "text/html; charset=utf-8")
// 		if err := render.HtmlBaseRenderResponse(w, renderState); err != nil {
// 			common.Error("Failed to render HTML response: %v", err)
// 		}
// 	})
// }

// setupStaticRoutes configures static file serving
func SetupStaticRoutes(router *chi.Mux, s core.Site) {
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
