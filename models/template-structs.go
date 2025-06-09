package models

import (
	"net/http"
	"strings"
)

// TemplateTagFunc is the function signature for custom tag renderers.
type TemplateTagFunc func(ctx *TemplateContext, sb *strings.Builder, tagContents, raw string, pos int) (newPos int, errs []error)

// TemplateTag represents a custom tag with a name and a render function.
type TemplateTag struct {
	Name   string
	Render TemplateTagFunc
}

// FunctionMap maps tag names to TemplateTag definitions.
type FunctionMap map[string]TemplateTag

// TemplateEngine is a template engine supporting per-render context and function map.
type TemplateEngine struct {
	StartTag string
	EndTag   string
	FuncMap  FunctionMap
	Render   func(raw string, ctx *TemplateContext) (string, []error)
}

type TemplateContext struct {
	InternalContext interface{}
	Data            interface{}
	Errors          []error
	Engine          *TemplateEngine
	Request         *http.Request
}

type TemplateSiteEngineContext struct {
	InternalContext interface{}
	Data            interface{}
	Errors          []error
	Engine          *TemplateEngine
	Site            *SiteInstance
	Page            *Page
}
