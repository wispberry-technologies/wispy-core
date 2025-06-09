package core

import (
	"fmt"
	"net/http"
	"wispy-core/models"
)

// RenderPage renders a page using its template and layout
func RenderPage(tmpl *models.TemplateEngine, w http.ResponseWriter, r *http.Request, instance *models.SiteInstance, page *models.Page) error {
	// Check if page requires authentication
	if page.RequireAuth {
		// TODO: Implement authentication check
		// For now, we'll skip this check
	}

	// Check if page is draft (only show to authenticated admin users)
	if page.IsDraft {
		// TODO: Implement admin check
		// For now, we'll show drafts
	}

	// Create template context
	// context := &models.TemplateContext{
	// 	Site:    instance.Site,
	// 	Page:    page,
	// 	Data:    make(map[string]interface{}),
	// 	Request: r,
	// }

	// Determine layout to use
	layoutName := page.Layout
	if layoutName == "" {
		layoutName = "default"
	}

	// Write response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write([]byte("response")); err != nil {
		return fmt.Errorf("error writing response: %w", err)
	}

	return nil
}
