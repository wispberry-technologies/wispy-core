package app

import (
	"net/http"
	"wispy-core/tpl"
	"wispy-core/wispytail"
)

func DashboardHandler(cms WispyCms) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		engine := cms.GetTemplateEngine()
		// Template paths are now relative to the template/layout directories defined in the engine
		pagePath := "index.html" // Using index.html as specified in your request
		layoutPath := "default.html"
		data := tpl.TemplateData{
			Title:       "Dashboard",
			Description: "Admin Dashboard",
			Site: tpl.SiteData{
				Name:    "Wispy CMS",
				Domain:  r.Host,
				BaseURL: "https://" + r.Host,
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
			},
		}
		var state tpl.RenderState
		var err error
		if state, err = engine.RenderWithLayout(pagePath, layoutPath, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//
		themeConfig := wispytail.DefaultThemeConfig()
		// Generate theme CSS
		trie := engine.GetWispyTailTrie()
		themeCss := wispytail.DefaultCssTheme
		baseTwCss := wispytail.GenerateThemeLayer(themeConfig)
		css := wispytail.Generate(state.GetBody(), themeConfig, trie)

		state.AddHeadInlineCSS(themeCss + "\n" + baseTwCss + "\n" + css)
		state.SetHeadTitle("Wispy CMS ~ Dashboard")

		// Set content type
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tpl.HtmlBaseRender(w, state)
	}
}
