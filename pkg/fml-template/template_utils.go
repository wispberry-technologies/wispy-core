package template

import (
	"fmt"
	"strings"
	"wispy-core/pkg/common"
)

// --------------------
// Template Functions
// --------------------
// TODO
// e.ParseExpression(expression string) (any, []error)
// /----

// SeekClosingTag finds the position of the closing tag for block-level tags.
// It handles nested tags properly by maintaining a tag stack.
// Returns the position after the closing tag and any errors encountered.
func (e *Engine) SeekClosingTag(raw string, tagName string, pos int) (newPos int, content string, errs []error) {
	startDelim := e.StartDelim
	endDelim := e.EndDelim

	// Skip past the current opening tag
	contentStart := pos

	stack := 1 // Start with 1 for the opening tag we've already found
	searchPos := pos

	// Check for special handling for if/else
	isIfTag := tagName == "if"
	foundElse := false
	elseTagPos := -1

	for stack > 0 && searchPos < len(raw) {
		// Find next opening tag
		nextOpenTag := strings.Index(raw[searchPos:], startDelim)

		if nextOpenTag == -1 {
			// No more opening tags, so we can't have a proper closing tag
			return len(raw), "", []error{fmt.Errorf("unclosed tag: %s", tagName)}
		}

		// Calculate the absolute position
		nextOpenTagPos := searchPos + nextOpenTag

		// Move past the opening delimiter
		searchPos = nextOpenTagPos + len(startDelim)

		// Check if we have a closing delimiter
		nextClosingDelim := strings.Index(raw[searchPos:], endDelim)
		if nextClosingDelim == -1 {
			// No closing delimiter, syntax error
			return len(raw), "", []error{fmt.Errorf("unclosed delimiter at position %d", searchPos)}
		}

		// Extract the tag content
		tagStr := raw[searchPos : searchPos+nextClosingDelim]
		tagParts := strings.Fields(tagStr)

		if len(tagParts) == 0 {
			// Empty tag, continue search
			searchPos = searchPos + nextClosingDelim + len(endDelim)
			continue
		}

		// Handle special case for "else" tag in if blocks
		if isIfTag && tagParts[0] == "else" && stack == 1 && !foundElse {
			// Found the else tag for our current if block
			foundElse = true
			elseTagPos = nextOpenTagPos
			// Continue searching
			searchPos = searchPos + nextClosingDelim + len(endDelim)
			continue
		}

		// Check if it's an opening or closing tag
		if tagParts[0] == tagName {
			// Found another opening tag of same type, increase stack
			stack++
		} else if tagParts[0] == "end"+tagName || tagParts[0] == "end" {
			// Found a closing tag, decrease stack
			stack--
			if stack == 0 {
				// We've found the matching closing tag
				contentEndPos := nextOpenTagPos

				// For if/else, we may need to return only the content up to else
				if isIfTag && foundElse {
					return searchPos + nextClosingDelim + len(endDelim), raw[contentStart:elseTagPos], nil
				}

				return searchPos + nextClosingDelim + len(endDelim), raw[contentStart:contentEndPos], nil
			}
		}

		// Move past this tag
		searchPos = searchPos + nextClosingDelim + len(endDelim)
	}

	// If we get here, we couldn't find the closing tag
	return len(raw), "", []error{fmt.Errorf("unclosed tag: %s", tagName)}
}

/*
//	Engine CRUD Operations
*/

// Add & Get blocks for template content
func (e *Engine) AddBlock(name string, content any) {
	if e.Blocks == nil {
		e.Blocks = make(map[string]any)
	}
	e.Blocks[name] = content
}

func (e *Engine) GetBlock(name string) (any, bool) {
	if e.Blocks == nil {
		return nil, false
	}
	content, exists := e.Blocks[name]
	return content, exists
}

// Get & Set local context variables
func (e *Engine) SetVariable(name string, value any) {
	if e.Variables == nil {
		e.Variables = make(map[string]any)
	}
	e.Variables[name] = value
}

func (e *Engine) GetVariable(name string) (any, bool) {
	if e.Variables == nil {
		return nil, false
	}
	value, exists := e.Variables[name]
	return value, exists
}

// GetValue retrieves a value from the template context.
func (e *Engine) GetValue(key string, pathKeys ...string) (any, bool) {
	if key == "." {
		// If the key is just a dot, return the current context
		common.Debug("Returning local data context for key: %s", key)
		return e.GetLocalDataContext(), true
	}

	key = strings.TrimPrefix(key, ".") // Remove leading dot if present
	common.Debug("Resolving value for key: %s with path keys: %v", key, pathKeys)
	value, ok := e.ResolveValue(key, pathKeys...)
	return value, ok
}

// LocalDataContext allows setting a local context for the template rendering.
func (e *Engine) SetLocalDataContext(ctx map[string]any) {
	if e.LocalDataContext == nil {
		e.LocalDataContext = make(map[string]any)
	}
	for k, v := range ctx {
		e.LocalDataContext[k] = v
	}
}

// SetLocalDataContextValue sets a value in the local data context for the template.
func (e *Engine) SetLocalDataContextValue(key string, value any) {
	if e.LocalDataContext == nil {
		e.LocalDataContext = make(map[string]any)
	}
	e.LocalDataContext[key] = value
}

// GetLocalDataContextValue retrieves a value from the local data context for the template.
func (e *Engine) GetLocalDataContextValue(key string) (any, bool) {
	if e.LocalDataContext == nil {
		return nil, false
	}
	value, exists := e.LocalDataContext[key]
	return value, exists
}

// GetLocalDataContext retrieves the local data context for the template.
func (e *Engine) GetLocalDataContext() map[string]any {
	if e.LocalDataContext == nil {
		return make(map[string]any)
	}
	return e.LocalDataContext
}

// Getters
func (e *Engine) GetDataAdapter(name string) (*DataAdapter, bool) {
	if e.DataAdapters == nil {
		return nil, false
	}
	adapter, exists := e.DataAdapters[name]
	return adapter, exists
}
