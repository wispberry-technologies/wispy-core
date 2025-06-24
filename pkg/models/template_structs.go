package models

import (
	"net/http"
	"reflect"
	"strings"
	"time"
)

// TemplateTagFunc is the function signature for custom tag renderers.
type TemplateTagFunc func(ctx *TemplateContext, sb *strings.Builder, tagContents []string, raw string, pos int) (newPos int, errs []error)

// TemplateTag represents a custom tag with a name and a render function.
type TemplateTag struct {
	Name string
	// func(ctx *TemplateContext, sb *strings.Builder, tagContents []string, raw string, pos int) (newPos int, errs []error)
	Render TemplateTagFunc
}

// FunctionMap maps tag names to TemplateTag definitions.
type FunctionMap map[string]TemplateTag

// FilterFunc defines the signature for template filter functions.
// It takes a value and arguments and returns a processed value.
type FilterFunc func(value interface{}, valueType reflect.Type, args []string, ctx *TemplateContext) interface{}

// FilterMap stores available filter functions by name
type FilterMap map[string]FilterFunc

// TemplateEngine is a template engine supporting per-render context and function map.
type TemplateEngine struct {
	FuncMap   FunctionMap
	FilterMap FilterMap
	Render    func(raw string, ctx *TemplateContext) (string, []error)
	// CloneCtx creates a new TemplateContext based on the existing one, merging in NewData.
	CloneCtx func(ctx *TemplateContext, NewData map[string]interface{}) *TemplateContext
	// NewCtx creates a new TemplateContext with the provided data, without merging.
	NewCtx func(ctx *TemplateContext, NewData map[string]interface{}) *TemplateContext
}

type InternalContext struct {
	Flags          map[string]interface{} // Internal flags for the template context
	Blocks         map[string]string
	TemplatesCache map[string]string
	// HtmlDocumentTags is used to handle importing scripts, styles, and other HTML tags in the document head.
	// as well as importing css and js as inline tags.
	HtmlDocumentTags []HtmlDocumentTags
	MetaTags         []HtmlMetaTag
	// ImportedResources tracks imported CSS/JS files to prevent duplicates
	// Key is the resource path, value is the import type ("css", "js", "inline-css", "inline-js")
	ImportedResources map[string]string
}

type UserContext struct {
	Email       string    `json:"email" db:"email"`
	FirstName   string    `json:"first_name" db:"first_name"`
	LastName    string    `json:"last_name" db:"last_name"`
	DisplayName string    `json:"display_name" db:"display_name"`
	Avatar      string    `json:"avatar" db:"avatar"`
	Roles       string    `json:"roles" db:"roles"` // JSON array stored as string
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
type TemplateContext struct {
	InternalContext *InternalContext
	Instance        *SiteInstance
	Page            *Page
	User            *UserContext
	Data            map[string]interface{}
	Errors          []error
	Engine          *TemplateEngine
	Request         *http.Request
}
