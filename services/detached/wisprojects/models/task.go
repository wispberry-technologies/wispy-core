package models

import (
	"time"
)

// Task SQL queries
var (
	TaskSelectByProjectSQL = "SELECT id, project_id, title, status, description, tags, due_date, created_at, updated_at FROM tasks WHERE project_id = ?"
	TaskSelectByIDSQL      = "SELECT id, project_id, title, status, description, tags, due_date, created_at, updated_at FROM tasks WHERE project_id = ? AND id = ?"
	TaskInsertSQL          = "INSERT INTO tasks (project_id, title, status, description, tags, due_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	TaskUpdateSQL          = "UPDATE tasks SET title = ?, status = ?, description = ?, tags = ?, due_date = ?, updated_at = ? WHERE id = ? AND project_id = ?"
	TaskDeleteSQL          = "DELETE FROM tasks WHERE id = ? AND project_id = ?"
)

type Task struct {
	ID          int64      `json:"id"`
	ProjectID   int64      `json:"project_id" validate:"required"`
	Title       string     `json:"title" validate:"required,min=2,max=200"`
	Status      string     `json:"status" validate:"required,oneof=todo doing done blocked"`
	Description string     `json:"description" validate:"max=4000"` // markdown
	Tags        []string   `json:"tags" validate:"dive,max=32"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

const TaskTableSQL = `CREATE TABLE IF NOT EXISTS tasks (
	   id INTEGER PRIMARY KEY AUTOINCREMENT,
	   project_id INTEGER NOT NULL,
	   title TEXT NOT NULL,
	   status TEXT NOT NULL,
	   description TEXT,
	   tags TEXT,
	   due_date DATETIME,
	   created_at DATETIME NOT NULL,
	   updated_at DATETIME NOT NULL,
	   FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
)`
