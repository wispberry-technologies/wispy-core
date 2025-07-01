package forms

import (
	"net/http"
	"wispy-core/common"
	"wispy-core/core/site"

	"github.com/go-chi/chi/v5"
)

type FormApi struct {
	siteManager site.SiteManager
}

func NewFormApi(siteManager site.SiteManager) *FormApi {
	return &FormApi{
		siteManager: siteManager,
	}
}

// type Site interface {
// 	GetID() string
// 	GetName() string
// 	GetDomain() string
// 	GetBaseURL() string
// 	GetTheme() *theme.Root
// 	GetContentDir() string
// 	GetData() map[string]interface{}
// 	SetData(key string, value interface{})
// 	GetCreatedAt() time.Time
// 	GetUpdatedAt() time.Time
// 	SetUpdatedAt(t time.Time)
// 	GetRouter() *chi.Mux
// }

func (f *FormApi) MountApi(r chi.Router) {
	// tenantApiRoutes := router.Group(func(r chi.Router) {
	// })

	// WARNING PUBLIC API ROUTES
	r.Post("/forms/submit", f.FormSubmission)
}

func (f *FormApi) FormSubmission(w http.ResponseWriter, r *http.Request) {
	site, err := f.siteManager.GetSite(r.Host)
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}
	common.Debug("Form submission received for site:", site.GetID())
	common.Debug("Site content Dir:", site.GetContentDir())
	common.Debug("Form submission data:", r.Form)

	w.Write([]byte("Form submission received successfully!"))
}
