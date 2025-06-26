package api_v1

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"wispy-core/internal/cache"
	"wispy-core/internal/core/pages"
	"wispy-core/internal/models"
	"wispy-core/pkg/common"

	"github.com/go-chi/chi/v5"
)

// Context keys for user and site
const (
	CtxUserKey         = "user"
	CtxSiteInstanceKey = "siteInstance"
)

func PagesRouter(r chi.Router) {
	r.Route("/pages", func(r chi.Router) {
		r.Post("/", CreatePage)
		r.Get("/", ListPages)
		r.Get("/{id}", GetPage)
		r.Put("/{id}", UpdatePage)
		r.Delete("/{id}", DeletePage)
	})
}

// CreatePage handles POST /api/v1/pages
func CreatePage(w http.ResponseWriter, r *http.Request) {
	var page models.Page
	if err := json.NewDecoder(r.Body).Decode(&page); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	db := getPagesDB(r)
	if db == nil {
		http.Error(w, "Could not get database connection", http.StatusInternalServerError)
		return
	}
	id, err := pages.InsertPage(db, &page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pageMap := map[string]interface{}{"id": id, "slug": page.Slug}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pageMap)
}

// GetPage handles GET /api/v1/pages/{id}
func GetPage(w http.ResponseWriter, r *http.Request) {
	idOrSlug := chi.URLParam(r, "id")
	db := getPagesDB(r)
	if db == nil {
		http.Error(w, "Could not get database connection", http.StatusInternalServerError)
		return
	}
	var page *models.Page
	var err error
	// Try as int (id), fallback to slug
	if _, convErr := strconv.ParseInt(idOrSlug, 10, 64); convErr == nil {
		// Not implemented: get by id
		http.Error(w, "Get by ID not implemented, use slug", http.StatusNotImplemented)
		return
	} else {
		page, err = pages.GetPageBySlug(db, idOrSlug)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(page)
}

// UpdatePage handles PUT /api/v1/pages/{id}
func UpdatePage(w http.ResponseWriter, r *http.Request) {
	idOrSlug := chi.URLParam(r, "id")
	db := getPagesDB(r)
	if db == nil {
		http.Error(w, "Could not get database connection", http.StatusInternalServerError)
		return
	}
	var page models.Page
	if err := json.NewDecoder(r.Body).Decode(&page); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	page.Slug = idOrSlug // enforce slug from URL
	if err := pages.UpdatePage(db, &page); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeletePage handles DELETE /api/v1/pages/{id}
func DeletePage(w http.ResponseWriter, r *http.Request) {
	idOrSlug := chi.URLParam(r, "id")
	db := getPagesDB(r)
	if db == nil {
		http.Error(w, "Could not get database connection", http.StatusInternalServerError)
		return
	}
	if err := pages.DeletePage(db, idOrSlug); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListPages handles GET /api/v1/pages
func ListPages(w http.ResponseWriter, r *http.Request) {
	db := getPagesDB(r)
	if db == nil {
		http.Error(w, "Could not get database connection", http.StatusInternalServerError)
		return
	}
	// Optional: support pagination via query params
	limit := 100
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	pagesList, err := pages.ListPages(db, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pagesList)
}

// Helper to get DB connection for a request using SiteInstance from context
func getPagesDB(r *http.Request) *sql.DB {
	ctx := r.Context()
	var domain string
	var db *sql.DB
	if siteVal := ctx.Value(CtxSiteInstanceKey); siteVal != nil {
		if siteInstance, ok := siteVal.(*models.SiteInstance); ok {
			domain = siteInstance.Domain
			db, _ = cache.GetConnection(siteInstance.DBCache, domain, "pages")
		}
	}
	if db == nil && domain == "" {
		domain = common.NormalizeHost(r.Host)
		db, _ = cache.GetConnection(nil, domain, "pages")
	}
	return db
}
