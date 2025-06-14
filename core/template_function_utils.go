package core

import (
	"fmt"
	"reflect"
	"strings"
	"wispy-core/common"
	"wispy-core/models"

	"github.com/microcosm-cc/bluemonday"
)

var policy *bluemonday.Policy

func init() {
	// Initialize the HTML sanitizer
	policy = bluemonday.UGCPolicy()
}

// ResolveFilterChain resolves a filter we assume the chain is a list of parts
// which have been split using common.FieldsRespectQuotes resulting in a slice of strings.
func ResolveFilterChain(filterChainString string, ctx TemplateCtx, filters models.FilterMap) (value interface{}, valueType reflect.Type, errors []error) {
	if len(filterChainString) == 0 {
		// Empty filter chain is not an error, just return nil
		return nil, nil, nil
	}

	// Trim whitespace from the entire string first
	filterChainString = strings.TrimSpace(filterChainString)

	splitFilters := strings.Split(filterChainString, "|")
	if len(splitFilters) == 1 {
		// No filters, just a single value
		value = resolveValue(strings.TrimSpace(splitFilters[0]), ctx)
		// Don't treat nil values as errors - they should just render as empty
		if value == nil {
			return nil, nil, nil // No error, just nil value
		}
		return value, reflect.TypeOf(value), errors
	} else {
		// Multiple filters found, process them
		value = resolveValue(strings.TrimSpace(splitFilters[0]), ctx)
		if value == nil {
			// If initial value is nil, don't apply filters, just return nil
			return nil, nil, nil
		}
		valueType = reflect.TypeOf(value)
		for _, filterExpr := range splitFilters[1:] {
			filterName, args := ParseFilterExpression(strings.TrimSpace(filterExpr))
			if filter, ok := filters[filterName]; ok {
				value = filter(value, valueType, args)
				valueType = reflect.TypeOf(value)
			} else {
				errors = append(errors, fmt.Errorf("unknown filter: %s", filterName))
			}
		}
	}
	return value, valueType, errors
}

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
		return nil
	}

	// Check for literal string values (surrounded by quotes)
	if common.IsQuotedString(valueIdentifier) {
		// Remove the surrounding quotes to get the literal value
		return valueIdentifier[1 : len(valueIdentifier)-1]
	}

	// Try to resolve from context data (handles dot notation for nested access)
	if ctx != nil && ctx.Data != nil {
		return ResolveDotNotation(ctx.Data, valueIdentifier)
	}

	return nil
}
