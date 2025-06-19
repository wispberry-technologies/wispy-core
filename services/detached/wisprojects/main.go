package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"wisprojects/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

// Example: API key validation middleware
func apiKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-Wispy-Api-Key")
		// In production, check against a DB or env var, or validate with Wispy Core
		validKey := os.Getenv("WISPY_CORE_PROXY_KEY")
		if apiKey == "" || apiKey != validKey {
			http.Error(w, "Unauthorized: missing or invalid API key", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Load .env file if it exists
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Printf("Error loading .env file: %v", err)
		}
	}

	// ensure the data directory exists
	if err := os.MkdirAll("./data", os.ModePerm); err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite3", "./data/projects.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create tables if not exists
	_, err = db.Exec(models.ProjectTableSQL)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(models.TaskTableSQL)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(models.DocumentTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(apiKeyMiddleware)
	r.Use(middleware.Recoverer)

	// Project endpoints
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, notes, owner_uuid, color, icon, created_at, updated_at FROM projects")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var projects []models.Project
		for rows.Next() {
			var p models.Project
			var created, updated string
			if err := rows.Scan(&p.ID, &p.Name, &p.Notes, &p.OwnerUUID, &p.Color, &p.Icon, &created, &updated); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			p.CreatedAt, _ = time.Parse(time.RFC3339, created)
			p.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
			// Load documents for each project
			docRows, err := db.Query("SELECT id, project_id, title, content, created_at, updated_at FROM documents WHERE project_id = ?", p.ID)
			if err == nil {
				var docs []models.Document
				for docRows.Next() {
					var d models.Document
					var dCreated, dUpdated string
					if err := docRows.Scan(&d.ID, &d.ProjectID, &d.Title, &d.Content, &dCreated, &dUpdated); err == nil {
						d.CreatedAt, _ = time.Parse(time.RFC3339, dCreated)
						d.UpdatedAt, _ = time.Parse(time.RFC3339, dUpdated)
						docs = append(docs, d)
					}
				}
				p.Documents = docs
				docRows.Close()
			}
			projects = append(projects, p)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var p models.Project
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		now := time.Now().UTC()
		res, err := db.Exec("INSERT INTO projects (name, notes, owner_uuid, color, icon, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)", p.Name, p.Notes, p.OwnerUUID, p.Color, p.Icon, now.Format(time.RFC3339), now.Format(time.RFC3339))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		id, _ := res.LastInsertId()
		p.ID = id
		p.CreatedAt = now
		p.UpdatedAt = now
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(p)
	})

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var p models.Project
		var created, updated string
		err := db.QueryRow("SELECT id, name, notes, owner_uuid, color, icon, created_at, updated_at FROM projects WHERE id = ?", id).Scan(&p.ID, &p.Name, &p.Notes, &p.OwnerUUID, &p.Color, &p.Icon, &created, &updated)
		// Load documents for this project
		docRows, err2 := db.Query("SELECT id, project_id, title, content, created_at, updated_at FROM documents WHERE project_id = ?", id)
		if err2 == nil {
			var docs []models.Document
			for docRows.Next() {
				var d models.Document
				var dCreated, dUpdated string
				if err := docRows.Scan(&d.ID, &d.ProjectID, &d.Title, &d.Content, &dCreated, &dUpdated); err == nil {
					d.CreatedAt, _ = time.Parse(time.RFC3339, dCreated)
					d.UpdatedAt, _ = time.Parse(time.RFC3339, dUpdated)
					docs = append(docs, d)
				}
			}
			p.Documents = docs
			docRows.Close()
		}
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		} else if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		p.CreatedAt, _ = time.Parse(time.RFC3339, created)
		p.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
	})

	r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var p models.Project
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		now := time.Now().UTC()
		_, err := db.Exec("UPDATE projects SET name = ?, notes = ?, owner_uuid = ?, color = ?, icon = ?, updated_at = ? WHERE id = ?", p.Name, p.Notes, p.OwnerUUID, p.Color, p.Icon, now.Format(time.RFC3339), id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	r.Delete("/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		_, err := db.Exec("DELETE FROM projects WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// Task endpoints
	r.Get("/{project_id}/tasks", func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "project_id")
		rows, err := db.Query("SELECT id, project_id, title, status, description, due_date, created_at, updated_at FROM tasks WHERE project_id = ?", projectID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var tasks []models.Task
		for rows.Next() {
			var t models.Task
			var due, created, updated sql.NullString
			if err := rows.Scan(&t.ID, &t.ProjectID, &t.Title, &t.Status, &t.Description, &due, &created, &updated); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			if due.Valid {
				parsed, _ := time.Parse(time.RFC3339, due.String)
				t.DueDate = &parsed
			}
			t.CreatedAt, _ = time.Parse(time.RFC3339, created.String)
			t.UpdatedAt, _ = time.Parse(time.RFC3339, updated.String)
			tasks = append(tasks, t)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	})

	r.Post("/{project_id}/tasks", func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "project_id")
		var t models.Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		now := time.Now().UTC()
		var due interface{} = nil
		if t.DueDate != nil {
			due = t.DueDate.Format(time.RFC3339)
		}
		res, err := db.Exec("INSERT INTO tasks (project_id, title, status, description, due_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)", projectID, t.Title, t.Status, t.Description, due, now.Format(time.RFC3339), now.Format(time.RFC3339))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		id, _ := res.LastInsertId()
		t.ID = id
		t.ProjectID, _ = parseInt64(projectID)
		t.CreatedAt = now
		t.UpdatedAt = now
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(t)
	})

	r.Get("/{project_id}/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "project_id")
		id := chi.URLParam(r, "id")
		var t models.Task
		var due, created, updated sql.NullString
		err := db.QueryRow("SELECT id, project_id, title, status, description, due_date, created_at, updated_at FROM tasks WHERE project_id = ? AND id = ?", projectID, id).Scan(&t.ID, &t.ProjectID, &t.Title, &t.Status, &t.Description, &due, &created, &updated)
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		} else if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if due.Valid {
			parsed, _ := time.Parse(time.RFC3339, due.String)
			t.DueDate = &parsed
		}
		t.CreatedAt, _ = time.Parse(time.RFC3339, created.String)
		t.UpdatedAt, _ = time.Parse(time.RFC3339, updated.String)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(t)
	})

	r.Put("/projects/{project_id}/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "project_id")
		id := chi.URLParam(r, "id")
		var t models.Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		now := time.Now().UTC()
		var due interface{} = nil
		if t.DueDate != nil {
			due = t.DueDate.Format(time.RFC3339)
		}
		_, err = db.Exec("UPDATE tasks SET title = ?, status = ?, description = ?, due_date = ?, updated_at = ? WHERE id = ? AND project_id = ?", t.Title, t.Status, t.Description, due, now.Format(time.RFC3339), id, projectID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// Document endpoints
	r.Get("/{project_id}/documents", func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "project_id")
		rows, err := db.Query("SELECT id, project_id, title, content, created_at, updated_at FROM documents WHERE project_id = ?", projectID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var docs []models.Document
		for rows.Next() {
			var d models.Document
			var created, updated string
			if err := rows.Scan(&d.ID, &d.ProjectID, &d.Title, &d.Content, &created, &updated); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			d.CreatedAt, _ = time.Parse(time.RFC3339, created)
			d.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
			docs = append(docs, d)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(docs)
	})

	r.Post("/projects/{project_id}/documents", func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "project_id")
		var d models.Document
		if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		now := time.Now().UTC()
		res, err := db.Exec("INSERT INTO documents (project_id, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", projectID, d.Title, d.Content, now.Format(time.RFC3339), now.Format(time.RFC3339))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		id, _ := res.LastInsertId()
		d.ID = id
		d.ProjectID, _ = parseInt64(projectID)
		d.CreatedAt = now
		d.UpdatedAt = now
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(d)
	})

	r.Get("/{project_id}/documents/{id}", func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "project_id")
		id := chi.URLParam(r, "id")
		var d models.Document
		var created, updated string
		err := db.QueryRow("SELECT id, project_id, title, content, created_at, updated_at FROM documents WHERE project_id = ? AND id = ?", projectID, id).Scan(&d.ID, &d.ProjectID, &d.Title, &d.Content, &created, &updated)
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		} else if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		d.CreatedAt, _ = time.Parse(time.RFC3339, created)
		d.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(d)
	})

	r.Put("/projects/{project_id}/documents/{id}", func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "project_id")
		id := chi.URLParam(r, "id")
		var d models.Document
		if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		now := time.Now().UTC()
		_, err := db.Exec("UPDATE documents SET title = ?, content = ?, updated_at = ? WHERE id = ? AND project_id = ?", d.Title, d.Content, now.Format(time.RFC3339), id, projectID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	r.Delete("/{project_id}/documents/{id}", func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "project_id")
		id := chi.URLParam(r, "id")
		_, err := db.Exec("DELETE FROM documents WHERE id = ? AND project_id = ?", id, projectID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	r.Delete("/projects/{project_id}/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "project_id")
		id := chi.URLParam(r, "id")
		_, err := db.Exec("DELETE FROM tasks WHERE id = ? AND project_id = ?", id, projectID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("wisprojects API running on :%s", port)
	http.ListenAndServe(":"+port, r)
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
