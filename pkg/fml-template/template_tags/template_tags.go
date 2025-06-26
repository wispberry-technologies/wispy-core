package template_tags

import (
	"fmt"
	"reflect"
	"strings"
	"wispy-core/pkg/template"
)

// isTruthy determines if a value should be considered true in a template condition
func isTruthy(value any) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v != ""
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() != 0
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() != 0
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0
	case []any:
		return len(v) > 0
	case map[string]any:
		return len(v) > 0
	default:
		// For other types, check if it's a zero value
		return !reflect.ValueOf(v).IsZero()
	}
}

// === RULES ===
// 1. Template tag will always start with the StartDelim and end with the EndDelim
//    e.g., {{ tagName .arg1 .arg2 }}
// 2. References to data will always be prefixed with a dot or $ if local variable (e.g., .variableName) (e.g., $variableName)
// 3.

// ====== Functions to use in Handlers ======
// e.SplitTagParts(contents string) (tag string, args []string)
// e.ResolveArgFuncsIfAny(args []string) (any, []error)
// // (Pipe output from function to next function)
// -> ResolveFunction(fnName string, args []string) (any, []error)
// -> -> e.ResolveValue(v string) (any, []error)
// ==========================================

// TODO: Implement SeekClosingTag
// -> SeekClosingTag must also handle nested tags correctly

/*
Example of a template tag definition
{{ if .condition }}
Value A
*/
var TemplateIfTag = &template.TemplateTag{
	Name:        "if",
	Description: "Conditional rendering based on the truthiness of the expression.",
	Handler: func(engine *template.Engine, sb *strings.Builder, args []string, raw string, pos int) (newPos int, errs []error) {
		if len(args) == 0 {
			return pos, []error{fmt.Errorf("'if' tag requires at least one argument")}
		}

		// Evaluate the condition
		condition, err := engine.ResolveArguments(nil, args)
		if err != nil {
			return pos, []error{err}
		}

		// Check if condition is truthy
		isConditionTruthy := isTruthy(condition)

		// We need to find the end tag or else tag
		endPos, content, errs := engine.SeekClosingTag(raw, "if", pos)
		if len(errs) > 0 {
			return endPos, errs
		}

		// Check for the presence of an 'else' tag in the raw string between pos and endPos
		elseIndex := strings.Index(raw[pos:endPos], engine.StartDelim+" else "+engine.EndDelim)
		if elseIndex != -1 {
			// We have an else tag
			if isConditionTruthy {
				// Render the "if" part
				ifContent := raw[pos : pos+elseIndex]
				result, renderErrs := engine.Render(ifContent, engine.GetLocalDataContext())
				if len(renderErrs) > 0 {
					errs = append(errs, renderErrs...)
				}
				sb.WriteString(result)
			} else {
				// Render the "else" part
				elseStartPos := pos + elseIndex + len(engine.StartDelim+" else "+engine.EndDelim)
				elseContent := raw[elseStartPos:endPos]
				result, renderErrs := engine.Render(elseContent, engine.GetLocalDataContext())
				if len(renderErrs) > 0 {
					errs = append(errs, renderErrs...)
				}
				sb.WriteString(result)
			}
		} else if isConditionTruthy {
			// No else tag, just render the if content if condition is true
			result, renderErrs := engine.Render(content, engine.GetLocalDataContext())
			if len(renderErrs) > 0 {
				errs = append(errs, renderErrs...)
			}
			sb.WriteString(result)
		}

		return endPos, errs
	},
}

/*
Example of a template tag definition
{{ if .condition }}
Value A
{{ else }}
Value B
{{ end }}
*/
var TemplateElseTag = &template.TemplateTag{
	Name:        "else",
	Description: "Conditional rendering for the 'else' branch.",
	Handler: func(engine *template.Engine, sb *strings.Builder, args []string, raw string, pos int) (newPos int, errs []error) {
		// Check if we should render the else block
		shouldRender, exists := engine.Flags["else"]
		if !exists {
			return pos, []error{fmt.Errorf("'else' tag found without preceding 'if' tag")}
		}

		// Find the end of the else block
		endPos, content, seekErrs := engine.SeekClosingTag(raw, "if", pos)
		if len(seekErrs) > 0 {
			return endPos, seekErrs
		}

		// If shouldRender is true, we render the content of the else block
		if shouldRender {
			result, renderErrs := engine.Render(content, engine.GetLocalDataContext())
			if len(renderErrs) > 0 {
				errs = append(errs, renderErrs...)
			}
			sb.WriteString(result)
		}

		// Reset the else flag
		engine.Flags["else"] = false

		return endPos, errs
	},
}

/*
Example of a template tag definition
{{ range .variables }}
<p>{{ . }}</p>
{{ end }}
--------
// in this example the args are resolved by
{{ range slice("foo" "bar") }}
<p>{{ . }}</p>
{{ end }}
*/
var TemplateRangeTag = &template.TemplateTag{
	Name:        "range",
	Description: "Iterates over a collection and renders the content for each item.",
	Handler: func(engine *template.Engine, sb *strings.Builder, args []string, raw string, pos int) (newPos int, errs []error) {
		if len(args) == 0 {
			return pos, []error{fmt.Errorf("'range' tag requires at least one argument")}
		}

		// Evaluate the collection to iterate over
		collection, err := engine.ResolveArguments(nil, args)
		if err != nil {
			return pos, []error{err}
		}

		// Find the content between range and endrange
		endPos, content, seekErrs := engine.SeekClosingTag(raw, "range", pos)
		if len(seekErrs) > 0 {
			return endPos, seekErrs
		}

		// Iterate over the collection based on its type
		switch v := collection.(type) {
		case []any:
			for i, item := range v {
				// Save current context
				oldContext := engine.GetLocalDataContext()

				// Create a local context with the item as "." and index info
				localContext := make(map[string]any)
				localContext["."] = item
				localContext["index"] = i

				// Set local context for this iteration
				engine.SetLocalDataContext(localContext)

				// Render the content for this item
				result, renderErrs := engine.Render(content, localContext)
				if len(renderErrs) > 0 {
					errs = append(errs, renderErrs...)
				}
				sb.WriteString(result)

				// Restore previous context
				engine.SetLocalDataContext(oldContext)
			}
		case map[string]any:
			index := 0
			for key, value := range v {
				// Save current context
				oldContext := engine.GetLocalDataContext()

				// Create a local context with the key and value
				localContext := make(map[string]any)
				localContext["."] = value
				localContext["key"] = key
				localContext["index"] = index

				// Set local context for this iteration
				engine.SetLocalDataContext(localContext)

				// Render the content for this item
				result, renderErrs := engine.Render(content, localContext)
				if len(renderErrs) > 0 {
					errs = append(errs, renderErrs...)
				}
				sb.WriteString(result)

				// Restore previous context
				engine.SetLocalDataContext(oldContext)
				index++
			}
		case string:
			for i, char := range v {
				// Save current context
				oldContext := engine.GetLocalDataContext()

				// Create a local context with the char as "." and index info
				localContext := make(map[string]any)
				localContext["."] = string(char)
				localContext["index"] = i

				// Set local context for this iteration
				engine.SetLocalDataContext(localContext)

				// Render the content for this item
				result, renderErrs := engine.Render(content, localContext)
				if len(renderErrs) > 0 {
					errs = append(errs, renderErrs...)
				}
				sb.WriteString(result)

				// Restore previous context
				engine.SetLocalDataContext(oldContext)
			}
		default:
			errs = append(errs, fmt.Errorf("'range' tag requires an iterable collection, got %T", collection))
		}

		return endPos, errs
	},
}

/*
Example of a template tag definition
{{ with .context }}
<p>{{ . }}</p>
{{ end }}
*/
var TemplateWithTag = &template.TemplateTag{
	Name:        "with",
	Description: "Sets a new context for the enclosed content.",
	Handler: func(engine *template.Engine, sb *strings.Builder, args []string, raw string, pos int) (newPos int, errs []error) {
		if len(args) == 0 {
			return pos, []error{fmt.Errorf("'with' tag requires at least one argument")}
		}

		// Evaluate the context value
		contextValue, err := engine.ResolveArguments(nil, args)
		if err != nil {
			return pos, []error{err}
		}

		// Check if the context value is truthy
		if !isTruthy(contextValue) {
			// If the context value is falsy, skip the content
			endPos, _, seekErrs := engine.SeekClosingTag(raw, "with", pos)
			if len(seekErrs) > 0 {
				errs = append(errs, seekErrs...)
			}
			return endPos, errs
		}

		// Find the content between with and endwith
		endPos, content, seekErrs := engine.SeekClosingTag(raw, "with", pos)
		if len(seekErrs) > 0 {
			return endPos, seekErrs
		}

		// Set up the new context based on the type of the context value
		localContext := make(map[string]any)
		localContext["."] = contextValue

		// Map type context values make their keys available directly
		if mapValue, ok := contextValue.(map[string]any); ok {
			for k, v := range mapValue {
				localContext[k] = v
			}
		}

		// Render the content with the new context
		result, renderErrs := engine.Render(content, localContext)
		if len(renderErrs) > 0 {
			errs = append(errs, renderErrs...)
		}
		sb.WriteString(result)

		return endPos, errs
	},
}

/*
Example of a template tag definition
{{ define "name" }}
<p>{{ .Example.Name }}</p>
{{ end }}
*/
var TemplateDefineTag = &template.TemplateTag{
	Name:        "define",
	Description: "Defines a reusable template block.",
	Handler: func(engine *template.Engine, sb *strings.Builder, args []string, raw string, pos int) (newPos int, errs []error) {
		if len(args) == 0 {
			return pos, []error{fmt.Errorf("'define' tag requires a block name argument")}
		}

		// The first argument is the block name
		blockName := strings.Trim(args[0], "\"'")

		// Find the content between define and enddefine
		endPos, content, seekErrs := engine.SeekClosingTag(raw, "define", pos)
		if len(seekErrs) > 0 {
			return endPos, seekErrs
		}

		// Store the raw content in the engine's blocks
		engine.AddBlock(blockName, content)

		// Define tags don't render anything directly
		return endPos, nil
	},
}

/*
Example of a template tag definition
{{ block "name" . }}
<p>{{ .Example.Name }}</p>
{{ end }}
*/
var TemplateBlockTag = &template.TemplateTag{
	Name:        "block",
	Description: "Defines a block of content that can be overridden in child templates.",
	Handler: func(engine *template.Engine, sb *strings.Builder, args []string, raw string, pos int) (newPos int, errs []error) {
		if len(args) < 1 {
			return pos, []error{fmt.Errorf("'block' tag requires a name argument")}
		}

		// The first argument is the block name
		blockName := strings.Trim(args[0], "\"'")

		// The second argument, if provided, is the context for the block
		var blockContext any
		var err error
		if len(args) > 1 {
			blockContext, err = engine.ResolveArguments(nil, args[1:])
			if err != nil {
				return pos, []error{err}
			}
		} else {
			// If no context provided, use the current data context
			blockContext = engine.GetLocalDataContext()
		}

		// Check if this block has been overridden
		existingBlock, exists := engine.GetBlock(blockName)
		if exists {
			// If the block exists, render it with the provided context
			switch content := existingBlock.(type) {
			case string:
				// Set up the context for rendering the block
				localContext := make(map[string]any)
				if mapContext, ok := blockContext.(map[string]any); ok {
					localContext = mapContext
				} else {
					localContext["."] = blockContext
				}

				result, renderErrs := engine.Render(content, localContext)
				if len(renderErrs) > 0 {
					errs = append(errs, renderErrs...)
				}
				sb.WriteString(result)
			default:
				errs = append(errs, fmt.Errorf("unexpected block content type: %T", existingBlock))
			}
		} else {
			// If the block doesn't exist, render the default content
			endPos, content, seekErrs := engine.SeekClosingTag(raw, "block", pos)
			if len(seekErrs) > 0 {
				return endPos, seekErrs
			}

			// Render the default content with the provided context
			localContext := make(map[string]any)
			if mapContext, ok := blockContext.(map[string]any); ok {
				localContext = mapContext
			} else {
				localContext["."] = blockContext
			}

			result, renderErrs := engine.Render(content, localContext)
			if len(renderErrs) > 0 {
				errs = append(errs, renderErrs...)
			}
			sb.WriteString(result)

			// Also store this default content for future use
			engine.AddBlock(blockName, content)

			return endPos, errs
		}

		// Skip past this block
		endPos, _, _ := engine.SeekClosingTag(raw, "block", pos)
		return endPos, errs
	},
}

/*
Example of a template tag definition
{{ template "name" . }}
--------
{{ template "name" .value }}
--------
{{ template "name" $value }}
This tag includes another template by name, passing the current context.
*/
var TemplateTemplateTag = &template.TemplateTag{
	Name:        "template",
	Description: "Includes another template by name.",
	Handler: func(engine *template.Engine, sb *strings.Builder, args []string, raw string, pos int) (newPos int, errs []error) {
		if len(args) < 1 {
			return pos, []error{fmt.Errorf("'template' tag requires a template name argument")}
		}

		// The first argument is the template name
		templateName := strings.Trim(args[0], "\"'")

		// The second argument, if provided, is the context for the template
		var templateContext map[string]any
		if len(args) > 1 {
			resolved, err := engine.ResolveArguments(nil, args[1:])
			if err != nil {
				return pos, []error{err}
			}

			// Convert the resolved value to a context map
			switch ctx := resolved.(type) {
			case map[string]any:
				templateContext = ctx
			default:
				// For other types, set as the "." value in the context
				templateContext = map[string]any{".": resolved}
			}
		} else {
			// If no context provided, use an empty context
			templateContext = make(map[string]any)
		}

		// Look up the template content from the engine's blocks
		templateContent, exists := engine.GetBlock(templateName)
		if !exists {
			return pos, []error{fmt.Errorf("template not found: %s", templateName)}
		}

		// Render the template with the context
		switch content := templateContent.(type) {
		case string:
			result, renderErrs := engine.Render(content, templateContext)
			if len(renderErrs) > 0 {
				errs = append(errs, renderErrs...)
			}
			sb.WriteString(result)
		default:
			errs = append(errs, fmt.Errorf("unexpected template content type: %T", templateContent))
		}

		return pos, errs
	},
}
