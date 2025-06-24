package template

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"wispy-core/internal/core/parser"
	"wispy-core/pkg/common"
	"wispy-core/pkg/models"

	"github.com/microcosm-cc/bluemonday"
)

var policy *bluemonday.Policy

func init() {
	// Initialize the HTML sanitizer
	policy = bluemonday.UGCPolicy()
}

// ResolveFilterChain resolves a filter chain string and applies filters
// Returns the resolved value, its type, and any errors encountered
func ResolveFilterChain(filterChainString string, ctx TemplateCtx, filters models.FilterMap) (value interface{}, valueType reflect.Type, errors []error) {
	if len(filterChainString) == 0 {
		// Empty filter chain is not an error, just return nil
		return nil, nil, nil
	}

	// Trim whitespace from the entire string first
	filterChainString = strings.TrimSpace(filterChainString)
	splitFilters := strings.Split(filterChainString, "|")
	common.Debug("--- Resolving filter:", "filters", splitFilters)

	value = ResolveValue(strings.TrimSpace(splitFilters[0]), ctx)
	common.Debug("--- Initial value resolved", "value", value)

	if len(splitFilters) == 1 {
		// No filters, just a single value
		// Don't treat nil values as errors - they should just render as empty
		if value == nil {
			return nil, nil, nil // No error, just nil value
		}
		return value, reflect.TypeOf(value), errors
	} else {
		// Multiple filters found, process them
		if value == nil {
			// If initial value is nil, don't apply filters, just return nil
			return nil, nil, nil
		}
		valueType = reflect.TypeOf(value)
		for _, filterExpr := range splitFilters[1:] {
			filterExpr = strings.TrimSpace(filterExpr)
			filterName, args := ParseFilterExpression(filterExpr)
			if filter, ok := filters[filterName]; ok {
				oldValue := value
				value = filter(value, valueType, args, ctx)
				valueType = reflect.TypeOf(value)
				common.Debug("--- Filter applied", "filterName", filterName, "oldValue", oldValue, "newValue", value)
			} else {
				common.Debug("--- Unknown filter", "filterName", filterName)
				errors = append(errors, fmt.Errorf("unknown filter: %s", filterName))
			}
		}
	}
	return value, valueType, errors
}

// ParseFilterExpression parses a filter expression into name and arguments
func ParseFilterExpression(filterExpr string) (string, []string) {
	// Split the filter expression into name and arguments
	parts := strings.SplitN(filterExpr, "(", 2)
	if len(parts) == 1 {
		return strings.TrimPrefix(parts[0], "("), nil // No arguments
	}
	name := parts[0]
	args := common.FieldsRespectQuotes(strings.TrimSuffix(parts[1], ")")) // Remove trailing parenthesis
	return name, args
}

// ResolveValue resolves a value identifier to its actual value
// Supports literal strings (quoted) and context variables (with dot notation)
func ResolveValue(valueIdentifier string, ctx TemplateCtx) interface{} {
	// Handle empty identifier
	if len(valueIdentifier) == 0 {
		return nil
	}

	// Check for literal string values (surrounded by quotes)
	if common.IsQuotedString(valueIdentifier) {
		// Remove the surrounding quotes to get the literal value
		result := valueIdentifier[1 : len(valueIdentifier)-1]
		return result
	}

	// Try to resolve from context data (handles dot notation for nested access)
	if ctx != nil && ctx.Data != nil {
		result := ResolveDotNotation(ctx, valueIdentifier)
		return result
	}

	common.Debug("No context or data available", "valueIdentifier", valueIdentifier)
	return nil
}

// ResolveDotNotation resolves dot notation (e.g., "user.name") in a context map
func ResolveDotNotation(ctx TemplateCtx, key string) interface{} {
	if key == "" {
		return nil
	}

	// Remove leading dot if present (e.g., ".Site" becomes "Site")
	key = strings.TrimPrefix(key, ".")

	parts := strings.Split(key, ".")
	var current interface{} = map[string]interface{}{}
	switch parts[0] {
	case "Page":
		current = ctx.Page.PageData
	case "Site":
		current = ctx.Page.SiteDetails
	default:
		current = ctx.Data
	}

	for _, part := range parts {
		if current == nil {
			return nil
		}

		switch val := current.(type) {
		case map[string]interface{}:
			current = val[part]

		default:
			// Handle struct and pointer types
			v := reflect.ValueOf(current)

			// Handle pointer types
			if v.Kind() == reflect.Ptr {
				if v.IsNil() {
					return nil
				}
				v = v.Elem()
			}

			// Handle structs
			if v.Kind() == reflect.Struct {
				field := v.FieldByName(part)
				if !field.IsValid() {
					common.Debug("Field not found ", "struct", v.Type().Name(), "field", part)
					return nil
				}
				current = field.Interface()
			} else {
				common.Debug("Unsupported type ", "type", v.Kind())
				return nil
			}
		}
	}

	// Sanitize string values for security
	if s, ok := current.(string); ok {
		return policy.Sanitize(s)
	}

	return current
}

// --------------------
// Template Functions
// --------------------

// SeekClosingHandleNested finds the index of a closing tag while ensuring it corresponds to an opening tag
func SeekClosingHandleNested(raw, closingTag, openingTag string, pos int) (newPos int, separatorLength int) {
	openCount := 0
	newPos = pos
	separatorLength = len(closingTag)

	for {
		closeIndex := strings.Index(raw[newPos:], closingTag)
		if closeIndex == -1 {
			// No more closing tags found, return -1
			return newPos, separatorLength
		}
		closeIndex += newPos

		// Only search for an opening tag within the range before this closing tag
		openIndex := strings.Index(raw[newPos:closeIndex], openingTag)
		if openIndex != -1 {
			openCount++
			newPos += openIndex + len(openingTag)
			continue
		}

		// If no unmatched opening tags remain, return this closing tag index
		if openCount == 0 {
			return closeIndex, separatorLength
		}

		// Otherwise, decrement open count and continue searching
		openCount--
		newPos = closeIndex + separatorLength
	}
}

// SeekEndTag finds the matching end tag for a given tag, handling nesting properly
func SeekEndTag(raw string, pos int, tagName string) (endTagPos int, errs []error) {
	openTag := fmt.Sprintf("{%% %s ", tagName)
	closeTag := fmt.Sprintf("{%% end%s %%}", tagName)

	endPos, endTagLen := SeekClosingHandleNested(raw, closeTag, openTag, pos)
	if endPos == -1 {
		return pos, []error{fmt.Errorf("unclosed tag %q at position %d", tagName, pos)}
	}

	return endPos + endTagLen, nil
}

// LoadTemplate loads a template from disk
func LoadTemplate(templatePath string) (string, error) {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to load template %s: %w", templatePath, err)
	}
	return string(content), nil
}

// IsTruthy determines if a value is considered "truthy" in template logic
func IsTruthy(value interface{}) bool {
	common.Info("value=%s", value)
	if value == nil {
		return false
	}

	// Get reflection value first
	rv := reflect.ValueOf(value)
	// Handle different kinds of values
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.String:
		str := strings.Trim(rv.String(), " \t\n\r\"'")
		// Handle literal "true"/"false" after getting the string value
		if str == "true" {
			return true
		}
		if str == "false" {
			return false
		}
		return str != ""
	case reflect.Array, reflect.Slice, reflect.Map:
		return rv.Len() > 0
	case reflect.Struct:
		return true // Non-nil structs are always truthy
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() != 0
	default:
		return false // For any other type, default to false
	}
}

func GetLayout(domain, name string) string {
	layoutPath := common.RootSitesPath(domain, "layouts", name+".html")
	if _, err := os.Stat(layoutPath); os.IsNotExist(err) {
		return "<!-- Error no layout found: " + err.Error() + "-->"
	}

	bytes, err := os.ReadFile(layoutPath)
	if err != nil {
		return "<!-- Error reading layout file: " + err.Error() + "-->"
	}

	return parser.RemoveMetadataFromContent(string(bytes))
}

func GetPage(domain, filePath string) string {
	pagePath := common.RootSitesPath(domain, "pages", filePath)
	if _, err := os.Stat(pagePath); os.IsNotExist(err) {
		return "<!-- Error no page found: " + err.Error() + "-->"
	}

	bytes, err := os.ReadFile(pagePath)
	if err != nil {
		return "<!-- Error reading page file: " + err.Error() + "-->"
	}

	// remove page metadata
	return parser.RemoveMetadataFromContent(string(bytes))
}

// resolveTemplatePath resolves template paths like "@app/account-login.html"
func resolveTemplatePath(templateName string, ctx TemplateCtx) (string, error) {
	// Handle @namespace/template.html format
	if strings.HasPrefix(templateName, "@") {
		parts := strings.SplitN(templateName[1:], "/", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid template path format: %s", templateName)
		}

		namespace := parts[0]
		filename := parts[1]

		// Map namespaces to actual directories
		// Global templates are in data/templates/, not data/sites/templates/
		var basePath string
		switch namespace {
		case "app":
			// Use WISPY_CORE_ROOT + data/templates/app instead of sites path
			coreRoot := common.MustGetEnv("WISPY_CORE_ROOT")
			basePath = filepath.Join(coreRoot, "data", "templates", "app")
		case "cms":
			coreRoot := common.MustGetEnv("WISPY_CORE_ROOT")
			basePath = filepath.Join(coreRoot, "data", "templates", "cms")
		case "marketing":
			coreRoot := common.MustGetEnv("WISPY_CORE_ROOT")
			basePath = filepath.Join(coreRoot, "data", "templates", "marketing")
		default:
			return "", fmt.Errorf("unknown template namespace: @%s", namespace)
		}

		return filepath.Join(basePath, filename), nil
	}

	// For non-namespaced paths, default to current site's templates (would need site info)
	// For now, just return error for non-namespaced templates
	return "", fmt.Errorf("non-namespaced template paths not supported yet: %s", templateName)
}

// isRemoteURL checks if a path is a remote URL
func isRemoteURL(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "//")
}

// validateAssetPath validates and sanitizes an asset path
func validateAssetPath(assetPath string, isInline bool) (string, error) {
	if assetPath == "" {
		return "", fmt.Errorf("asset path cannot be empty")
	}
	if strings.Contains(assetPath, "..") {
		return "", fmt.Errorf("asset path cannot contain '..'")
	}
	return strings.TrimSpace(assetPath), nil
}

// resolveAssetPath resolves asset paths like @app/style.css, @cms/foo.js, or site-relative assets
// Uses TemplateCtx.Instance (SiteInstance) for site context, matching resolveTemplatePath logic
func resolveAssetPath(assetPath string, ctx TemplateCtx) (string, error) {
	if isRemoteURL(assetPath) {
		return assetPath, nil
	}

	// Handle @namespace/asset.ext format
	if strings.HasPrefix(assetPath, "@") {
		parts := strings.SplitN(assetPath[1:], "/", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid asset path format: %s", assetPath)
		}
		namespace := parts[0]
		filename := parts[1]
		var basePath string
		switch namespace {
		case "app":
			coreRoot := common.MustGetEnv("WISPY_CORE_ROOT")
			basePath = filepath.Join(coreRoot, "data", "templates", "app")
		case "cms":
			coreRoot := common.MustGetEnv("WISPY_CORE_ROOT")
			basePath = filepath.Join(coreRoot, "data", "templates", "cms")
		case "marketing":
			coreRoot := common.MustGetEnv("WISPY_CORE_ROOT")
			basePath = filepath.Join(coreRoot, "data", "templates", "marketing")
		case "assets":
			// Assets are typically site-specific, so use the instance's base path
			if ctx.Instance == nil {
				return "", fmt.Errorf("no site instance available for resolving assets")
			}
			basePath = filepath.Join(ctx.Instance.BasePath, "assets")
		case "public":
			// Public assets are also site-specific, use the instance's base path
			if ctx.Instance == nil {
				return "", fmt.Errorf("no site instance available for resolving public assets")
			}
			basePath = filepath.Join(ctx.Instance.BasePath, "public")
		default:
			return "", fmt.Errorf("unknown asset namespace: @%s", namespace)
		}
		return filepath.Join(basePath, filename), nil
	}
	return "", fmt.Errorf("non-namespaced asset paths not supported yet: %s", assetPath)
}
