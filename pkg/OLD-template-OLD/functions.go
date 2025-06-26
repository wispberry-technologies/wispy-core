// Package template provides functions for template parsing and rendering
package template

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"wispy-core/pkg/common"
	"wispy-core/pkg/models"
)

// TemplateCtx is an alias for template context
type TemplateCtx = *models.TemplateContext

// IfTemplate handles conditional rendering
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
		condition := strings.TrimSpace(strings.Join(parts[1:], " "))

		// Find the matching endif
		endPos, seekErrs := SeekEndTag(raw, pos, "if")
		if len(seekErrs) > 0 {
			errs = append(errs, seekErrs...)
			return pos, errs
		}

		// Extract content between if and endif - need to find the start of content
		// pos is currently at the end of the opening tag
		contentStart := pos
		contentEnd := endPos
		content := raw[contentStart : contentEnd-len("{% endif %}")]

		// Check if there's an else tag inside the content
		elsePos := strings.Index(content, "{% else %}")
		var ifContent, elseContent string

		if elsePos >= 0 {
			// Split the content into if and else parts
			ifContent = content[:elsePos]
			// add offsets to exclude wrapping tags from rendering
			elseContent = content[elsePos+len("{% else %}"):]
		} else {
			// No else found, all content belongs to if part
			ifContent = content
		}

		// Resolve the condition value - be graceful with nil values
		value, _, resolveErrs := ResolveFilterChain(condition, ctx, ctx.Engine.FilterMap)
		if len(resolveErrs) > 0 {
			// Only log filter errors, not unresolved variable errors
			errs = append(errs, resolveErrs...)
		}

		// Render content based on condition
		boolCondition := IsTruthy(value)
		if boolCondition {
			// Condition is true, render the if part
			rendered, renderErrs := ctx.Engine.Render(ifContent, ctx)
			if len(renderErrs) > 0 {
				errs = append(errs, renderErrs...)
			}
			sb.WriteString(rendered)
		} else if elsePos >= 0 {
			// Condition is false and there's an else part, render the else part
			rendered, renderErrs := ctx.Engine.Render(elseContent, ctx)
			if len(renderErrs) > 0 {
				errs = append(errs, renderErrs...)
			}
			sb.WriteString(rendered)
		}

		return endPos, errs
	},
}

// ForTag handles iteration over collections
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
		refVal := reflect.ValueOf(collection)
		switch refVal.Kind() {
		case reflect.Slice, reflect.Array:
			// Handle slice or array iteration
			for i := 0; i < refVal.Len(); i++ {
				// Create a nested context with loop variable
				nestedCtx := ctx.Engine.CloneCtx(ctx, map[string]interface{}{
					itemVar:          refVal.Index(i).Interface(),
					"loop_index":     i,
					"loop_index1":    i + 1,
					"loop_first":     i == 0,
					"loop_last":      i == refVal.Len()-1,
					"loop_length":    refVal.Len(),
					"loop_remaining": refVal.Len() - i - 1,
				})

				// Render content with nested context
				rendered, renderErrs := ctx.Engine.Render(content, nestedCtx)
				if len(renderErrs) > 0 {
					errs = append(errs, renderErrs...)
				}
				sb.WriteString(rendered)
			}
		case reflect.Map:
			// Handle map iteration
			mapKeys := refVal.MapKeys()
			for i, key := range mapKeys {
				// Create a nested context with loop variable and key
				value := refVal.MapIndex(key)
				nestedCtx := ctx.Engine.CloneCtx(ctx, map[string]interface{}{
					itemVar:          value.Interface(),
					itemVar + "_key": key.Interface(),
					"loop_index":     i,
					"loop_index1":    i + 1,
					"loop_first":     i == 0,
					"loop_last":      i == len(mapKeys)-1,
					"loop_length":    len(mapKeys),
					"loop_remaining": len(mapKeys) - i - 1,
				})

				// Render content with nested context
				rendered, renderErrs := ctx.Engine.Render(content, nestedCtx)
				if len(renderErrs) > 0 {
					errs = append(errs, renderErrs...)
				}
				sb.WriteString(rendered)
			}
		default:
			// Not an iterable type
			errs = append(errs, fmt.Errorf("for tag requires an iterable collection, got %T", collection))
		}

		return endPos, errs
	},
}

// IncludeTag handles including another template
var IncludeTag = models.TemplateTag{
	Name: "include",
	Render: func(ctx TemplateCtx, sb *strings.Builder, parts []string, raw string, pos int) (newPos int, errs []error) {
		if len(parts) < 2 {
			return pos, []error{fmt.Errorf("include tag requires a template name")}
		}

		// Get template name (might be a variable or quoted string)
		templateName := parts[1]
		// Get with parameter index if present
		withIndex := -1
		for i, part := range parts {
			if part == "with" {
				withIndex = i
				break
			}
		}

		// Resolve template name if it's a variable
		if !common.IsQuotedString(templateName) {
			resolvedName, _, resolveErrs := ResolveFilterChain(templateName, ctx, ctx.Engine.FilterMap)
			if len(resolveErrs) > 0 {
				errs = append(errs, resolveErrs...)
			}
			if resolvedName != nil {
				templateName = fmt.Sprintf("%v", resolvedName)
			} else {
				// If template name resolves to nil, treat as error
				errs = append(errs, fmt.Errorf("include template name resolved to nil"))
				return pos, errs
			}
		} else {
			// Remove quotes if it's a quoted string
			templateName = templateName[1 : len(templateName)-1]
		}

		// Process with parameters if present
		var withData map[string]interface{}
		if withIndex > 0 && withIndex < len(parts)-1 {
			withData = make(map[string]interface{})
			for i := withIndex + 1; i < len(parts); i++ {
				// Parse key=value pairs
				keyValue := strings.SplitN(parts[i], "=", 2)
				if len(keyValue) == 2 {
					key := keyValue[0]
					valueStr := keyValue[1]

					// Resolve value
					value := ResolveValue(valueStr, ctx)
					withData[key] = value
				}
			}
		}

		// Look up the template in cache first
		var templateContent string
		var ok bool

		// Check if it's already in the templates cache
		if ctx.InternalContext.TemplatesCache != nil {
			templateContent, ok = ctx.InternalContext.TemplatesCache[templateName]
		}

		if !ok {
			// Not in cache, load from disk
			templatePath, err := resolveTemplatePath(templateName, ctx)
			if err != nil {
				errs = append(errs, err)
			} else {
				templateContent, err = LoadTemplate(templatePath)
				if err != nil {
					errs = append(errs, err)
				}
			}
		}

		// Create nested context for the included template
		nestedCtx := ctx
		if withData != nil {
			nestedCtx = ctx.Engine.CloneCtx(ctx, withData)
		}

		// Render the included template
		rendered, renderErrs := ctx.Engine.Render(templateContent, nestedCtx)
		if len(renderErrs) > 0 {
			errs = append(errs, renderErrs...)
		}
		sb.WriteString(rendered)

		return pos, errs
	},
}

// RenderTag is an alias for IncludeTag for compatibility
var RenderTag = models.TemplateTag{
	Name:   "render",
	Render: IncludeTag.Render, // Use the same function as include
}

// BlockTag handles block definitions
var BlockTag = models.TemplateTag{
	Name: "block",
	Render: func(ctx TemplateCtx, sb *strings.Builder, parts []string, raw string, pos int) (newPos int, errs []error) {
		if len(parts) < 2 {
			return pos, []error{fmt.Errorf("block tag requires a name")}
		}

		blockName := parts[1]
		if common.IsQuotedString(blockName) {
			blockName = blockName[1 : len(blockName)-1]
		}

		// Find the matching endblock
		endPos, seekErrs := SeekEndTag(raw, pos, "block")
		if len(seekErrs) > 0 {
			return pos, seekErrs
		}

		// Extract content between block and endblock
		contentStart := pos
		contentEnd := endPos - len("{% endblock %}")
		content := raw[contentStart:contentEnd]

		if ctx.InternalContext.Blocks != nil {
			if overrideContent, ok := ctx.InternalContext.Blocks[blockName]; ok {
				// Use the override content
				content = overrideContent
			}
		}

		// Store this block content in case it's referenced later
		if ctx.InternalContext.Blocks == nil {
			ctx.InternalContext.Blocks = make(map[string]string)
		}
		ctx.InternalContext.Blocks[blockName] = content

		// Render the block content
		rendered, renderErrs := ctx.Engine.Render(content, ctx)
		if len(renderErrs) > 0 {
			errs = append(errs, renderErrs...)
		}
		sb.WriteString(rendered)

		return endPos, errs
	},
}

// DefineTag handles block definitions
var DefineTag = models.TemplateTag{
	Name: "define",
	Render: func(ctx TemplateCtx, sb *strings.Builder, parts []string, raw string, pos int) (newPos int, errs []error) {
		if len(parts) < 2 {
			return pos, []error{fmt.Errorf("define tag requires a block name")}
		}

		// Get block name (remove quotes if present)
		blockName := parts[1]
		if common.IsQuotedString(blockName) {
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

		// Store the block content in the context
		if ctx.InternalContext.Blocks == nil {
			ctx.InternalContext.Blocks = make(map[string]string)
		}
		ctx.InternalContext.Blocks[blockName] = content

		// For define tags, we don't render anything to the output - just store the block
		return endPos, errs
	},
}

// ExtendTag handles template inheritance
var ExtendTag = models.TemplateTag{
	Name: "extend",
	Render: func(ctx TemplateCtx, sb *strings.Builder, parts []string, raw string, pos int) (newPos int, errs []error) {
		if len(parts) < 2 {
			return pos, []error{fmt.Errorf("extend tag requires a template name")}
		}

		// Get parent template name
		parentName := parts[1]
		if common.IsQuotedString(parentName) {
			parentName = parentName[1 : len(parentName)-1]
		} else {
			// Try to resolve if it's a variable
			resolved, _, resolveErrs := ResolveFilterChain(parentName, ctx, ctx.Engine.FilterMap)
			if len(resolveErrs) > 0 {
				errs = append(errs, resolveErrs...)
			}
			if resolved != nil {
				parentName = fmt.Sprintf("%v", resolved)
			}
		}

		// Process current template to find all block definitions
		// This will populate ctx.InternalContext.Blocks with all blocks from the current template
		_, processErrs := ctx.Engine.Render(raw, ctx)
		if len(processErrs) > 0 {
			errs = append(errs, processErrs...)
		}

		// Look up the parent template
		var parentContent string
		var ok bool

		// Check if parent is in template cache
		if ctx.InternalContext.TemplatesCache != nil {
			parentContent, ok = ctx.InternalContext.TemplatesCache[parentName]
		}

		if !ok {
			// Not in cache, should handle loading from disk
			// This is a placeholder - real implementation would need proper template loading
			parentContent = "Parent template not found: " + parentName
			errs = append(errs, fmt.Errorf("parent template not found: %s", parentName))
		}

		// Now render the parent template with the current blocks
		rendered, parentErrs := ctx.Engine.Render(parentContent, ctx)
		if len(parentErrs) > 0 {
			errs = append(errs, parentErrs...)
		}

		sb.WriteString(rendered)

		// Skip the rest of the current template since it's just block definitions
		// that have already been processed
		return len(raw), errs
	},
}

var VerbatimTag = models.TemplateTag{
	Name: "verbatim",
	Render: func(ctx TemplateCtx, sb *strings.Builder, parts []string, raw string, pos int) (newPos int, errs []error) {
		// Find the matching endverbatim
		endPos, seekErrs := SeekEndTag(raw, pos, "verbatim")
		if len(seekErrs) > 0 {
			errs = append(errs, seekErrs...)
			return pos, errs
		}

		// Extract content between verbatim and endverbatim - need to find the start of content
		// pos is currently at the end of the opening tag
		contentStart := pos
		contentEnd := endPos - len("{% endverbatim %}")
		content := raw[contentStart:contentEnd]

		// Output the content literally (without template processing)
		sb.WriteString(content)

		return endPos, errs
	},
}

var IconTag = models.TemplateTag{
	Name: "icon",
	Render: func(ctx TemplateCtx, sb *strings.Builder, parts []string, raw string, pos int) (newPos int, errs []error) {
		if len(parts) < 2 {
			errs = append(errs, fmt.Errorf("icon tag requires an icon name"))
			return pos, errs
		}

		// read icon svg from ./data/static/icons/*
		iconName := parts[1]
		if common.IsQuotedString(iconName) {
			iconName = iconName[1 : len(iconName)-1]
		}
		iconPath := filepath.Join(common.GetEnv("WISPY_ROOT", "./"), "data", "static", "icons", iconName+".svg")
		content, readErr := os.ReadFile(iconPath)

		if len(parts) >= 3 && strings.HasPrefix(parts[2], "class=") {
			classStr := parts[2]
			content = []byte(strings.Replace(string(content), "fill=\"none\"", "fill=\"none\" "+classStr, -1))
		}

		if readErr != nil {
			errs = append(errs, fmt.Errorf("failed to read icon file %s: %v", iconPath, readErr))
			return pos, errs
		}

		sb.WriteString(string(content))
		return pos, errs
	},
}

var AssetTag = models.TemplateTag{
	Name: "asset",
	Render: func(ctx TemplateCtx, sb *strings.Builder, parts []string, raw string, pos int) (newPos int, errs []error) {
		if len(parts) < 3 {
			errs = append(errs, fmt.Errorf("asset tag requires asset type and file path"))
			return pos, errs
		}

		// Parse arguments: asset "css" "path/to/file.css" args...
		args := common.FieldsRespectQuotes(strings.Join(parts[1:], " "))
		if len(args) < 2 {
			errs = append(errs, fmt.Errorf("asset tag requires asset type and file path"))
			return pos, errs
		}

		assetType := strings.Trim(args[0], `"'`)
		filePath := strings.Trim(args[1], `"'`)

		// Validate asset type
		validTypes := []string{"css", "css-inline", "js", "js-inline"}
		isValidType := false
		for _, validType := range validTypes {
			if assetType == validType {
				isValidType = true
				break
			}
		}
		if !isValidType {
			errs = append(errs, fmt.Errorf("invalid asset type '%s', must be one of: %s", assetType, strings.Join(validTypes, ", ")))
			return pos, errs
		}

		isInline := strings.HasSuffix(assetType, "-inline")
		isCSS := strings.HasPrefix(assetType, "css")
		location := "head" // Default location

		// Parse additional parameters for JS
		if !isCSS {
			for _, arg := range args[2:] {
				if strings.HasPrefix(arg, "location=") {
					location = strings.Trim(strings.TrimPrefix(arg, "location="), `"'`)
				}
			}
		}

		// Validate asset path
		validatedPath, pathErr := validateAssetPath(filePath, isInline)
		if pathErr != nil {
			// Log error but continue rendering - graceful error handling
			errs = append(errs, fmt.Errorf("asset validation failed: %v", pathErr))
			return pos, errs // Skip this asset but continue rendering at current position
		}

		// Create unique key for deduplication
		importType := assetType
		dedupeKey := validatedPath + "|" + importType

		// Check for conflicts (same file with different import types)
		for existingKey, existingType := range ctx.InternalContext.ImportedResources {
			existingPath := strings.Split(existingKey, "|")[0]
			if existingPath == validatedPath && existingType != importType {
				errs = append(errs, fmt.Errorf("asset %s already imported as %s, cannot import as %s", validatedPath, existingType, importType))
				return pos, errs // Skip this asset but continue rendering at current position
			}
		}

		// Check for exact duplicates
		if _, exists := ctx.InternalContext.ImportedResources[dedupeKey]; exists {
			// Same file with same import type - skip silently
			return pos, errs
		}

		// Mark as imported
		ctx.InternalContext.ImportedResources[dedupeKey] = importType

		if isInline {
			// Handle inline assets
			if isRemoteURL(validatedPath) {
				errs = append(errs, fmt.Errorf("cannot inline remote assets"))
				return pos, errs // Skip this asset but continue rendering
			}

			// Resolve and read file content
			fullPath, resolveErr := resolveAssetPath(validatedPath, ctx)
			if resolveErr != nil {
				// Log error but continue rendering - graceful error handling
				errs = append(errs, fmt.Errorf("%v", resolveErr))
				return pos, errs // Skip this asset but continue rendering
			}

			content, readErr := os.ReadFile(fullPath)
			if readErr != nil {
				// Log error but continue rendering - graceful error handling
				errs = append(errs, fmt.Errorf("failed to read asset file %s: %v", validatedPath, readErr))
				return pos, errs // Skip this asset but continue rendering
			}

			if isCSS {
				ctx.InternalContext.HtmlDocumentTags = append(ctx.InternalContext.HtmlDocumentTags, models.HtmlDocumentTags{
					TagType:     "style",
					TagName:     "style",
					Location:    "head",
					Contents:    string(content),
					Priority:    20,
					Attributes:  map[string]string{"type": "text/css"},
					SelfClosing: false,
				})
			} else {
				priority := 25
				if location == "pre-footer" {
					priority = 30
				}

				ctx.InternalContext.HtmlDocumentTags = append(ctx.InternalContext.HtmlDocumentTags, models.HtmlDocumentTags{
					TagType:     "script",
					TagName:     "script",
					Location:    location,
					Contents:    string(content),
					Priority:    priority,
					Attributes:  map[string]string{"type": "text/javascript"},
					SelfClosing: false,
				})
			}
		} else {
			// Handle external assets
			if isCSS {
				// For remote URLs, use as-is; for local files, convert to web path
				webPath := validatedPath
				if !isRemoteURL(validatedPath) {
					// Convert local path to web path (remove leading slash if present, then add it back)
					webPath = "/" + strings.TrimPrefix(validatedPath, "/")
				}

				ctx.InternalContext.HtmlDocumentTags = append(ctx.InternalContext.HtmlDocumentTags, models.HtmlDocumentTags{
					TagType:  "link",
					TagName:  "link",
					Location: "head",
					Contents: "",
					Priority: 15,
					Attributes: map[string]string{
						"rel":  "stylesheet",
						"href": webPath,
						"type": "text/css",
					},
					SelfClosing: true,
				})
			} else {
				// For remote URLs, use as-is; for local files, convert to web path
				webPath := validatedPath
				if !isRemoteURL(validatedPath) {
					// Convert local path to web path
					webPath = "/" + strings.TrimPrefix(validatedPath, "/")
				}

				priority := 20
				if location == "pre-footer" {
					priority = 25
				}

				ctx.InternalContext.HtmlDocumentTags = append(ctx.InternalContext.HtmlDocumentTags, models.HtmlDocumentTags{
					TagType:  "script",
					TagName:  "script",
					Location: location,
					Contents: "",
					Priority: priority,
					Attributes: map[string]string{
						"src":  webPath,
						"type": "text/javascript",
					},
					SelfClosing: false,
				})
			}
		}

		return pos, errs
	},
}
