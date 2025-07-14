package site

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"wispy-core/common"
	"wispy-core/config"
	"wispy-core/tpl"
	"wispy-core/wispytail"

	"github.com/go-chi/chi/v5"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// LoadPage loads a page from the content directory
func LoadPage(site Site, path string, tenantsRoot string) (*Page, error) {
	contentPath := filepath.Join(tenantsRoot, site.GetDomain(), "pages", path)
	data, err := os.ReadFile(contentPath)
	if err != nil {
		return nil, err
	}

	// // Split front matter and content
	// parts := strings.SplitN(string(data), "+++", 3) // TOML uses +++ for front matter
	// if len(parts) < 3 {
	// 	return nil, fmt.Errorf("invalid page format in %s", contentPath)
	// }

	// if err := toml.Unmarshal([]byte(parts[1]), &page); err != nil {
	// 	return nil, err
	// }
	page := Page{
		Title:       getPageTitle(path, site.GetName()),
		Slug:        strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		Path:        path,
		Layout:      "default",
		Theme:       "default",
		Content:     string(data),
		FrontMatter: make(map[string]interface{}),
	}

	return &page, nil
}

// createPageRoute creates a route for a specific page
func CreatePageRoute(router chi.Router, s Site, templateEngine tpl.TemplateEngine, pagePath string) {
	route := common.PathToRoute(pagePath)
	router.Get(route, func(w http.ResponseWriter, r *http.Request) {
		// Prepare template data
		templateData := tpl.TemplateData{
			Title:       getPageTitle(pagePath, s.GetName()),
			Description: getPageDescription(pagePath),
			Site: tpl.SiteData{
				Name:    s.GetName(),
				Domain:  s.GetDomain(),
				BaseURL: s.GetBaseURL(),
			},
			Data: make(map[string]interface{}),
		}

		state, err := templateEngine.RenderWithLayout(pagePath, "default.html", templateData)
		if err != nil {
			common.Error("Failed to render page %s: %v", pagePath, err)
			common.RespondWithError(w, r, http.StatusInternalServerError, "Failed to render page", err)
			return
		}

		// TODO: proper page context handling
		themeCss, err := s.GetTheme("default")
		if err != nil {
			common.Error("Failed to get theme for site %s: %v", s.GetName(), err)
			themeCss = wispytail.DefaultCssTheme
		}
		themeConfig := wispytail.DefaultThemeConfig()
		trie := templateEngine.GetWispyTailTrie()
		baseTwCss := wispytail.GenerateThemeLayer(themeConfig)
		css := wispytail.Generate(state.GetBody(), themeConfig, trie)

		state.AddHeadInlineCSS(baseTwCss + "\n" + themeCss + "\n" + css)
		state.SetHeadTitle(templateData.Title)

		// Set content type
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tpl.HtmlBaseRender(w, state)
	})
}

// setupStaticRoutes configures static file serving
func SetupStaticRoutes(router chi.Router, s Site) {
	gConfig := config.GetGlobalConfig()
	// Assets route
	router.Get("/assets/*", func(w http.ResponseWriter, r *http.Request) {
		// Serve assets from the tenant's assets directory
		gConfig.GetCoreAuthMiddleware().RequireAuthFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path[len("/assets/"):]
			assetPath := filepath.Join("_data/tenants", s.GetDomain(), "assets", path)
			http.ServeFile(w, r, assetPath)
		}))
	})

	// Public files route
	router.Get("/public/*", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/public/"):]
		publicPath := filepath.Join(gConfig.GetSitesPath(), s.GetDomain(), "public", path)
		http.ServeFile(w, r, publicPath)
	})
}

// Helper functions
func getPageTitle(pagePath, siteName string) string {
	// Extract title from page path
	title := strings.TrimSuffix(filepath.Base(pagePath), ".html")
	title = strings.ReplaceAll(title, "_", " ")
	title = strings.ReplaceAll(title, "-", " ")
	title = cases.Title(language.English).String(title)

	if title == "Index" {
		return cases.Title(language.English).String(siteName)
	}

	return title + " | " + siteName
}

func getPageDescription(pagePath string) string {
	// Default description - could be enhanced to read from front matter
	return "Page content for " + strings.TrimSuffix(filepath.Base(pagePath), ".html")
}
