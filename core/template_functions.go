package core

import (
	"fmt"
	"reflect"
	"strings"
	"wispy-core/common"
	"wispy-core/models"
)

var IfTemplate = models.TemplateTag{
	Name: "if",
	Render: func(ctx TemplateCtx, sb *strings.Builder, parts []string, raw string, pos int) (newPos int, errs []error) {
		var tagName = "if"
		var args = parts[1:] // First part is the tag name, rest are arguments
		if len(args) < 1 {
			errs = append(errs, fmt.Errorf("if tag requires at least one argument"))
			return pos + 2, errs // advance position to avoid infinite loop
		}
		resolvedValue, resolvedValueType, resolvedErrs := ResolveFilterChain(args[0], ctx, DefaultFilters())
		if len(resolvedErrs) > 0 {
			errs = append(errs, resolvedErrs...)
		}

		var endTagLen = len(common.WrapBraces(" endif "))
		relativeEndTagPos, seekErrs := SeekEndTag(raw, pos, tagName)
		fmt.Println("seekErrs")
		fmt.Println(seekErrs)
		endTagPos := pos + relativeEndTagPos
		if len(seekErrs) > 0 {
			errs = append(errs, seekErrs...)
			return pos + 2, errs // advance position to avoid infinite loop
		} else if len(resolvedErrs) > 0 {
			return endTagPos + endTagLen, errs // advance position to avoid infinite loop
		}

		// Grab tag contents
		ifContent := raw[pos:endTagPos]
		// TODO: handle nested tags properly
		// TODO: handle nested else tag

		if resolvedValueType == nil {
			errs = append(errs, fmt.Errorf("resolved value type is nil for %s tag", tagName))
		} else {
			switch resolvedValueType.Kind() {
			case reflect.Bool:
				if resolvedValue.(bool) {
					// If the condition is true, render the content
					sb.WriteString(ifContent)
				}
			case reflect.String:
				if resolvedValue.(string) != "" {
					// If the string is not empty, render the content
					sb.WriteString(ifContent)
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if resolvedValue.(int) != 0 {
					// If the integer is not zero, render the content
					sb.WriteString(ifContent)
				}
			case reflect.Float32, reflect.Float64:
				if resolvedValue.(float64) != 0.0 {
					// If the float is not zero, render the content
					sb.WriteString(ifContent)
				}
			case reflect.Slice, reflect.Array, reflect.Map:
				if reflect.ValueOf(resolvedValue).Len() > 0 {
					// If the slice or array is not empty, render the content
					sb.WriteString(ifContent)
				}
			case reflect.Interface:
				if resolvedValue != nil {
					// If the interface is not nil, render the content
					sb.WriteString(ifContent)
				}
			default:
				errs = append(errs, fmt.Errorf("unsupported type %s for %s tag", resolvedValueType.Kind(), tagName))
			}
		}

		return endTagPos + endTagLen, errs // advance position to avoid infinite loop
	},
}

var DefineTag = models.TemplateTag{
	Name: "define",
	Render: func(ctx TemplateCtx, sb *strings.Builder, parts []string, raw string, pos int) (newPos int, errs []error) {
		var tagName = "define"
		if len(parts) < 2 {
			errs = append(errs, fmt.Errorf("define tag requires at least one argument"))
			return pos + 2, errs // advance position to avoid infinite loop
		}
		blockName := parts[1]
		if blockName == "" {
			errs = append(errs, fmt.Errorf("block name cannot be empty in %s tag", tagName))
			return pos + 2, errs // advance position to avoid infinite loop
		}

		var endTagLen = len(common.WrapBraces(" enddefine "))
		relativeEndTagPos, seekErrs := SeekEndTag(raw, pos, tagName)
		endTagPos := pos + relativeEndTagPos
		if len(seekErrs) > 0 {
			errs = append(errs, seekErrs...)
			return pos + 2, errs // advance position to avoid infinite loop
		}

		blockContent := raw[pos:endTagPos]
		ctx.InternalContext.Blocks[blockName] = blockContent

		return endTagPos + endTagLen, errs // advance position to avoid infinite loop
	},
}

var RenderTag = models.TemplateTag{
	Name: "render",
	Render: func(ctx TemplateCtx, sb *strings.Builder, parts []string, raw string, pos int) (newPos int, errs []error) {
		var tagName = "render"
		if len(parts) < 2 {
			errs = append(errs, fmt.Errorf("render tag requires at least one argument"))
			return pos + 2, errs // advance position to avoid infinite loop
		}
		blockName := parts[1]
		if blockName == "" {
			errs = append(errs, fmt.Errorf("block name cannot be empty in %s tag", tagName))
			return pos + 2, errs // advance position to avoid infinite loop
		}

		var endTagLen = len(common.WrapBraces(" endrender "))
		blockContent, exists := ctx.InternalContext.Blocks[blockName]
		// If the block does not exist, we try to find an end tag for the block
		if !exists {
			relativeEndTagPos, seekErrs := SeekEndTag(raw, pos, tagName)
			endTagPos := pos + relativeEndTagPos
			if len(seekErrs) > 0 {
				errs = append(errs, seekErrs...)
				return pos + 2, errs // advance position to avoid infinite loop
			}

			fallbackContent := raw[pos:endTagPos]
			renderedContent, renderErrs := ctx.Engine.Render(fallbackContent, ctx)
			if len(renderErrs) > 0 {
				errs = append(errs, renderErrs...)
			}
			sb.WriteString(renderedContent)
			return endTagPos + endTagLen, errs // advance position to avoid infinite loop
		}

		// Render the block content
		// TODO: add support for passing arguments to the block
		// For now, we just render the block content with the current context
		renderedContent, renderErrs := ctx.Engine.Render(blockContent, ctx)
		if len(renderErrs) > 0 {
			errs = append(errs, renderErrs...)
		}
		sb.WriteString(renderedContent)
		return pos + endTagLen, errs // advance position to avoid infinite loop
	},
}

var ForTag = models.TemplateTag{
	Name: "for",
	Render: func(ctx TemplateCtx, sb *strings.Builder, parts []string, raw string, pos int) (newPos int, errs []error) {
		var tagName = "for"
		var endTagLen = len(common.WrapBraces(" endfor "))
		if len(parts) < 3 {
			errs = append(errs, fmt.Errorf("for tag requires at least two arguments \"foo in .bars\""))
			return pos + 2, errs // advance position to avoid infinite loop
		}
		itemName := parts[1]
		if itemName == "" {
			errs = append(errs, fmt.Errorf("collection name cannot be empty in %s tag", tagName))
			return pos + 2, errs // advance position to avoid infinite loop
		}
		if parts[2] != "in" {
			errs = append(errs, fmt.Errorf("expected 'in' after item name in %s tag, got '%s'", tagName, parts[2]))
			return pos + 2, errs // advance position to avoid infinite loop
		}
		collectionName := parts[3]
		if collectionName == "" {
			errs = append(errs, fmt.Errorf("collection name cannot be empty in %s tag", tagName))
			return pos + 2, errs // advance position to avoid infinite loop
		}

		relativeEndTagPos, seekErrs := SeekEndTag(raw, pos, tagName)
		endTagPos := pos + relativeEndTagPos
		if len(seekErrs) > 0 {
			errs = append(errs, seekErrs...)
			return pos + 2, errs // advance position to avoid infinite loop
		}

		// Grab tag contents
		forContent := raw[pos:endTagPos]
		// Resolve the collection
		resolvedValue, resolvedValueType, resolvedErrs := ResolveFilterChain(collectionName, ctx, DefaultFilters())
		if len(resolvedErrs) > 0 {
			errs = append(errs, resolvedErrs...)
		}

		switch resolvedValueType.Kind() {
		case reflect.Slice, reflect.Array:
			// If the resolved value is a slice or array, iterate over it
			val := reflect.ValueOf(resolvedValue)
			for i := 0; i < val.Len(); i++ {
				itemValue := val.Index(i).Interface()
				// Create a new context for each item
				var newData = map[string]interface{}{}
				newData[itemName] = itemValue
				// Render the content with the new context
				itemContent, itemErrs := ctx.Engine.Render(forContent, ctx.Engine.CloneCtx(ctx, newData))
				if len(itemErrs) > 0 {
					errs = append(errs, itemErrs...)
					continue // skip this iteration if there are errors
				}
				sb.WriteString(itemContent)
			}
		case reflect.Map:
			// If the resolved value is a map, iterate over it
			val := reflect.ValueOf(resolvedValue)
			for _, key := range val.MapKeys() {
				itemValue := val.MapIndex(key).Interface()
				// Create a new context for each item
				var newData = map[string]interface{}{}
				newData[itemName] = itemValue
				// Render the content with the new context
				itemContent, itemErrs := ctx.Engine.Render(forContent, ctx.Engine.CloneCtx(ctx, newData))
				if len(itemErrs) > 0 {
					errs = append(errs, itemErrs...)
					continue // skip this iteration if there are errors
				}
				sb.WriteString(itemContent)
			}
		case reflect.String:
			// If the resolved value is a string, treat it as a slice of runes
			val := reflect.ValueOf(resolvedValue)
			for i := 0; i < val.Len(); i++ {
				itemValue := string(val.Index(i).Interface().(rune))
				// Create a new context for each item
				var newData = map[string]interface{}{}
				newData[itemName] = itemValue
				// Render the content with the new context
				itemContent, itemErrs := ctx.Engine.Render(forContent, ctx.Engine.CloneCtx(ctx, newData))
				if len(itemErrs) > 0 {
					errs = append(errs, itemErrs...)
					continue // skip this iteration if there are errors
				}
				sb.WriteString(itemContent)
			}

		}

		return endTagPos + endTagLen, errs // advance position to avoid infinite loop
	},
}

// DefaultFunctionMap provides all built-in tags.
func DefaultFunctionMap() models.FunctionMap {
	return models.FunctionMap{
		"if":     IfTemplate,
		"for":    ForTag,
		"define": DefineTag,
		"render": RenderTag,
	}
}
