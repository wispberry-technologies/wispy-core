package core

import (
	"fmt"
	"maps"
	"os"
	"strings"
	"sync"

	"wispy-core/cache"
	"wispy-core/common"
	"wispy-core/models"
)

type FunctionMap = models.FunctionMap
type TemplateEngine = models.TemplateEngine
type TemplateTag = models.TemplateTag
type TemplateCtx = *models.TemplateContext

// -------------------
// Constructors
// -------------------

func NewSiteInstance(domain string) *models.SiteInstance {
	siteInstance := &models.SiteInstance{
		Domain:   domain,
		Name:     domain,
		BasePath: common.RootSitesPath(domain),
		IsActive: true,
		Theme:    "default",
		Config: models.SiteConfig{
			CssProcessor: "wispy-tail",
		},
		DBCache:        cache.NewDBCache(),
		SecurityConfig: &models.SiteSecurityConfig{},
		Templates:      make(map[string]string),
		Pages:          make(map[string]*models.Page), // routes for this site
		Mu:             sync.RWMutex{},                // mutex for thread-safe route access
		DBManager:      nil,                           // Will be initialized separately
	}
	return siteInstance
}

// NewTemplateContext creates a TemplateContext with proper defaults to avoid nil pointers.
func NewTemplateContext(data map[string]interface{}, engine *models.TemplateEngine) *models.TemplateContext {
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
	ctx := &models.TemplateContext{
		Data:            data,
		Engine:          engine,
		InternalContext: internal,
		Errors:          []error{},
	}
	if ctx.Engine == nil {
		ctx.Engine = NewTemplateEngine(nil) // Ensure we have a default engine if none provided
	}
	return ctx
}

// NewTemplateEngine creates a new TemplateEngine instance.
func NewTemplateEngine(funcMap FunctionMap) *TemplateEngine {
	var value = TemplateEngine{
		FuncMap:   funcMap,
		FilterMap: DefaultFilters(),
	}

	value.Render = func(raw string, ctx TemplateCtx) (string, []error) {
		// For simplicity just handle both types of tags in the same pass
		return Render(raw, &value, ctx)
	}

	value.CloneCtx = func(ctx TemplateCtx, newData map[string]interface{}) *models.TemplateContext {
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
			Errors:          make([]error, 0, len(ctx.Errors)),
		}
		// Copy existing errors to the new context
		newCtx.Errors = append(newCtx.Errors, ctx.Errors...)
		// Copy the request if it exists
		if ctx.Request != nil {
			newCtx.Request = ctx.Request
		}
		//
		return newCtx
	}

	return &value
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

// ResolveDotNotation resolves dot notation (e.g., "user.name") in a map[string]interface{} context.
func UnsafeResolveDotNotation(ctx interface{}, key string) interface{} {
	m, ok := ctx.(map[string]interface{})
	if !ok {
		return nil
	}
	parts := strings.Split(key, ".")
	var val interface{} = m
	for _, part := range parts {
		if mm, ok := val.(map[string]interface{}); ok {
			val = mm[part]
		} else {
			return nil
		}
	}
	return val
}

func ResolveDotNotation(ctx interface{}, key string) interface{} {
	if key == "" {
		return nil
	}
	var val interface{}
	if strings.Contains(key, ".") {
		val = UnsafeResolveDotNotation(ctx, key)
	} else if m, ok := ctx.(map[string]interface{}); ok {
		val = m[key]
	} else {
		return nil
	}
	if s, ok := val.(string); ok {
		return policy.Sanitize(s)
	}
	return val
}

// IsTruthy returns true if a value is considered 'true' for template logic.
func IsTruthy(val interface{}) bool {
	switch v := val.(type) {
	case bool:
		return v
	case string:
		return v != "" && v != "0" && v != "-1" && v != "false"
	case int, int64, float64:
		return v != 0
	case nil:
		return false
	default:
		return v != nil
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

	return RemoveMetadataFromContent(string(bytes))
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
	return RemoveMetadataFromContent(string(bytes))
}

// Render renders the template with the given TemplateContext and returns the result string and any errors.
// This function processes template tags (e.g., {{ variable }}, {% if condition %}) and replaces them with
// their corresponding values or executes the associated logic.
func Render(raw string, te *TemplateEngine, ctx TemplateCtx) (string, []error) {
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
