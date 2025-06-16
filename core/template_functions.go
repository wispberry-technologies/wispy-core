package core

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"wispy-core/common"
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

		templateName := strings.Trim(parts[1], `"'`)
		if newName, ok := strings.CutPrefix(templateName, "@app/"); ok {
			templateName = "app/" + newName
		} else if newName, ok := strings.CutPrefix(templateName, "@marketing/"); ok {
			templateName = "marketing/" + newName
		} else {
			errs = append(errs, fmt.Errorf("render tag requires a valid template name starting with @app/ or @marketing/"))
			return pos, errs
		}

		absolutePath, _ := filepath.Abs("./templates")
		if !strings.HasSuffix(templateName, ".html") {
			templateName += ".html" // Ensure it has .html extension
		}
		templateContent, err := os.ReadFile(filepath.Join(absolutePath, templateName))
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read template file: %w", err))
			return pos, errs
		}

		// Render the included template with current context
		rendered, renderErrs := ctx.Engine.Render(string(templateContent), ctx)
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
			fullPath, resolveErr := resolveAssetPath(ctx, validatedPath)
			if resolveErr != nil {
				// Log error but continue rendering - graceful error handling
				errs = append(errs, fmt.Errorf("failed to resolve asset path: %v", resolveErr))
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

// getSitePath returns the root path for the current site based on the page context
func getSitePath(ctx TemplateCtx) string {
	// Try to get domain from Page context first
	if pageData, exists := ctx.Data["Page"]; exists {
		if page, ok := pageData.(*models.Page); ok {
			// In test environment, use current working directory
			if os.Getenv("WISPY_CORE_ROOT") == "" {
				if wd, err := os.Getwd(); err == nil {
					// If we're in the core directory (during tests), go up one level
					if strings.HasSuffix(wd, "/core") {
						wd = filepath.Dir(wd)
					}
					return filepath.Join(wd, "sites", page.SiteDetails.Domain)
				}
				return filepath.Join("sites", page.SiteDetails.Domain)
			}
			return common.RootSitesPath(page.SiteDetails.Domain)
		}
	}

	// Fallback to request host if available
	if ctx.Request != nil {
		domain := ctx.Request.Host
		// In test environment, use current working directory
		if os.Getenv("WISPY_CORE_ROOT") == "" {
			if wd, err := os.Getwd(); err == nil {
				// If we're in the core directory (during tests), go up one level
				if strings.HasSuffix(wd, "/core") {
					wd = filepath.Dir(wd)
				}
				return filepath.Join(wd, "sites", domain)
			}
			return filepath.Join("sites", domain)
		}
		return common.RootSitesPath(domain)
	}

	return ""
}

// isRemoteURL checks if a path is a remote URL (starts with https://)
func isRemoteURL(path string) bool {
	return strings.HasPrefix(strings.ToLower(path), "https://")
}

// validateAssetPath validates and normalizes asset paths for security
func validateAssetPath(path string, inline bool) (string, error) {
	if isRemoteURL(path) {
		if inline {
			return "", fmt.Errorf("cannot inline remote assets, remote URLs are only allowed for external linking")
		}
		// Validate that it's a proper URL
		if _, err := url.Parse(path); err != nil {
			return "", fmt.Errorf("invalid remote URL: %v", err)
		}
		return path, nil
	}

	// For local files, ensure they start with allowed prefixes
	allowedPrefixes := []string{"assets/", "public/", "/assets/", "/public/"}
	hasValidPrefix := false
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(path, prefix) {
			hasValidPrefix = true
			break
		}
	}

	if !hasValidPrefix {
		return "", fmt.Errorf("asset path must start with 'assets/', 'public/', '/assets/', or '/public/', or be a remote HTTPS URL")
	}

	// Remove leading slash if present for consistent path handling
	path = strings.TrimPrefix(path, "/")

	return path, nil
}

// resolveAssetPath resolves the full file path for a local asset
func resolveAssetPath(ctx TemplateCtx, path string) (string, error) {
	if isRemoteURL(path) {
		return path, nil // Remote URLs are returned as-is
	}

	sitePath := getSitePath(ctx)
	if sitePath == "" {
		return "", fmt.Errorf("unable to determine site path")
	}

	fullPath := filepath.Join(sitePath, path)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("asset file not found: %s", path)
	}

	return fullPath, nil
}

// DefaultFunctionMap provides all built-in tags.
func DefaultFunctionMap() models.FunctionMap {
	return models.FunctionMap{
		"if":       IfTemplate,
		"for":      ForTag,
		"define":   DefineTag,
		"render":   RenderTag,
		"block":    BlockTag,
		"verbatim": VerbatimTag,
		"asset":    AssetTag,
	}
}
