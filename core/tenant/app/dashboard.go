package app

import (
	"net/http"
	"os"
	"path/filepath"
	"wispy-core/auth"
	"wispy-core/common"
	"wispy-core/core/tenant/app/providers"
	"wispy-core/tpl"
	"wispy-core/wispytail"
)

func DashboardHandler(cms WispyCms) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		user, err := auth.UserFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		engine := cms.GetTemplateEngine()
		// Template paths are now relative to the template/layout directories defined in the engine
		pagePath := "index.html" // Using index.html as specified in your request
		layoutPath := "default.html"

		// Create provider manager for this domain
		domain := common.NormalizeHost(r.Host)
		siteInstance, err := cms.GetSiteManager().GetSite(domain)
		if err != nil {
			http.Error(w, "Could not get site for domain "+domain, http.StatusNotFound)
			return
		}

		providerManager := providers.NewProviderManager(siteInstance)
		defer providerManager.Close()

		data := tpl.TemplateData{
			Title:       "Dashboard",
			Description: "Admin Dashboard",
			Site: tpl.SiteData{
				Name:    "Wispy CMS",
				Domain:  domain,
				BaseURL: "https://" + domain,
			},
			Content: "",
			Data: map[string]interface{}{
				// Example of adding styles and scripts
				"__styles": []string{
					// "/static/css/admin.css",
					// "/static/css/dashboard.css",
				},
				"__scripts": []string{
					// "/static/js/admin.js",
				},
				"__inlineCSS": "",
				"user":        user,
				"pageTitle":   "Dashboard",
			},
		}

		// Add provider functions to template context
		templateContext := providerManager.CreateTemplateContext(r.Context())
		for key, value := range templateContext {
			data.Data[key] = value
		}
		var state tpl.RenderState
		state, err = engine.RenderWithLayout(pagePath, layoutPath, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//
		themeConfig := wispytail.DefaultThemeConfig()
		// Generate theme CSS
		trie := engine.GetWispyTailTrie()
		// _data/design/systems/themes
		var themeCss = wispytail.DefaultCssTheme
		themePath := filepath.Join(filepath.Join("_data", "design", "systems", "themes", cms.GetTheme()+".css"))
		themeBytes, err := os.ReadFile(themePath)
		if err != nil {
			common.Error("Failed to read theme file %s: %v", themePath, err)
		} else {
			themeCss += string(themeBytes)
		}

		baseTwCss := wispytail.GenerateThemeLayer(themeConfig)
		css := wispytail.Generate(state.GetBody(), themeConfig, trie)

		state.AddHeadInlineCSS(themeCss + "\n" + baseTwCss + "\n" + css)
		state.SetHeadTitle("Wispy CMS ~ Dashboard")

		// Set content type
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tpl.HtmlBaseRender(w, state)
	}
}
