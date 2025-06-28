package page

import (
	"wispy-core/core/site"
)

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
	Site       *site.Site
	Components map[string]interface{}
	LocalData  map[string]interface{}
}
