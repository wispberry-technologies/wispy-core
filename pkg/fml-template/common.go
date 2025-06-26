package template

import "strings"

// (Engine, StringBuilder, Args, RawString, Pos) -> (NewPos, Errors)
type TemplateTagResolver func(engine *Engine, sb *strings.Builder, args []string, raw string, pos int) (newPos int, errs []error)
type TemplateTag struct {
	Name        string
	Description string
	Handler     TemplateTagResolver
}

type TemplateFuncResolver func(engine *Engine, args []string) (value any, errs []error)
type TemplateFunc struct {
	Name        string
	Description string
	Handler     TemplateFuncResolver
}

type EngineSanitizer struct {
	Enabled        bool
	SanitizeFunc   func(input string) string
	SanitizePolicy any // Placeholder for any specific sanitization policy or configuration
}
