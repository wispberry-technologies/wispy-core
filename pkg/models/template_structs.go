package models

import (
	"net/http"
	"reflect"
	"strings"
)

// TemplateTagFunc is the function signature for custom tag renderers.
type TemplateTagFunc func(ctx *TemplateContext, sb *strings.Builder, tagContents []string, raw string, pos int) (newPos int, errs []error)

// TemplateTag represents a custom tag with a name and a render function.
type TemplateTag struct {
	Name   string
	Render TemplateTagFunc
}

// FunctionMap maps tag names to TemplateTag definitions.
type FunctionMap map[string]TemplateTag

// FilterFunc defines the signature for template filter functions.
// It takes a value and arguments and returns a processed value.
type FilterFunc func(value interface{}, valueType reflect.Type, args []string) interface{}

// FilterMap stores available filter functions by name
type FilterMap map[string]FilterFunc

// TemplateEngine is a template engine supporting per-render context and function map.
type TemplateEngine struct {
	FuncMap   FunctionMap
	FilterMap FilterMap
	Render    func(raw string, ctx *TemplateContext) (string, []error)
	CloneCtx  func(ctx *TemplateContext, NewData map[string]interface{}) *TemplateContext
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

type TemplateContext struct {
	InternalContext *InternalContext
	Instance        *SiteInstance
	Page            *Page
	Data            map[string]interface{}
	Errors          []error
	Engine          *TemplateEngine
	Request         *http.Request
}
