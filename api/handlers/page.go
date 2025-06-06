package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/wispberry-technologies/wispy-core/common"
)

// PageHandler handles page-related HTTP requests
type PageHandler struct {
	siteManager  *common.SiteManager
	renderEngine *common.RenderEngine
}

// NewPageHandler creates a new page handler
func NewPageHandler(siteManager *common.SiteManager, renderEngine *common.RenderEngine) *PageHandler {
	return &PageHandler{
		siteManager:  siteManager,
		renderEngine: renderEngine,
	}
}

// PageResponse represents a page in API responses
type PageResponse struct {
	Slug        string            `json:"slug"`
	Path        string            `json:"path"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Keywords    []string          `json:"keywords"`
	Author      string            `json:"author"`
	Template    string            `json:"template"`
	Layout      string            `json:"layout"`
	IsDraft     bool              `json:"is_draft"`
	IsStatic    bool              `json:"is_static"`
	RequireAuth bool              `json:"require_auth"`
	URL         string            `json:"url"`
	Fetch       string            `json:"fetch,omitempty"`
	Protected   string            `json:"protected,omitempty"`
	CustomData  map[string]string `json:"custom_data,omitempty"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	PublishedAt *string           `json:"published_at,omitempty"`
	Content     string            `json:"content"`
	Sections    []string          `json:"sections"`
}

// PageCreateRequest represents a request to create a page
type PageCreateRequest struct {
	Slug        string            `json:"slug"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Keywords    []string          `json:"keywords,omitempty"`
	Author      string            `json:"author,omitempty"`
	Template    string            `json:"template,omitempty"`
	Layout      string            `json:"layout,omitempty"`
	IsDraft     bool              `json:"is_draft"`
	IsStatic    bool              `json:"is_static"`
	RequireAuth bool              `json:"require_auth"`
	URL         string            `json:"url,omitempty"`
	Fetch       string            `json:"fetch,omitempty"`
	Protected   string            `json:"protected,omitempty"`
	CustomData  map[string]string `json:"custom_data,omitempty"`
	Content     string            `json:"content"`
	Sections    []string          `json:"sections,omitempty"`
}

// convertPageToResponse converts a page to API response format
func (ph *PageHandler) convertPageToResponse(page *common.Page) PageResponse {
	response := PageResponse{
		Slug:        page.Slug,
		Path:        page.Path,
		Title:       page.Meta.Title,
		Description: page.Meta.Description,
		Keywords:    page.Meta.Keywords,
		Author:      page.Meta.Author,
		Template:    page.Meta.Template,
		Layout:      page.Meta.Layout,
		IsDraft:     page.Meta.IsDraft,
		IsStatic:    page.Meta.IsStatic,
		RequireAuth: page.Meta.RequireAuth,
		URL:         page.Meta.URL,
		Fetch:       page.Meta.Fetch,
		Protected:   page.Meta.Protected,
		CustomData:  page.Meta.CustomData,
		CreatedAt:   page.Meta.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   page.Meta.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Content:     page.Content,
		Sections:    page.Sections,
	}

	if page.Meta.PublishedAt != nil {
		publishedAt := page.Meta.PublishedAt.Format("2006-01-02T15:04:05Z07:00")
		response.PublishedAt = &publishedAt
	}

	return response
}

// ListPages handles GET /admin/api/v1/pages?site=<domain>
func (ph *PageHandler) ListPages(w http.ResponseWriter, r *http.Request) {
	siteDomain := r.URL.Query().Get("site")
	if siteDomain == "" {
		http.Error(w, "site parameter is required", http.StatusBadRequest)
		return
	}

	includeUnpublishedStr := r.URL.Query().Get("include_unpublished")
	includeUnpublished, _ := strconv.ParseBool(includeUnpublishedStr)

	// Load site
	site, err := ph.siteManager.LoadSite(siteDomain)
	if err != nil {
		http.Error(w, fmt.Sprintf("Site not found: %v", err), http.StatusNotFound)
		return
	}

	// Get pages
	pageManager := common.NewPageManager(site)
	pages, err := pageManager.ListPages(includeUnpublished)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error listing pages: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var response []PageResponse
	for _, page := range pages {
		response = append(response, ph.convertPageToResponse(page))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPage handles GET /admin/api/v1/pages/{slug}?site=<domain>
func (ph *PageHandler) GetPage(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	siteDomain := r.URL.Query().Get("site")

	if siteDomain == "" {
		http.Error(w, "site parameter is required", http.StatusBadRequest)
		return
	}

	// Load site
	site, err := ph.siteManager.LoadSite(siteDomain)
	if err != nil {
		http.Error(w, fmt.Sprintf("Site not found: %v", err), http.StatusNotFound)
		return
	}

	// Get page
	pageManager := common.NewPageManager(site)
	page, err := pageManager.GetPage(slug)
	if err != nil {
		http.Error(w, fmt.Sprintf("Page not found: %v", err), http.StatusNotFound)
		return
	}

	response := ph.convertPageToResponse(page)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreatePage handles POST /admin/api/v1/pages?site=<domain>
func (ph *PageHandler) CreatePage(w http.ResponseWriter, r *http.Request) {
	siteDomain := r.URL.Query().Get("site")
	if siteDomain == "" {
		http.Error(w, "site parameter is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req PageCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Slug == "" {
		http.Error(w, "slug is required", http.StatusBadRequest)
		return
	}
	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	// Load site
	site, err := ph.siteManager.LoadSite(siteDomain)
	if err != nil {
		http.Error(w, fmt.Sprintf("Site not found: %v", err), http.StatusNotFound)
		return
	}

	// Create page metadata
	meta := common.PageMeta{
		Title:         req.Title,
		Description:   req.Description,
		Keywords:      req.Keywords,
		Author:        req.Author,
		Template:      req.Template,
		Layout:        req.Layout,
		IsDraft:       req.IsDraft,
		IsStatic:      req.IsStatic,
		RequireAuth:   req.RequireAuth,
		RequiredRoles: []string{},
		URL:           req.URL,
		Fetch:         req.Fetch,
		Protected:     req.Protected,
		CustomData:    req.CustomData,
	}

	// Set defaults
	if meta.Template == "" {
		meta.Template = "default"
	}
	if meta.Layout == "" {
		meta.Layout = "base"
	}
	if meta.CustomData == nil {
		meta.CustomData = make(map[string]string)
	}

	// Create page
	pageManager := common.NewPageManager(site)
	if err := pageManager.CreatePage(req.Slug, meta, req.Content, req.Sections); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, fmt.Sprintf("Page already exists: %v", err), http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf("Error creating page: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Return created page
	page, err := pageManager.GetPage(req.Slug)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving created page: %v", err), http.StatusInternalServerError)
		return
	}

	response := ph.convertPageToResponse(page)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdatePage handles PUT /admin/api/v1/pages/{slug}?site=<domain>
func (ph *PageHandler) UpdatePage(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	siteDomain := r.URL.Query().Get("site")

	if siteDomain == "" {
		http.Error(w, "site parameter is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req PageCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Load site
	site, err := ph.siteManager.LoadSite(siteDomain)
	if err != nil {
		http.Error(w, fmt.Sprintf("Site not found: %v", err), http.StatusNotFound)
		return
	}

	// Create page metadata
	meta := common.PageMeta{
		Title:         req.Title,
		Description:   req.Description,
		Keywords:      req.Keywords,
		Author:        req.Author,
		Template:      req.Template,
		Layout:        req.Layout,
		IsDraft:       req.IsDraft,
		IsStatic:      req.IsStatic,
		RequireAuth:   req.RequireAuth,
		RequiredRoles: []string{},
		URL:           req.URL,
		Fetch:         req.Fetch,
		Protected:     req.Protected,
		CustomData:    req.CustomData,
	}

	// Set defaults
	if meta.Template == "" {
		meta.Template = "default"
	}
	if meta.Layout == "" {
		meta.Layout = "base"
	}
	if meta.CustomData == nil {
		meta.CustomData = make(map[string]string)
	}

	// Update page
	pageManager := common.NewPageManager(site)
	if err := pageManager.UpdatePage(slug, meta, req.Content, req.Sections); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, fmt.Sprintf("Page not found: %v", err), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Error updating page: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Return updated page
	page, err := pageManager.GetPage(slug)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving updated page: %v", err), http.StatusInternalServerError)
		return
	}

	response := ph.convertPageToResponse(page)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeletePage handles DELETE /admin/api/v1/pages/{slug}?site=<domain>
func (ph *PageHandler) DeletePage(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	siteDomain := r.URL.Query().Get("site")

	if siteDomain == "" {
		http.Error(w, "site parameter is required", http.StatusBadRequest)
		return
	}

	// Load site
	site, err := ph.siteManager.LoadSite(siteDomain)
	if err != nil {
		http.Error(w, fmt.Sprintf("Site not found: %v", err), http.StatusNotFound)
		return
	}

	// Delete page
	pageManager := common.NewPageManager(site)
	if err := pageManager.DeletePage(slug); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, fmt.Sprintf("Page not found: %v", err), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Error deleting page: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
