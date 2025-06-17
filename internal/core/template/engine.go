package template

import (
	"fmt"
	"maps"
	"net/http"
	"strings"

	"wispy-core/pkg/common"
	"wispy-core/pkg/models"
)

// -------------------
// Constructors
// -------------------

// NewTemplateEngine creates a new TemplateEngine instance.
func NewTemplateEngine(data map[string]interface{}, request *http.Request, site *models.SiteInstance, page *models.Page) (engine *models.TemplateEngine, ctx TemplateCtx) {
	engine = &models.TemplateEngine{
		FuncMap:   GetDefaultFunctions(),
		FilterMap: GetDefaultFilters(),
	}

	if data == nil {
		data = make(map[string]interface{})
	}
	// Always ensure InternalContext is a non-nil map[string]string
	var internal = &models.InternalContext{
		Flags:             make(map[string]interface{}),
		Blocks:            make(map[string]string),
		TemplatesCache:    make(map[string]string),
		HtmlDocumentTags:  []models.HtmlDocumentTags{},
		MetaTags:          []models.HtmlMetaTag{},
		ImportedResources: make(map[string]string),
	}

	ctx = &models.TemplateContext{
		Data:            data,
		Engine:          engine,
		Page:            page,
		Instance:        site,
		InternalContext: internal,
		Errors:          []error{},
	}

	engine.Render = func(raw string, ctx TemplateCtx) (string, []error) {
		// For simplicity just handle both types of tags in the same pass
		return Render(raw, engine, ctx)
	}

	engine.CloneCtx = func(ctx TemplateCtx, newData map[string]interface{}) *models.TemplateContext {
		clonedCtxData := maps.Clone[map[string]interface{}](ctx.Data)
		for k, v := range newData {
			// If the key already exists, we overwrite it with the new value
			clonedCtxData[k] = v
		}
		// Create a new context with the existing internal context and new data
		newCtx := &models.TemplateContext{
			InternalContext: ctx.InternalContext,
			Data:            clonedCtxData,
			Engine:          ctx.Engine,
			Errors:          ctx.Errors,
		}
		// Copy the request if it exists
		if ctx.Request != nil {
			newCtx.Request = ctx.Request
		}
		// Clone the context
		return newCtx
	}

	return engine, ctx
}

// Render renders the template with the given TemplateContext and returns the result string and any errors.
// This function processes template tags (e.g., {{ variable }}, {% if condition %}) and replaces them with
// their corresponding values or executes the associated logic.
func Render(raw string, te *models.TemplateEngine, ctx TemplateCtx) (string, []error) {
	if len(raw) == 0 {
		return "", nil
	}

	// Pre-allocate with estimated capacity to reduce memory allocations
	var errs []error

	// Pre-allocate builder with estimated capacity for efficiency
	sb := strings.Builder{}
	sb.Grow(len(raw) + len(raw)/4) // Estimate 25% expansion for typical templates

	pos := 0
	rawLen := len(raw)

	for pos < rawLen {
		// Find the next template tag or variable
		varStart := strings.Index(raw[pos:], "{{")
		tagStart := strings.Index(raw[pos:], "{%")

		// Determine which comes first (or if neither exists)
		nextTag := -1
		isVariable := false

		if varStart != -1 && (tagStart == -1 || varStart < tagStart) {
			nextTag = pos + varStart
			isVariable = true
		} else if tagStart != -1 {
			nextTag = pos + tagStart
			isVariable = false
		}

		// If no more tags found, append the rest and break
		if nextTag == -1 {
			sb.WriteString(raw[pos:])
			break
		}

		// Append content before the tag
		sb.WriteString(raw[pos:nextTag])

		if isVariable {
			// Process variable interpolation {{ }}
			newPos, err := processVariable(raw, nextTag, &sb, ctx, te.FilterMap)
			if err != nil {
				errs = append(errs, err)
				// Continue processing from the end of the variable tag or skip past {{
				if newPos > nextTag {
					pos = newPos
				} else {
					// Try to find the closing }} and skip past it, or skip minimal amount
					if closePos := strings.Index(raw[nextTag:], "}}"); closePos != -1 {
						pos = nextTag + closePos + 2
					} else {
						pos = nextTag + 2 // Skip past {{ to avoid infinite loop
					}
				}
			} else {
				pos = newPos
			}
		} else {
			// Process template tag {% %}
			newPos, tagErrs := processTemplateTag(raw, nextTag, &sb, ctx, te.FuncMap)
			if len(tagErrs) > 0 {
				errs = append(errs, tagErrs...)
				// Continue processing from the end of the tag or skip past {%
				if newPos > nextTag {
					pos = newPos
				} else {
					// Try to find the closing %} and skip past it, or skip minimal amount
					if closePos := strings.Index(raw[nextTag:], "%}"); closePos != -1 {
						pos = nextTag + closePos + 2
					} else {
						pos = nextTag + 2 // Skip past {% to avoid infinite loop
					}
				}
			} else {
				pos = newPos
			}
		}
	}

	return sb.String(), errs
}

// processVariable handles {{ variable | filter }} interpolation
func processVariable(raw string, pos int, sb *strings.Builder, ctx TemplateCtx, filters models.FilterMap) (newPos int, err error) {
	// Find the closing }}
	endPos := strings.Index(raw[pos:], "}}")
	if endPos == -1 {
		return pos + 2, fmt.Errorf("unclosed variable tag at position %d", pos)
	}
	endPos += pos

	// Extract the content between {{ and }}
	content := strings.TrimSpace(raw[pos+2 : endPos])
	if len(content) == 0 {
		// Empty variable tag - just skip it
		return endPos + 2, nil
	}

	// Resolve the filter chain - be graceful with errors
	value, _, resolveErrs := ResolveFilterChain(content, ctx, filters)
	if len(resolveErrs) > 0 {
		// Instead of returning an error, just log it and render nothing
		err = fmt.Errorf("warning: could not resolve variable '%s': %v", content, resolveErrs)
		// Don't write anything to the builder, just continue
		return endPos + 2, err
	}

	// Convert to string and write to builder only if value exists
	if value != nil {
		sb.WriteString(fmt.Sprintf("%v", value))
	}
	// If value is nil, render nothing (empty string)

	return endPos + 2, nil
}

// processTemplateTag handles {% tag %} processing
func processTemplateTag(raw string, pos int, sb *strings.Builder, ctx TemplateCtx, funcMap models.FunctionMap) (newPos int, errs []error) {
	// Find the closing %}
	endPos := strings.Index(raw[pos:], "%}")
	if endPos == -1 {
		return pos + 2, []error{fmt.Errorf("unclosed template tag at position %d", pos)}
	}
	endPos += pos

	// Extract the content between {% and %}
	content := strings.TrimSpace(raw[pos+2 : endPos])
	if len(content) == 0 {
		return endPos + 2, []error{fmt.Errorf("empty template tag at position %d", pos)}
	}

	// Parse the tag name and arguments
	parts := common.FieldsRespectQuotes(content)
	if len(parts) == 0 {
		return endPos + 2, []error{fmt.Errorf("invalid template tag format at position %d", pos)}
	}

	tagName := parts[0]

	// Look up the tag in the function map
	if tag, exists := funcMap[tagName]; exists {
		return tag.Render(ctx, sb, parts, raw, endPos+2)
	}

	// Unknown tag - return error but continue processing
	return endPos + 2, []error{fmt.Errorf("unknown template tag: %s at position %d", tagName, pos)}
}
