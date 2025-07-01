package core

import (
	"html/template"
	"io"
	"time"
	"wispy-core/common"
	"wispy-core/core/site/theme"
	"wispy-core/tpl"

	"github.com/go-chi/chi/v5"
)

type Site interface {
	GetID() string
	GetName() string
	GetDomain() string
	GetBaseURL() string
	GetTheme() *theme.Root
	GetContentDir() string
	GetData() map[string]interface{}
	SetData(key string, value interface{})
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetUpdatedAt(t time.Time)
	GetRouter() *chi.Mux
}

type SiteTplEngine interface {
	LoadTemplate(templatePath string) (*template.Template, error)
	RenderTemplate(templatePath string, data tpl.TemplateData) (string, error)
	RenderTemplateTo(w io.Writer, templatePath string, data tpl.TemplateData) error
	ScanPages() ([]string, error)
	GetTrie() *common.Trie
}

// Page represents a rendered page in the system
type Page struct {
	ID          string                 `toml:"id" json:"id"`
	Title       string                 `toml:"title" json:"title"`
	Slug        string                 `toml:"slug" json:"slug"`
	Content     string                 `toml:"content" json:"content"`
	Template    string                 `toml:"template" json:"template"`
	Layout      string                 `toml:"layout" json:"layout"`
	FrontMatter map[string]interface{} `toml:"front_matter" json:"front_matter"`
}

// PageContext contains runtime information for page rendering
type PageContext struct {
	*Page
	Site       *Site
	Components map[string]interface{}
	LocalData  map[string]interface{}
}
