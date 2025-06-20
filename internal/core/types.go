package core

import (
	"wispy-core/pkg/auth"
	"wispy-core/pkg/template"
)

// Instance represents a site instance
type Instance struct {
	SitePath       string
	TemplateEngine *template.Engine
	AuthConfig     *auth.Config
}

// PageData represents a page's data
type PageData struct {
	Template string
	Data     map[string]interface{}
	Title    string
	Path     string
}
