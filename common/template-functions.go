package common

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

// createTemplateFuncMap creates the function map for templates
func (re *RenderEngine) createTemplateFuncMap() template.FuncMap {
	return re.createTemplateFuncMapWithRequest(nil)
}

// createTemplateFuncMapWithRequest creates the function map for templates with request context
func (re *RenderEngine) createTemplateFuncMapWithRequest(r *http.Request) template.FuncMap {
	return template.FuncMap{
		// String functions
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": func(s string) string {
			// Simple title case implementation
			if s == "" {
				return s
			}
			return strings.ToUpper(string(s[0])) + strings.ToLower(s[1:])
		},
		"trim":      strings.TrimSpace,
		"replace":   strings.ReplaceAll,
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,

		// Utility functions
		"dict": func(values ...interface{}) map[string]interface{} {
			if len(values)%2 != 0 {
				return nil
			}
			dict := make(map[string]interface{})
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					continue
				}
				dict[key] = values[i+1]
			}
			return dict
		},

		"slice": func(items ...interface{}) []interface{} {
			return items
		},

		"add": func(a, b int) int {
			return a + b
		},

		"sub": func(a, b int) int {
			return a - b
		},

		"mul": func(a, b int) int {
			return a * b
		},

		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},

		"isTrue": func(conditions ...any) bool {
			// example: {{ if is-true .Condition1 .Condition2 .Condition3 }}
			for _, cond := range conditions {
				switch v := cond.(type) {
				case bool:
					if v {
						return true
					}
				case string:
					if strings.TrimSpace(v) != "" {
						return true
					}
				case int:
					if v != 0 {
						return true
					}
				case []interface{}:
					if len(v) > 0 {
						return true
					}
				default:
					return false
				}
			}
			return false
		},

		// Array/slice functions
		"first": func(slice []interface{}) interface{} {
			if len(slice) == 0 {
				return nil
			}
			return slice[0]
		},

		"last": func(slice []interface{}) interface{} {
			if len(slice) == 0 {
				return nil
			}
			return slice[len(slice)-1]
		},

		"len": func(v interface{}) int {
			switch val := v.(type) {
			case []interface{}:
				return len(val)
			case []string:
				return len(val)
			case string:
				return len(val)
			case map[string]interface{}:
				return len(val)
			default:
				return 0
			}
		},

		// Default value function
		"default": func(defaultValue, value interface{}) interface{} {
			if value == nil || value == "" {
				return defaultValue
			}
			return value
		},

		"import": func(importType string, site *Site, relPath string) template.HTML {
			// Use secure file reader for the site
			content, err := SecureReadFile(relPath)
			if err != nil {
				return template.HTML(fmt.Sprintf("<!-- import-inline error: %s -->", err.Error()))
			}

			if importType == "css" {
				return template.HTML(fmt.Sprintf("<style>%s</style>", template.HTMLEscapeString(string(content))))
			}
			if importType == "js" {
				return template.HTML(fmt.Sprintf("<script>%s</script>", template.HTMLEscapeString(string(content))))
			}
			if importType == "html" {
				return template.HTML(template.HTMLEscapeString(string(content)))
			}
			return template.HTML("<!-- import-inline error: unsupported type -->")
		},

		// API calling function for server-side rendering
		"api": func(method, path string, options ...map[string]interface{}) interface{} {
			// Validate inputs
			if method == "" {
				return map[string]interface{}{
					"error":  "HTTP method cannot be empty",
					"_debug": "Please provide a valid HTTP method (GET, POST, PUT, DELETE, etc.)",
				}
			}

			if path == "" {
				return map[string]interface{}{
					"error":  "API path cannot be empty",
					"_debug": "Please provide a valid API endpoint path",
				}
			}

			// Check if API dispatcher is available
			if re.apiDispatcher == nil {
				return map[string]interface{}{
					"error":  "API dispatcher not available",
					"_debug": "Internal API routing is not properly configured",
				}
			}

			// Parse options with validation
			var headers map[string]string
			var body []byte
			var debugInfo []string

			if len(options) > 0 {
				opts := options[0]

				// Extract headers with type validation
				if h, ok := opts["headers"].(map[string]interface{}); ok {
					headers = make(map[string]string)
					for k, v := range h {
						if s, ok := v.(string); ok {
							headers[k] = s
						} else {
							debugInfo = append(debugInfo, fmt.Sprintf("Header '%s' value must be string, got %T", k, v))
						}
					}
				} else if opts["headers"] != nil {
					debugInfo = append(debugInfo, fmt.Sprintf("Headers must be a map, got %T", opts["headers"]))
				}

				// Extract body with type validation and JSON marshaling
				if b, ok := opts["body"]; ok {
					switch v := b.(type) {
					case string:
						body = []byte(v)
					case []byte:
						body = v
					case map[string]interface{}:
						// Convert to JSON
						if jsonBytes, err := json.Marshal(v); err == nil {
							body = jsonBytes
							if headers == nil {
								headers = make(map[string]string)
							}
							// Only set Content-Type if not already specified
							if headers["Content-Type"] == "" {
								headers["Content-Type"] = "application/json"
							}
						} else {
							return map[string]interface{}{
								"error":  fmt.Sprintf("Failed to marshal body to JSON: %v", err),
								"_debug": "Body object could not be converted to JSON",
							}
						}
					case nil:
						// Explicitly set to nil - this is valid for GET requests
						body = nil
					default:
						return map[string]interface{}{
							"error":  fmt.Sprintf("Invalid body type: %T", v),
							"_debug": "Body must be string, []byte, map[string]interface{}, or nil",
						}
					}
				}

				// Validate other options
				for key := range opts {
					if key != "headers" && key != "body" {
						debugInfo = append(debugInfo, fmt.Sprintf("Unknown option '%s' - supported options: headers, body", key))
					}
				}
			}

			// Make the internal API call using the request context from the parameter
			response, err := re.apiDispatcher.Call(method, path, body, headers, r)
			if err != nil {
				result := map[string]interface{}{
					"error":  fmt.Sprintf("API call failed: %v", err),
					"_debug": fmt.Sprintf("Method: %s, Path: %s", method, path),
				}
				if len(debugInfo) > 0 {
					result["_validation_warnings"] = debugInfo
				}
				return result
			}

			// Add debug information to successful responses
			if response.Data != nil {
				if len(debugInfo) > 0 {
					response.Data["_validation_warnings"] = debugInfo
				}
				return response.Data
			}

			// Return full response for non-JSON responses
			result := map[string]interface{}{
				"status_code": response.StatusCode,
				"headers":     response.Headers,
				"body":        response.Body,
			}
			if len(debugInfo) > 0 {
				result["_validation_warnings"] = debugInfo
			}
			return result
		},

		// Helper functions for API responses
		"isAPIError": func(response interface{}) bool {
			if m, ok := response.(map[string]interface{}); ok {
				_, hasError := m["error"]
				return hasError
			}
			return false
		},

		"getAPIError": func(response interface{}) string {
			if m, ok := response.(map[string]interface{}); ok {
				if err, ok := m["error"].(string); ok {
					return err
				}
			}
			return ""
		},

		"getAPIData": func(response interface{}, key string) interface{} {
			if m, ok := response.(map[string]interface{}); ok {
				return m[key]
			}
			return nil
		},

		"isAPISuccess": func(response interface{}) bool {
			if m, ok := response.(map[string]interface{}); ok {
				// Check if there's no error
				if _, hasError := m["error"]; hasError {
					return false
				}
				// Check status code if available
				if statusCode, ok := m["status_code"].(int); ok {
					return statusCode >= 200 && statusCode < 300
				}
				// If no status code, assume success if no error
				return true
			}
			return false
		},

		// JSON helper functions
		"toJSON": func(v interface{}) string {
			if jsonBytes, err := json.Marshal(v); err == nil {
				return string(jsonBytes)
			}
			return "{}"
		},

		"fromJSON": func(jsonStr string) interface{} {
			var result interface{}
			if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
				return result
			}
			return nil
		},
	}
}
