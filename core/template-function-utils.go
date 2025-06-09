package core

import (
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var policy *bluemonday.Policy

func init() {
	// Initialize the HTML sanitizer
	policy = bluemonday.UGCPolicy()
}

// SeekEndTag returns the index and length of the end tag after pos in raw, or -1, 0 if not found.
func SeekEndTag(raw, endTag string, pos int) (int, int) {
	idx := strings.Index(raw[pos:], endTag)
	if idx == -1 {
		return -1, 0
	}
	return pos + idx, len(endTag)
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
