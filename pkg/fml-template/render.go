package template

import (
	"fmt"
	"strings"
)

// Render renders the template with the given TemplateContext and returns the result string and any errors.
func (e *Engine) Render(raw string, localCtx map[string]any) (result string, errs []error) {
	if len(raw) == 0 {
		return result, errs
	}

	// Create a new local context for this render
	newCtx := make(map[string]any)

	// First add global context
	for k, v := range e.GlobalDataContext {
		newCtx[k] = v
	}

	// Then add/override with local context
	for k, v := range localCtx {
		newCtx[k] = v
	}

	// Set the local context for this render
	e.LocalDataContext = newCtx

	// Pre-allocate builder with estimated capacity for efficiency
	sb := &strings.Builder{}
	sb.Grow(len(raw) + len(raw)/5) // Estimate 20% expansion for typical templates

	pos := 0
	rawLen := len(raw)

	for pos < rawLen {
		// Find the next template tag or variable
		nextTag := strings.Index(raw[pos:], e.StartDelim)
		// If no more tags found, append the rest and break
		if nextTag == -1 {
			sb.WriteString(raw[pos:])
			break
		} else {
			// Append content before the tag
			sb.WriteString(raw[pos : pos+nextTag])
		}
		pos = e.iterate(raw, sb, pos+nextTag, &errs)
	}

	result = sb.String()
	return result, errs
}

// iterate expects to be when a template tag is found.
func (e *Engine) iterate(raw string, sb *strings.Builder, pos int, errs *[]error) int {
	// Process template tag {{ ... }}
	endPos := strings.Index(raw[pos:], e.EndDelim)
	if endPos == -1 {
		// If no closing delimiter found, append the rest and break
		*errs = append(*errs, fmt.Errorf("unclosed template tag starting at position %d", pos))
		sb.WriteString(raw[pos:])
		// No closing delimiter found, return the end of the string
		return len(raw)
	} else {
		endPos += pos // Adjust endPos to be relative to the start of the string
	}

	// Process the tag content
	newPos, tagErrs := e.ResolveRawTag(raw, sb, pos, endPos)

	// Error handling for template
	if len(tagErrs) > 0 {
		*errs = append(*errs, tagErrs...)
	}

	return newPos
}
