package app

import (
	"fmt"
	"net/http"
	"wispy-core/auth"
	"wispy-core/tpl"
)

func DebugHandler(cms WispyCms) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		user, err := auth.UserFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		engine := cms.GetTemplateEngine()

		// Test template loading
		fmt.Fprintf(w, "=== Template Loading Debug ===\n")

		// Try loading different templates
		templates := []string{
			"index.html",
			"settings.html",
			"forms/index.html",
			"forms/submissions/index.html",
		}

		for _, templatePath := range templates {
			fmt.Fprintf(w, "\nTesting template: %s\n", templatePath)
			content, err := engine.LoadTemplate(templatePath)
			if err != nil {
				fmt.Fprintf(w, "  ERROR: %v\n", err)
			} else {
				fmt.Fprintf(w, "  SUCCESS: loaded %d bytes\n", len(content))
				fmt.Fprintf(w, "  First 100 chars: %s...\n", string(content[:min(100, len(content))]))
			}
		}

		// Test rendering
		fmt.Fprintf(w, "\n=== Template Rendering Debug ===\n")

		data := tpl.TemplateData{
			Title:       "Debug",
			Description: "Debug Test",
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
				"pageTitle":   "Debug",
			},
		}

		for _, templatePath := range templates {
			fmt.Fprintf(w, "\nTesting render: %s\n", templatePath)
			state, err := engine.RenderWithLayout(templatePath, "default.html", data)
			if err != nil {
				fmt.Fprintf(w, "  ERROR: %v\n", err)
			} else {
				fmt.Fprintf(w, "  SUCCESS: rendered %d bytes\n", len(state.GetBody()))
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
