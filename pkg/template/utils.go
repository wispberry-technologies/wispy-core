package template

import (
	"fmt"
	"log/slog"
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

// getContextKeys returns a slice of keys from a context map for debugging
func getContextKeys(data map[string]interface{}) []string {
	if data == nil {
		return []string{}
	}
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys
}

// getFilterNames returns a slice of filter names for debugging
func getFilterNames(filters models.FilterMap) []string {
	names := make([]string, 0, len(filters))
	for name := range filters {
		names = append(names, name)
	}
	return names
}

// debug prints a debug message with key-value pairs
func debug(msg string, kvs ...interface{}) {
	slog.Debug(msg, kvs...)
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

	debug("Processing filter chain", "chain", filterChainString)

	splitFilters := strings.Split(filterChainString, "|")
	debug("Split filters", "splitFilters", splitFilters)

	if len(splitFilters) == 1 {
		// No filters, just a single value
		value = resolveValue(strings.TrimSpace(splitFilters[0]), ctx)
		debug("No filters, resolved value", "value", value)
		// Don't treat nil values as errors - they should just render as empty
		if value == nil {
			return nil, nil, nil // No error, just nil value
		}
		return value, reflect.TypeOf(value), errors
	} else {
		// Multiple filters found, process them
		value = resolveValue(strings.TrimSpace(splitFilters[0]), ctx)
		debug("Initial value resolved", "value", value)
		if value == nil {
			// If initial value is nil, don't apply filters, just return nil
			return nil, nil, nil
		}
		valueType = reflect.TypeOf(value)
		for i, filterExpr := range splitFilters[1:] {
			filterExpr = strings.TrimSpace(filterExpr)
			filterName, args := ParseFilterExpression(filterExpr)
			debug("Processing filter", "index", i, "filterExpr", filterExpr, "filterName", filterName, "args", args)
			if filter, ok := filters[filterName]; ok {
				oldValue := value
				value = filter(value, valueType, args)
				valueType = reflect.TypeOf(value)
				debug("Filter applied", "filterName", filterName, "oldValue", oldValue, "newValue", value)
			} else {
				debug("Unknown filter", "filterName", filterName, "availableFilters", getFilterNames(filters))
				errors = append(errors, fmt.Errorf("unknown filter: %s", filterName))
			}
		}
	}
	return value, valueType, errors
}

// ParseFilterExpression parses a filter expression into name and arguments
func ParseFilterExpression(filterExpr string) (string, []string) {
	// Split the filter expression into name and arguments
	parts := strings.SplitN(filterExpr, ":", 2)
	if len(parts) == 1 {
		return parts[0], nil // No arguments
	}
	name := parts[0]
	args := common.FieldsRespectQuotes(parts[1])
	return name, args
}

// resolveValue resolves a value identifier to its actual value
// Supports literal strings (quoted) and context variables (with dot notation)
func resolveValue(valueIdentifier string, ctx TemplateCtx) interface{} {
	// Handle empty identifier
	if len(valueIdentifier) == 0 {
		debug("Empty identifier", "valueIdentifier", valueIdentifier)
		return nil
	}

	debug("Resolving value", "valueIdentifier", valueIdentifier)

	// Check for literal string values (surrounded by quotes)
	if common.IsQuotedString(valueIdentifier) {
		// Remove the surrounding quotes to get the literal value
		result := valueIdentifier[1 : len(valueIdentifier)-1]
		debug("Literal string value", "valueIdentifier", valueIdentifier, "result", result)
		return result
	}

	// Try to resolve from context data (handles dot notation for nested access)
	if ctx != nil && ctx.Data != nil {
		result := ResolveDotNotation(ctx.Data, valueIdentifier)
		debug("Context resolution", "valueIdentifier", valueIdentifier, "result", result, "contextKeys", getContextKeys(ctx.Data))
		return result
	}

	debug("No context or data available", "valueIdentifier", valueIdentifier)
	return nil
}

// ResolveDotNotation resolves dot notation (e.g., "user.name") in a context map
func ResolveDotNotation(ctx interface{}, key string) interface{} {
	if key == "" {
		return nil
	}

	// Remove leading dot if present (e.g., ".Site" becomes "Site")
	key = strings.TrimPrefix(key, ".")

	// Handle simple keys (no dots)
	if !strings.Contains(key, ".") {
		if m, ok := ctx.(map[string]interface{}); ok {
			return m[key]
		}
		return nil
	}

	// Handle dot notation (e.g., "user.profile.name")
	parts := strings.Split(key, ".")
	var current interface{} = ctx

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
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
func SeekEndTag(raw string, pos int, tagName string) (endTagPos int, errs []error) {
	startTagString := "{% " + tagName + " "
	endTagString := "{% end" + tagName + " %}"

	// Count nested tags to find the correct closing tag
	depth := 1
	searchPos := pos

	for depth > 0 && searchPos < len(raw) {
		// Find next occurrence of start tag or end tag
		nextStart := strings.Index(raw[searchPos:], startTagString)
		nextEnd := strings.Index(raw[searchPos:], endTagString)

		// Adjust positions to be absolute
		if nextStart != -1 {
			nextStart += searchPos
		}
		if nextEnd != -1 {
			nextEnd += searchPos
		}

		// If no end tag found, this is an error
		if nextEnd == -1 {
			errs = append(errs, fmt.Errorf("could not find end tag: %s", endTagString))
			return pos, errs
		}

		// If start tag comes before end tag (or no start tag), process end tag
		if nextStart == -1 || nextEnd < nextStart {
			depth--
			if depth == 0 {
				return nextEnd + len(endTagString), nil
			}
			searchPos = nextEnd + len(endTagString)
		} else {
			// Start tag comes first, increase depth
			depth++
			searchPos = nextStart + len(startTagString)
		}
	}

	return pos, nil
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
	}

	// For other types (slices, maps, etc.), check if they're empty
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return rv.Len() > 0
	case reflect.Ptr:
		return !rv.IsNil()
	}

	// All other types are truthy
	return true
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
