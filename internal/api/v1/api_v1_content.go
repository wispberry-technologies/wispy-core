package api_v1

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"wispy-core/internal/cache"
	"wispy-core/pkg/common"
	"wispy-core/pkg/models"

	"github.com/go-chi/chi/v5"
)

// Helper to get SiteInstance from request context
func getSiteInstanceFromRequest(r *http.Request) *models.SiteInstance {
	ctx := r.Context()
	instance, ok := ctx.Value("siteInstance").(*models.SiteInstance)
	if ok && instance != nil {
		return instance
	}
	return nil
}

func ContentRouter(r chi.Router) {
	// Content routes
	r.Route("/content", func(r chi.Router) {
		// r.Use(AuthMiddleware)
		// List all content
		r.Get("/", ListContent)
		// Get content details
		r.Get("/{content_id}", GetContent)
		// Create new content
		r.Post("/", CreateContent)
		// Update content
		r.Put("/{content_id}", UpdateContent)
		// Delete content
		r.Delete("/{content_id}", DeleteContent)
	})
}

// ListContent returns a list of all content (pages only, with lang support)
func ListContent(w http.ResponseWriter, r *http.Request) {
	siteInstance := getSiteInstanceFromRequest(r)
	if siteInstance == nil {
		common.PlainTextError(w, 500, "Site context missing")
		return
	}
	domain := siteInstance.Domain
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}

	dbCache := siteInstance.DBCache
	db, err := cache.GetConnection(dbCache, domain, "pages")
	if err != nil {
		common.PlainTextError(w, 500, "DB error", err.Error())
		return
	}

	rows, err := db.Query(`SELECT p.id, p.title, p.slug, p.description, p.author, p.layout, c.lang, c.keywords, c.meta_tags, c.content_json FROM pages p LEFT JOIN page_content c ON p.id = c.page_id AND c.lang = ?`, lang)
	if err != nil {
		common.PlainTextError(w, 500, "Query error", err.Error())
		return
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id, title, slug, description, author, layout, lang, keywords, metaTags, contentJSON sql.NullString
		if err := rows.Scan(&id, &title, &slug, &description, &author, &layout, &lang, &keywords, &metaTags, &contentJSON); err != nil {
			continue
		}
		item := map[string]interface{}{
			"id":          id.String,
			"type":        "page",
			"title":       title.String,
			"slug":        slug.String,
			"description": description.String,
			"author":      author.String,
			"layout":      layout.String,
			"lang":        lang.String,
		}
		// Optionally parse keywords/meta_tags/content_json
		item["keywords"] = keywords.String
		item["meta_tags"] = metaTags.String
		item["content_json"] = contentJSON.String
		result = append(result, item)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetContent returns a single content item by ID (pages only, with lang support)
func GetContent(w http.ResponseWriter, r *http.Request) {
	siteInstance := getSiteInstanceFromRequest(r)
	if siteInstance == nil {
		common.PlainTextError(w, 500, "Site context missing")
		return
	}
	domain := siteInstance.Domain
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}
	id := chi.URLParam(r, "content_id")

	dbCache := siteInstance.DBCache
	db, err := cache.GetConnection(dbCache, domain, "pages")
	if err != nil {
		common.PlainTextError(w, 500, "DB error", err.Error())
		return
	}

	row := db.QueryRow(`SELECT p.id, p.title, p.slug, p.description, p.author, p.layout, c.lang, c.keywords, c.meta_tags, c.content_json FROM pages p LEFT JOIN page_content c ON p.id = c.page_id AND c.lang = ? WHERE p.id = ?`, lang, id)
	var pid, title, slug, description, author, layout, clang, keywords, metaTags, contentJSON sql.NullString
	if err := row.Scan(&pid, &title, &slug, &description, &author, &layout, &clang, &keywords, &metaTags, &contentJSON); err != nil {
		common.PlainTextError(w, 404, "Not found", err.Error())
		return
	}
	item := map[string]interface{}{
		"id":           pid.String,
		"type":         "page",
		"title":        title.String,
		"slug":         slug.String,
		"description":  description.String,
		"author":       author.String,
		"layout":       layout.String,
		"lang":         clang.String,
		"keywords":     keywords.String,
		"meta_tags":    metaTags.String,
		"content_json": contentJSON.String,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// CreateContent creates a new content item (page + page_content)
func CreateContent(w http.ResponseWriter, r *http.Request) {
	siteInstance := getSiteInstanceFromRequest(r)
	if siteInstance == nil {
		common.PlainTextError(w, 500, "Site context missing")
		return
	}
	domain := siteInstance.Domain
	dbCache := siteInstance.DBCache
	db, err := cache.GetConnection(dbCache, domain, "pages")
	if err != nil {
		common.PlainTextError(w, 500, "DB error", err.Error())
		return
	}

	var req struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		Author      string `json:"author"`
		Layout      string `json:"layout"`
		Lang        string `json:"lang"`
		Keywords    string `json:"keywords"`
		MetaTags    string `json:"meta_tags"`
		ContentJSON string `json:"content_json"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.PlainTextError(w, 400, "Invalid JSON", err.Error())
		return
	}
	if req.ID == "" {
		req.ID = common.GenerateUUID()
	}
	if req.Lang == "" {
		req.Lang = "en"
	}

	tx, err := db.Begin()
	if err != nil {
		common.PlainTextError(w, 500, "DB error", err.Error())
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`INSERT INTO pages (id, title, slug, description, author, layout, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, req.ID, req.Title, req.Slug, req.Description, req.Author, req.Layout)
	if err != nil {
		common.PlainTextError(w, 500, "Insert page error", err.Error())
		return
	}
	_, err = tx.Exec(`INSERT INTO page_content (id, page_id, lang, keywords, meta_tags, content_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, common.GenerateUUID(), req.ID, req.Lang, req.Keywords, req.MetaTags, req.ContentJSON)
	if err != nil {
		common.PlainTextError(w, 500, "Insert page_content error", err.Error())
		return
	}
	if err := tx.Commit(); err != nil {
		common.PlainTextError(w, 500, "Commit error", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "id": req.ID})
}

// UpdateContent updates an existing content item (page + page_content)
func UpdateContent(w http.ResponseWriter, r *http.Request) {
	siteInstance := getSiteInstanceFromRequest(r)
	if siteInstance == nil {
		common.PlainTextError(w, 500, "Site context missing")
		return
	}
	domain := siteInstance.Domain
	dbCache := siteInstance.DBCache
	db, err := cache.GetConnection(dbCache, domain, "pages")
	if err != nil {
		common.PlainTextError(w, 500, "DB error", err.Error())
		return
	}
	id := chi.URLParam(r, "content_id")

	var req struct {
		Title       string `json:"title"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		Author      string `json:"author"`
		Layout      string `json:"layout"`
		Lang        string `json:"lang"`
		Keywords    string `json:"keywords"`
		MetaTags    string `json:"meta_tags"`
		ContentJSON string `json:"content_json"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.PlainTextError(w, 400, "Invalid JSON", err.Error())
		return
	}
	if req.Lang == "" {
		req.Lang = "en"
	}
	tx, err := db.Begin()
	if err != nil {
		common.PlainTextError(w, 500, "DB error", err.Error())
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE pages SET title=?, slug=?, description=?, author=?, layout=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, req.Title, req.Slug, req.Description, req.Author, req.Layout, id)
	if err != nil {
		common.PlainTextError(w, 500, "Update page error", err.Error())
		return
	}
	res, err := tx.Exec(`UPDATE page_content SET keywords=?, meta_tags=?, content_json=?, updated_at=CURRENT_TIMESTAMP WHERE page_id=? AND lang=?`, req.Keywords, req.MetaTags, req.ContentJSON, id, req.Lang)
	affected, _ := res.RowsAffected()
	if err != nil || affected == 0 {
		// If no row updated, insert new
		_, err = tx.Exec(`INSERT INTO page_content (id, page_id, lang, keywords, meta_tags, content_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, common.GenerateUUID(), id, req.Lang, req.Keywords, req.MetaTags, req.ContentJSON)
		if err != nil {
			common.PlainTextError(w, 500, "Insert page_content error", err.Error())
			return
		}
	}
	if err := tx.Commit(); err != nil {
		common.PlainTextError(w, 500, "Commit error", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "id": id})
}

// DeleteContent deletes a content item (page + all page_content)
func DeleteContent(w http.ResponseWriter, r *http.Request) {
	siteInstance := getSiteInstanceFromRequest(r)
	if siteInstance == nil {
		common.PlainTextError(w, 500, "Site context missing")
		return
	}
	domain := siteInstance.Domain
	dbCache := siteInstance.DBCache
	db, err := cache.GetConnection(dbCache, domain, "pages")
	if err != nil {
		common.PlainTextError(w, 500, "DB error", err.Error())
		return
	}
	id := chi.URLParam(r, "content_id")

	tx, err := db.Begin()
	if err != nil {
		common.PlainTextError(w, 500, "DB error", err.Error())
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM page_content WHERE page_id=?`, id)
	if err != nil {
		common.PlainTextError(w, 500, "Delete page_content error", err.Error())
		return
	}
	_, err = tx.Exec(`DELETE FROM pages WHERE id=?`, id)
	if err != nil {
		common.PlainTextError(w, 500, "Delete page error", err.Error())
		return
	}
	if err := tx.Commit(); err != nil {
		common.PlainTextError(w, 500, "Commit error", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "id": id})
}
