package models

import "time"

// Project SQL queries
var (
	ProjectSelectAllSQL  = "SELECT id, name, notes, owner_uuid, color, icon, created_at, updated_at FROM projects"
	ProjectSelectByIDSQL = "SELECT id, name, notes, owner_uuid, color, icon, created_at, updated_at FROM projects WHERE id = ?"
	ProjectInsertSQL     = "INSERT INTO projects (name, notes, owner_uuid, color, icon, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)"
	ProjectUpdateSQL     = "UPDATE projects SET name = ?, notes = ?, owner_uuid = ?, color = ?, icon = ?, updated_at = ? WHERE id = ?"
	ProjectDeleteSQL     = "DELETE FROM projects WHERE id = ?"

	DocumentSelectByProjectSQL = "SELECT id, project_id, title, content, created_at, updated_at FROM documents WHERE project_id = ?"
	DocumentSelectByIDSQL      = "SELECT id, project_id, title, content, created_at, updated_at FROM documents WHERE project_id = ? AND id = ?"
	DocumentInsertSQL          = "INSERT INTO documents (project_id, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?)"
	DocumentUpdateSQL          = "UPDATE documents SET title = ?, content = ?, updated_at = ? WHERE id = ? AND project_id = ?"
	DocumentDeleteSQL          = "DELETE FROM documents WHERE id = ? AND project_id = ?"
)

type Project struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name" validate:"required,min=2,max=100"`
	Notes     string     `json:"notes" validate:"max=1000"`
	OwnerUUID string     `json:"owner_uuid" validate:"required,uuid4"`
	Color     string     `json:"color" validate:"max=32"`
	Icon      string     `json:"icon" validate:"max=64"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Documents []Document `json:"documents,omitempty"`
}

type Document struct {
	ID        int64     `json:"id"`
	ProjectID int64     `json:"project_id" validate:"required"`
	Title     string    `json:"title" validate:"required,min=2,max=200"`
	Content   string    `json:"content" validate:"required"` // markdown
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const ProjectTableSQL = `CREATE TABLE IF NOT EXISTS projects (
	   id INTEGER PRIMARY KEY AUTOINCREMENT,
	   name TEXT NOT NULL,
	   notes TEXT,
	   owner_uuid TEXT NOT NULL,
	   color TEXT,
	   icon TEXT,
	   created_at DATETIME NOT NULL,
	   updated_at DATETIME NOT NULL
)`

const DocumentTableSQL = `CREATE TABLE IF NOT EXISTS documents (
	   id INTEGER PRIMARY KEY AUTOINCREMENT,
	   project_id INTEGER NOT NULL,
	   title TEXT NOT NULL,
	   content TEXT NOT NULL,
	   created_at DATETIME NOT NULL,
	   updated_at DATETIME NOT NULL,
	   FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
)`
