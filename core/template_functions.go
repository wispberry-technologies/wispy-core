package core

import (
	"fmt"
	"os"
	"strings"
	"wispy-core/models"
)

// # Utils to use
// - common.WrapTemplateDelims(tagName) // Wraps the tag name with template delimiters
// - common.FieldsRespectQuotes // Splits a string by spaces while respecting quoted substrings
// - SeekEndTag(tagName) // Finds the end tag position including tag length
// - ResolveFilterChain(filter, ctx, filters) // Resolves a filter chain against the context
// - ResolveDotNotation // Resolves a dot notation path against the context
var IfTemplate = models.TemplateTag{
	Name: "if",
	Render: func(ctx TemplateCtx, sb *strings.Builder, parts []string, raw string, pos int) (newPos int, errs []error) {
		if len(parts) < 2 {
			errs = append(errs, fmt.Errorf("if tag requires a condition"))
			// Try to find and skip to the endif to continue processing
			if endPos, seekErrs := SeekEndTag(raw, pos, "if"); len(seekErrs) == 0 {
				return endPos, errs
			}
			return pos, errs
		}

		// Extract condition (everything after "if")
		condition := strings.Join(parts[1:], " ")

		// Find the matching endif
		endPos, seekErrs := SeekEndTag(raw, pos, "if")
		if len(seekErrs) > 0 {
			errs = append(errs, seekErrs...)
			return pos, errs
		}

		// Extract content between if and endif - need to find the start of content
		// pos is currently at the end of the opening tag
		contentStart := pos
		contentEnd := endPos - len("{% endif %}")
		content := raw[contentStart:contentEnd]

		// Resolve the condition value - be graceful with nil values
		value, _, resolveErrs := ResolveFilterChain(condition, ctx, ctx.Engine.FilterMap)
		if len(resolveErrs) > 0 {
			// Only log filter errors, not unresolved variable errors
			errs = append(errs, resolveErrs...)
		}

		// Render content if condition is truthy (nil is falsy, so unresolved conditions are false)
		if IsTruthy(value) {
			rendered, renderErrs := ctx.Engine.Render(content, ctx)
			if len(renderErrs) > 0 {
				errs = append(errs, renderErrs...)
			}
			sb.WriteString(rendered)
		}

		return endPos, errs
	},
}

var ForTag = models.TemplateTag{
	Name: "for",
	Render: func(ctx TemplateCtx, sb *strings.Builder, parts []string, raw string, pos int) (newPos int, errs []error) {
		if len(parts) < 4 || parts[2] != "in" {
			return pos, []error{fmt.Errorf("for tag requires format: for item in items")}
		}

		itemVar := parts[1]
		collectionExpr := strings.Join(parts[3:], " ")

		// Find the matching endfor
		endPos, seekErrs := SeekEndTag(raw, pos, "for")
		if len(seekErrs) > 0 {
			return pos, seekErrs
		}

		// Extract content between for and endfor
		contentStart := pos
		contentEnd := endPos - len("{% endfor %}")
		content := raw[contentStart:contentEnd]

		// Resolve the collection
		collection, _, resolveErrs := ResolveFilterChain(collectionExpr, ctx, ctx.Engine.FilterMap)
		if len(resolveErrs) > 0 {
			// Only propagate filter errors, not unresolved variable errors
			errs = append(errs, resolveErrs...)
		}

		// If collection is nil (unresolved), just skip the loop - no error
		if collection == nil {
			return endPos, errs
		}

		// Handle different collection types
		switch col := collection.(type) {
		case []interface{}:
			for _, item := range col {
				// Create new context with the loop variable
				newData := map[string]interface{}{itemVar: item}
				loopCtx := ctx.Engine.CloneCtx(ctx, newData)

				// Render the loop content
				rendered, renderErrs := ctx.Engine.Render(content, loopCtx)
				if len(renderErrs) > 0 {
					errs = append(errs, renderErrs...)
				}
				sb.WriteString(rendered)
			}
		case []string:
			for _, item := range col {
				newData := map[string]interface{}{itemVar: item}
				loopCtx := ctx.Engine.CloneCtx(ctx, newData)

				rendered, renderErrs := ctx.Engine.Render(content, loopCtx)
				if len(renderErrs) > 0 {
					errs = append(errs, renderErrs...)
				}
				sb.WriteString(rendered)
			}
		default:
			errs = append(errs, fmt.Errorf("cannot iterate over type %T", collection))
		}

		return endPos, errs
	},
}

var DefineTag = models.TemplateTag{
	Name: "define",
	Render: func(ctx TemplateCtx, sb *strings.Builder, parts []string, raw string, pos int) (newPos int, errs []error) {
		if len(parts) < 2 {
			return pos, []error{fmt.Errorf("define tag requires a block name")}
		}

		blockName := parts[1]

		// Remove quotes if present
		if len(blockName) >= 2 && blockName[0] == '"' && blockName[len(blockName)-1] == '"' {
			blockName = blockName[1 : len(blockName)-1]
		} else if len(blockName) >= 2 && blockName[0] == '\'' && blockName[len(blockName)-1] == '\'' {
			blockName = blockName[1 : len(blockName)-1]
		}

		// Find the matching enddefine
		endPos, seekErrs := SeekEndTag(raw, pos, "define")
		if len(seekErrs) > 0 {
			return pos, seekErrs
		}

		// Extract content between define and enddefine
		contentStart := pos
		contentEnd := endPos - len("{% enddefine %}")
		content := raw[contentStart:contentEnd]

		// Store the block content in the context for later use
		ctx.InternalContext.Blocks[blockName] = content

		// Define tags don't output anything directly
		return endPos, errs
	},
}

var RenderTag = models.TemplateTag{
	Name: "render",
	Render: func(ctx TemplateCtx, sb *strings.Builder, parts []string, raw string, pos int) (newPos int, errs []error) {
		if len(parts) < 2 {
			return pos, []error{fmt.Errorf("render tag requires a template name")}
		}

		templateName := parts[1]

		// Remove quotes if present
		if len(templateName) >= 2 && templateName[0] == '"' && templateName[len(templateName)-1] == '"' {
			templateName = templateName[1 : len(templateName)-1]
		} else if len(templateName) >= 2 && templateName[0] == '\'' && templateName[len(templateName)-1] == '\'' {
			templateName = templateName[1 : len(templateName)-1]
		}

		// Check cache first
		var templateContent string
		if cached, exists := ctx.InternalContext.TemplatesCache[templateName]; exists {
			templateContent = cached
		} else {
			// Load template from file system
			// Assume we can extract domain from context or it's stored somewhere accessible
			if ctx.Request != nil {
				domain := ctx.Request.Host
				// Try different template locations
				templatePaths := []string{
					"templates/sections/" + templateName + ".html",
					"templates/partials/" + templateName + ".html",
					"templates/" + templateName + ".html",
				}

				for _, path := range templatePaths {
					fullPath := "/Users/theo/Desktop/wispy-core/sites/" + domain + "/" + path
					if content, err := os.ReadFile(fullPath); err == nil {
						templateContent = string(content)
						// Cache for future use
						ctx.InternalContext.TemplatesCache[templateName] = templateContent
						break
					}
				}
			}
		}

		if templateContent == "" {
			return pos, []error{fmt.Errorf("template not found: %s", templateName)}
		}

		// Render the included template with current context
		rendered, renderErrs := ctx.Engine.Render(templateContent, ctx)
		if len(renderErrs) > 0 {
			errs = append(errs, renderErrs...)
		}
		sb.WriteString(rendered)

		return pos, errs
	},
}

var BlockTag = models.TemplateTag{
	Name: "block",
	Render: func(ctx TemplateCtx, sb *strings.Builder, parts []string, raw string, pos int) (newPos int, errs []error) {
		if len(parts) < 2 {
			return pos, []error{fmt.Errorf("block tag requires a block name")}
		}

		blockName := parts[1]

		// Remove quotes if present
		if len(blockName) >= 2 && blockName[0] == '"' && blockName[len(blockName)-1] == '"' {
			blockName = blockName[1 : len(blockName)-1]
		} else if len(blockName) >= 2 && blockName[0] == '\'' && blockName[len(blockName)-1] == '\'' {
			blockName = blockName[1 : len(blockName)-1]
		}

		// Find the matching endblock
		endPos, seekErrs := SeekEndTag(raw, pos, "block")
		if len(seekErrs) > 0 {
			errs = append(errs, seekErrs...)
			return pos, errs
		}

		// Extract default content between block and endblock
		contentStart := pos
		endBlockStart := endPos - len("{% endblock %}")
		if endBlockStart > contentStart {
			defaultContent := raw[contentStart:endBlockStart]

			// Look up the block content (if defined elsewhere)
			if content, exists := ctx.InternalContext.Blocks[blockName]; exists {
				// Render the defined block content
				rendered, renderErrs := ctx.Engine.Render(content, ctx)
				if len(renderErrs) > 0 {
					errs = append(errs, renderErrs...)
				}
				sb.WriteString(rendered)
			} else {
				// Render the default content between the block tags
				rendered, renderErrs := ctx.Engine.Render(defaultContent, ctx)
				if len(renderErrs) > 0 {
					errs = append(errs, renderErrs...)
				}
				sb.WriteString(rendered)
			}
		} else {
			// No content between tags, just render defined content if it exists
			if content, exists := ctx.InternalContext.Blocks[blockName]; exists {
				rendered, renderErrs := ctx.Engine.Render(content, ctx)
				if len(renderErrs) > 0 {
					errs = append(errs, renderErrs...)
				}
				sb.WriteString(rendered)
			}
		}

		return endPos, errs
	},
}

// DefaultFunctionMap provides all built-in tags.
func DefaultFunctionMap() models.FunctionMap {
	return models.FunctionMap{
		"if":     IfTemplate,
		"for":    ForTag,
		"define": DefineTag,
		"render": RenderTag,
		"block":  BlockTag,
	}
}
