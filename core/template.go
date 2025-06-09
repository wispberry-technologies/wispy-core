package core

import (
	"fmt"
	"strings"

	"wispy-core/common"
	"wispy-core/models"
)

type FunctionMap = models.FunctionMap
type TemplateEngine = models.TemplateEngine
type TemplateTag = models.TemplateTag
type TemplateCtx = *models.TemplateContext

// NewTemplateEngine creates a new TemplateEngine instance.
func NewTemplateEngine(funcMap FunctionMap) *TemplateEngine {
	var value = TemplateEngine{
		StartTag: common.GetEnv("TEMPLATE_DELIMITER_OPEN", "{{"),
		EndTag:   common.GetEnv("TEMPLATE_DELIMITER_CLOSE", "}}"),
	}

	if funcMap == nil {
		value.FuncMap = funcMap
	}

	value.Render = func(raw string, ctx TemplateCtx) (string, []error) {
		return Render(raw, &value, ctx)
	}

	return &value
}

// Render renders the template with the given TemplateContext and returns the result string and any errors.
func Render(raw string, te *TemplateEngine, ctx TemplateCtx) (string, []error) {
	var sb strings.Builder
	var errs []error
	pos := 0

	var StartTag = te.StartTag
	var EndTag = te.EndTag

	for pos < len(raw) {
		start := strings.Index(raw[pos:], StartTag)
		if start == -1 {
			sb.WriteString(raw[pos:])
			break
		}
		start += pos
		sb.WriteString(raw[pos:start])
		end := strings.Index(raw[start+len(StartTag):], EndTag)
		if end == -1 {
			sb.WriteString(raw[start:])
			break
		}
		end += start + len(StartTag)
		tagContents := raw[start+len(StartTag) : end]
		// Check for block tag (e.g., if, for)
		parts := strings.Fields(tagContents)
		if len(parts) > 0 {
			if tag, ok := te.FuncMap[parts[0]]; ok {
				newPos, tagErrs := tag.Render(ctx, &sb, tagContents, raw, end+len(EndTag))
				if len(tagErrs) > 0 {
					errs = append(errs, tagErrs...)
				}
				pos = newPos
				continue
			}
		}
		// Simple tag: try to resolve from context (as TemplateContext.Data)
		var val interface{}
		if ctx != nil && ctx.Data != nil {
			// Support dot notation for nested maps
			val = ResolveDotNotation(ctx.Data, tagContents)
		}
		if val != nil {
			sb.WriteString(fmt.Sprint(val))
		} else {
			// skip unresolved tag (do not write it back)
		}
		pos = end + len(EndTag)
	}
	return sb.String(), errs
}
