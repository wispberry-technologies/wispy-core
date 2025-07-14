package app

import (
	"net/http"
	"wispy-core/auth"
	"wispy-core/common"
	"wispy-core/tpl"
)

func SettingsHandler(cms WispyCms) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		user, err := auth.UserFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Handle POST request for settings update
		if r.Method == http.MethodPost {
			handleSettingsPost(w, r, cms, user)
			return
		}

		// Handle GET request - show settings page
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		engine := cms.GetTemplateEngine()
		pagePath := "settings.html"
		layoutPath := "default.html"

		// Try to load the template first to check if it exists
		_, err = engine.LoadTemplate(pagePath)
		if err != nil {
			// Template doesn't exist, fall back to index.html
			pagePath = "index.html"
		}

		data := tpl.TemplateData{
			Title:       "Settings",
			Description: "CMS Settings",
			Site: tpl.SiteData{
				Name:    "Wispy CMS",
				Domain:  r.Host,
				BaseURL: "https://" + r.Host,
			},
			Content: "",
			Data: map[string]interface{}{
				"__styles":    []string{},
				"__scripts":   []string{},
				"__inlineCSS": "",
				"user":        user,
				"pageTitle":   "Settings",
			},
		}

		state, err := renderCMSTemplate(engine, pagePath, layoutPath, data, cms.GetTheme())
		if err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		state.SetHeadTitle("Wispy CMS ~ Settings")
		tpl.HtmlBaseRender(w, state)
	}
}

func handleSettingsPost(w http.ResponseWriter, r *http.Request, cms WispyCms, user *auth.User) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Extract form values
	siteName := r.FormValue("site-name")
	theme := r.FormValue("theme")

	// Validate form data
	if siteName == "" {
		http.Error(w, "Site name is required", http.StatusBadRequest)
		return
	}

	validThemes := []string{"robot-green", "midnight-wisp", "pale-wisp"}
	themeValid := false
	for _, validTheme := range validThemes {
		if theme == validTheme {
			themeValid = true
			break
		}
	}

	if !themeValid {
		http.Error(w, "Invalid theme selected", http.StatusBadRequest)
		return
	}

	// TODO: Save settings to database/config file
	// For now, just redirect back to settings with success message
	common.Info("Settings updated: Site Name: %s, Theme: %s", siteName, theme)

	// Redirect to settings page with success message
	http.Redirect(w, r, "/wispy-cms/settings?updated=true", http.StatusFound)
}
