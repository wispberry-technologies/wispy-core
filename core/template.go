package core

import (
	"fmt"
	"maps"
	"os"
	"reflect"
	"strings"

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

// NewTemplateContext creates a TemplateContext with proper defaults to avoid nil pointers.
func NewTemplateContext(data map[string]interface{}, engine *models.TemplateEngine) *models.TemplateContext {
	if data == nil {
		data = make(map[string]interface{})
	}
	// Always ensure InternalContext is a non-nil map[string]string
	var internal = &models.InternalContext{
		Flags:          make(map[string]interface{}),
		Blocks:         make(map[string]string),
		TemplatesCache: make(map[string]string),
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
func SeekEndTag(raw string, pos int, endTag string) (endTagPos int, errs []error) {
	endIdx := strings.Index(raw[pos:], endTag)
	if endIdx == -1 {
		errs = append(errs, fmt.Errorf("could not find end tag: %s", endTag))
		return pos, errs
	}
	return pos + endIdx + len(endTag), nil
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

// SetContextValue sets a value in a map[string]interface{} context.
func SetContextValue(ctx interface{}, key string, value interface{}) {
	if m, ok := ctx.(map[string]interface{}); ok {
		m[key] = value
	}
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

	return string(bytes)
}

func GetPage(domain, name string) string {
	pagePath := common.RootSitesPath(domain, "pages", name+".html")
	if _, err := os.Stat(pagePath); os.IsNotExist(err) {
		return "<!-- Error no page found: " + err.Error() + "-->"
	}

	bytes, err := os.ReadFile(pagePath)
	if err != nil {
		return "<!-- Error reading page file: " + err.Error() + "-->"
	}

	return string(bytes)
}

// Render renders the template with the given TemplateContext and returns the result string and any errors.
// This function processes template tags (e.g., {{ variable }}, {% if condition %}) and replaces them with
// their corresponding values or executes the associated logic.
//
// Template processing flow:
// 1. Scan for template tags using start/end delimiters
// 2. For each tag, determine if it's a block tag (if, for, etc.) or variable expression
// 3. Block tags are delegated to their specific render functions
// 4. Variable expressions are resolved from context data and filters are applied
// 5. Results are written to the output string builder
func Render(raw string, te *TemplateEngine, ctx TemplateCtx) (string, []error) {
	// Input validation - early return for edge cases
	if len(raw) == 0 {
		return "", nil
	}
	if te == nil {
		return raw, []error{fmt.Errorf("template engine cannot be nil")}
	}

	// Initialize output buffer and error collection
	var sb strings.Builder
	var errs []error

	// Pre-calculate delimiter lengths for performance (avoid repeated len() calls)
	startDelim := common.GetEnv("TEMPLATE_DELIMITER_OPEN", "{")
	endDelim := common.GetEnv("TEMPLATE_DELIMITER_CLOSE", "}")

	// Current position in the raw template string
	pos := 0

	// Main template processing loop - scan through the entire template string
	for pos < len(raw) {
		// Find the next template tag starting from current position
		delimOpenIndex := strings.Index(raw[pos:], startDelim)

		if delimOpenIndex == -1 {
			// No more template tags found - append remaining content and exit
			sb.WriteString(raw[pos:])
			break
		}

		// Calculate absolute position of tag start
		absoluteTagStart := pos + delimOpenIndex

		// Write all content before the tag to output buffer
		sb.WriteString(raw[pos:absoluteTagStart])

		switch {
		case strings.HasPrefix(raw[absoluteTagStart:], startDelim+"{"):
			pos = absoluteTagStart + len(startDelim+"{")
			// Handle variable expression tag (e.g., {{ variable }})
			// Find the closing delimiter for this variable tag
			tagEndPos := strings.Index(raw[pos:], endDelim+"}")
			if tagEndPos == -1 { // Malformed tag - no closing delimiter found
				errs = append(errs, fmt.Errorf("unclosed template variable starting at position %d", absoluteTagStart))
				pos += len("}" + endDelim) // Move past the start delimiter
				break
			}
			// Calculate absolute position of tag end
			absoluteTagEnd := absoluteTagStart + len(startDelim) + tagEndPos
			// Extract and clean tag contents (remove surrounding whitespace)
			tagContents := strings.TrimSpace(raw[pos:absoluteTagEnd])
			// Process the tag contents as a variable expression
			resolvedValue, resolvedValueType, filterErrs := ResolveFilterChain(tagContents, ctx, te.FilterMap)
			// Collect any errors from tag processing
			if len(filterErrs) > 0 {
				errs = append(errs, filterErrs...)
			}
			if resolvedValueType == nil {
				errs = append(errs, fmt.Errorf("could not resolve value for tag '%s'", tagContents))
				pos = absoluteTagEnd + len("}"+endDelim) // Move past the end delimiter
				break
			}
			switch resolvedValueType.Kind() {
			case reflect.Invalid:
				// If resolved value is nil, write an empty string
				sb.WriteString("")
			case reflect.String:
				// If resolved value is a string, sanitize and write it
				if strValue, ok := resolvedValue.(string); ok {
					sb.WriteString(policy.Sanitize(strValue))
				} else {
					errs = append(errs, fmt.Errorf("expected string value for tag '%s', got %T", tagContents, resolvedValue))
				}
			default:
				// For other types, convert to string and write it
				sb.WriteString(fmt.Sprintf("%v", resolvedValue))
			}
			// Update position for next iteration
			pos = absoluteTagEnd + len("}"+endDelim) // Move past the end delimiter
			//
		case strings.HasPrefix(raw[absoluteTagStart:], startDelim+"#"):
			pos = absoluteTagStart + len(startDelim+"#")
			// Handle comment tag (e.g., {# This is a comment #})
			// Find the closing delimiter for this comment tag
			tagEndPos := strings.Index(raw[pos+2:], "#"+endDelim)
			if tagEndPos == -1 { // Malformed comment tag - no closing delimiter found
				errs = append(errs, fmt.Errorf("unclosed template comment starting at position %d", absoluteTagStart))
				pos += len("#" + endDelim) // Move past the start delimiter
				break
			}
			// Calculate absolute position of tag end
			// Write the comment to output (or skip it entirely)
			// Here we just skip comments, but you could log them if needed
			// sb.WriteString(raw[absoluteTagStart:absoluteTagEnd+len(endDelim)]) // Uncomment to keep comments
			//
			// Update position for next iteration
			pos = pos + tagEndPos + len(endDelim)
		case strings.HasPrefix(raw[absoluteTagStart:], startDelim+"%"):
			pos = absoluteTagStart + len(startDelim+"%")
			// Handle block tag (e.g., {% if condition %})
			// Find the closing delimiter for this block tag
			tagEndPos := strings.Index(raw[pos:], "%"+endDelim)
			if tagEndPos == -1 {
				snipLen := 10
				if len(raw)-pos < snipLen {
					snipLen = len(raw) - pos
				}
				errs = append(errs, fmt.Errorf("unclosed template block starting at position %d \"%s\"", absoluteTagStart, raw[absoluteTagStart:snipLen]))
				pos += len("%" + endDelim) // Move past the start delimiter
				break
			}

			// Split arguments by spaces, respecting quotes
			args := common.FieldsRespectQuotes(raw[pos : pos+tagEndPos-len(endDelim)])
			if len(args) == 0 {
				errs = append(errs, fmt.Errorf("empty template tag starting at position %d", absoluteTagStart))
				pos = absoluteTagStart + tagEndPos + len(endDelim) // Move past the end delimiter
				break
			}
			tagName := strings.Trim(args[0], " \t\"'")
			// Call the tag function with the name and arguments
			if templateTag, ok := te.FuncMap[tagName]; ok {
				newPos, tagErrs := templateTag.Render(ctx, &sb, append([]string{tagName}, args...), raw[pos:absoluteTagStart+tagEndPos], absoluteTagStart)
				if len(tagErrs) > 0 {
					errs = append(errs, tagErrs...)
				}
				pos = newPos // Update position to the end of the tag
			} else {
				errs = append(errs, fmt.Errorf("unrecognized template tag '%s' starting! :(", raw[absoluteTagStart:pos+tagEndPos]))
				// Update position for next iteration
				pos = absoluteTagStart + tagEndPos + len(endDelim)
			}
		default:
			// Unrecognized tag format - write as-is and continue
			sb.WriteString(raw[absoluteTagStart:])
			errs = append(errs, fmt.Errorf("unrecognized template tag starting at position %d", absoluteTagStart))
			pos = absoluteTagStart + len(startDelim) // Move past the start delimiter
		}
	}

	return sb.String(), errs
}
